package handlers

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dpapathanasiou/go-recaptcha"
	"github.com/gamaops/mono-sso/pkg/cache"
	"github.com/gamaops/mono-sso/pkg/constants"
	httpserver "github.com/gamaops/mono-sso/pkg/http-server"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"github.com/gamaops/mono-sso/pkg/oauth2"
	"github.com/gamaops/mono-sso/pkg/session"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
)

func IndexHandler(
	httpServer *httpserver.HTTPServer,
	authzModel *session.AuthorizationModel,
	authnModel *session.AuthenticationModel,
	oauth2jose *oauth2.OAuth2Jose,
	chc *cache.Cache,
	w http.ResponseWriter,
	r *http.Request,
) {
	replacers := &SignInTemplateReplacers{
		RequireSignIn: false,
	}
	if !AuthorizationHandler(httpServer, authzModel, authnModel, oauth2jose, chc, w, r, replacers) {
		return
	}
	// Validate if is already logged in through cookie
	// Validate id_token_hint
	// On front end don't forget about login_hint
	// If MFA enabled requires MFA before authorization
	// Validate max_age to verify if the user has logged in a while
	err := authnModel.IndexTemplate.Execute(w, replacers)
	if err != nil {
		httpServer.Logger.Errorf("Error parsing sign in template: %v", err)
	}
}

func AuthenticateHandler(httpServer *httpserver.HTTPServer, model *session.AuthenticationModel, chc *cache.Cache, w http.ResponseWriter, r *http.Request) {
	authReq := &session.AuthenticationRequest{}
	if !httpserver.ReadJSONRequestBody(authReq, httpServer, w, r) {
		return
	}
	clientIPs, err := httpserver.ClientIPsFromRequest(r)
	if err != nil {
		model.Logger.Errorf("Error when getting client IPs: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}
	// Validate recaptcha
	isValid, err := recaptcha.Confirm(clientIPs.SourceIP, authReq.RecaptchaResponse)
	if !isValid || err != nil {
		model.Logger.Warnf("Invalid recaptcha: %v", err)
		w.WriteHeader(http.StatusForbidden)
		w.Write(constants.InvalidRecaptchaResponse)
		return
	}

	// Get the current session from cookie
	sessCookie, err := r.Cookie(model.SessionCookieKey)
	var sessID string
	if err == http.ErrNoCookie {
		sessIDSrc, err := sessionIDGenerator.New()
		if err != nil {
			model.Logger.Errorf("Error when generating session ID: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
		}
		sessID = sessIDSrc.Base32()
	} else if err != nil {
		model.Logger.Errorf("Error when getting session cookie: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	} else {
		sessID = sessCookie.Value
	}

	// Send request to sso-service AccountService.SignIn
	signInReq := &sso.SignInRequest{
		Identifier:             authReq.Identifier,
		Password:               authReq.Password,
		SessionId:              sessID,
		ForwardedIps:           clientIPs.ForwardedIPs,
		ClientIp:               clientIPs.ClientIP,
		SourceIp:               clientIPs.SourceIP,
		UserAgent:              r.Header.Get("User-Agent"),
		ActivationCodeDuration: model.Options.MFASessionDuration.String(),
		TenantId:               model.Options.TenantID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpServer.Options.RequestDeadline)
	defer cancel()

	signInRes, err := model.Options.AccountServiceClient.SignIn(ctx, signInReq)

	if err != nil {
		model.Logger.Errorf("Error while sending AccountService.SignIn request to gRPC server: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}
	if signInRes.Status != nil && len(signInRes.Status.Errors) > 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(signInRes)
		return
	}

	// Generate session cookie (store in Redis)
	redisSessID, err := chc.CreateID(":sess:", signInReq.SessionId, ':', signInRes.Subject)
	if err != nil {
		model.Logger.Errorf("Error when generating session cache ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}
	sessSubID := redisSessID.String()

	// Saves some data in session:
	// Last SignIn Time
	// Last Access Time
	// Current Activation Challenge (only if MFA is enabled)
	// "Remember Me" time

	sessSub, err := session.GetCachedSessionSubject(chc, sessSubID)
	if err != nil {
		model.Logger.Errorf("Error when getting cached session subject: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}
	if sessSub == nil {
		sessSub = &sso.SessionSubject{
			AuthenticatedAt: time.Now().Unix(),
			ExpiresAt:       0,
			ActivatedAt:     0,
		}
	}

	expiration, remember := session.GetExpirationFromAuthentication(model, authReq)

	sessCookie = &http.Cookie{
		Name:     model.SessionCookieKey,
		Value:    signInReq.SessionId,
		Domain:   model.Options.SessionCookieDomain,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     model.Options.SessionCookiePath,
	}
	subCookie := &http.Cookie{
		Name:     model.SubjectCookieKey,
		Value:    signInRes.Subject,
		Domain:   model.Options.SessionCookieDomain,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     model.Options.SessionCookiePath,
	}

	if remember {
		sessCookie.Expires = time.Now().Add(expiration)
		sessSub.ExpiresAt = sessCookie.Expires.Unix()
	}

	http.SetCookie(w, sessCookie)
	http.SetCookie(w, subCookie)

	authRes := &session.AuthenticationResponse{
		ActivationMethod: signInRes.ActivationMethod,
		Subject:          signInRes.Subject,
		Expiration:       sessSub.ExpiresAt,
		Name:             signInRes.Profile.Name,
	}

	sessExp := expiration

	// MFA disabled
	if signInRes.ActivationMethod == sso.ActivationMethod_NONE {
		sessSub.ActivatedAt = sessSub.AuthenticatedAt
	} else {

		redisSessID.WriteString(":clg")

		challenge, err := challengeGenerator.New()
		if err != nil {
			model.Logger.Errorf("Error when generating challenge: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
		}

		challengeStr := challenge.Base32()

		challengeHash := sha1.New()
		challengeHash.Write([]byte(signInReq.UserAgent))
		challengeHash.Write([]byte(challengeStr))

		err = chc.Client.Set(redisSessID.String(), hex.EncodeToString(challengeHash.Sum(nil)), model.Options.MFASessionDuration).Err()
		if err != nil {
			model.Logger.Errorf("Error when setting cache for MFA challenge string: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
			return
		}

		authRes.Challenge = challengeStr
		// Only resets the session cache if it's the first authentication
		if sessSub.ActivatedAt == 0 {
			sessExp = model.Options.MFASessionDuration
		}
	}

	err = session.UpdateSessionSubject(chc, sessSubID, sessSub, sessExp)
	if err != nil {
		model.Logger.Errorf("Error when updating session subject: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	resPld, _ := json.Marshal(authRes)
	w.Write(resPld)

}

func ActivateSessionHandler(httpServer *httpserver.HTTPServer, model *session.AuthenticationModel, chc *cache.Cache, w http.ResponseWriter, r *http.Request) {
	actSessReq := &session.ActivateSessionRequest{}
	if !httpserver.ReadJSONRequestBody(actSessReq, httpServer, w, r) {
		return
	}

	sessCookie, err := r.Cookie(model.SessionCookieKey)
	var sessID string
	if err == nil {
		sessID = sessCookie.Value
	} else {
		model.Logger.Warnf("Unauthorized request to activate session: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(constants.UnauthorizedResponse)
		return
	}

	redisSessID, err := chc.CreateID(":sess:", sessID, ':', actSessReq.Subject)
	if err != nil {
		model.Logger.Errorf("Error when generating session cache ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	sessSubID := redisSessID.String()

	redisSessID.WriteString(":clg")
	sessClgID := redisSessID.String()

	pipe := chc.Client.TxPipeline()

	getClg := pipe.Get(sessClgID)
	pipe.Del(sessClgID)

	pipe.Exec()

	userAgent := r.Header.Get("User-Agent")

	challengeHash := sha1.New()
	challengeHash.Write([]byte(userAgent))
	challengeHash.Write([]byte(actSessReq.Challenge))
	challengeShaSum := hex.EncodeToString(challengeHash.Sum(nil))

	if getClg.Val() != challengeShaSum {
		model.Logger.Warnf("Invalid activation request challenge: %v", err)
		w.WriteHeader(http.StatusForbidden)
		w.Write(constants.InvalidActivationResponse)
		return
	}

	currentSess := chc.Client.Get(sessSubID)
	currentSessProto, err := currentSess.Bytes()
	var sessSub *sso.SessionSubject
	if err == redis.Nil {
		model.Logger.Warnf("Unauthorized request to activate session (invalid session): %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(constants.UnauthorizedResponse)
		return
	} else if err != nil {
		model.Logger.Errorf("Error when getting session from Redis: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	} else {
		sessSub = new(sso.SessionSubject)
		proto.Unmarshal(currentSessProto, sessSub)
	}

	actReq := &sso.ActivateSessionRequest{
		SessionId:      sessID,
		ActivationCode: actSessReq.ActivationCode,
		Subject:        actSessReq.Subject,
		UserAgent:      userAgent,
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpServer.Options.RequestDeadline)
	defer cancel()

	actRes, err := model.Options.AccountServiceClient.ActivateSession(ctx, actReq)

	if err != nil {
		model.Logger.Errorf("Error while sending AccountService.ActivateSession request to gRPC server: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}
	if actRes.Status != nil && len(actRes.Status.Errors) > 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(actRes)
		return
	}

	sessSub.ActivatedAt = time.Now().Unix()
	sessExp := 0 * time.Second

	if sessSub.ExpiresAt > 0 {
		sessExp = time.Since(time.Unix(sessSub.ExpiresAt, 0))
	}

	session.UpdateSessionSubject(chc, sessSubID, sessSub, sessExp)

	actSessRes := &session.ActivateSessionResponse{
		Expiration: sessSub.ExpiresAt,
	}

	resPld, _ := json.Marshal(actSessRes)
	w.Write(resPld)
}
