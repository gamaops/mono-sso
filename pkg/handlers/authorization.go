package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gamaops/mono-sso/pkg/cache"
	"github.com/gamaops/mono-sso/pkg/constants"
	httpserver "github.com/gamaops/mono-sso/pkg/http-server"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"github.com/gamaops/mono-sso/pkg/oauth2"
	"github.com/gamaops/mono-sso/pkg/session"
	"github.com/golang/protobuf/proto"
	"github.com/square/go-jose/v3/jwt"
)

func AuthorizationHandler(
	httpServer *httpserver.HTTPServer,
	authzModel *session.AuthorizationModel,
	authnModel *session.AuthenticationModel,
	oauth2jose *oauth2.OAuth2Jose,
	chc *cache.Cache,
	w http.ResponseWriter,
	r *http.Request,
	replacers *SignInTemplateReplacers,
) bool {

	sessionCookie, subjectCookie, err := authnModel.GetSessionAndSubjectCookies(r)

	if err != nil {
		httpServer.Logger.Fatalf("Error while getting session/subject cookie for authorization: %v", err)
	}

	query := session.NewOAuth2URLQuery(r)

	if subjectCookie != nil {
		query.SubjectID = subjectCookie.Value
	}

	if sessionCookie != nil {
		query.SessionID = sessionCookie.Value
	}

	if !query.IsValidAuthorizationRequest() {
		return true
	}

	redirectURL, err := url.Parse(query.RedirectURI)
	if err != nil {
		httpServer.Logger.Warnf("Invalid client redirect uri: %v", err)
		return true
	}

	redisSessID, err := chc.CreateID(":sess:", query.SessionID, ':', query.SubjectID)
	if err != nil {
		httpServer.Logger.Errorf("Error when generating session cache ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return false
	}
	sessSubID := redisSessID.String()

	var sessSub *sso.SessionSubject = nil
	sessSub, err = session.GetCachedSessionSubject(chc, sessSubID)
	if err != nil {
		httpServer.Logger.Errorf("Error when getting cached session subject: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return false
	}

	if sessSub == nil {
		replacers.RequireSignIn = true
		return true
	}

	hashParameters := &url.Values{}

	if len(query.State) > 0 {
		hashParameters.Add("state", query.State)
	}

	// TODO: Check max_age against sessSub.ActivatedAt

	var tokenCacheID *strings.Builder
	var existsCacheCount int64 = 0
	if query.ResponseType == "token" {
		tokenCacheID, err = session.GenerateTokenCacheID(chc, query)
		if err != nil {
			httpServer.Logger.Errorf("Error when generating token cache ID: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
			return false
		}
		existsCacheCount, err = chc.Client.Exists(tokenCacheID.String()).Result()
		if err != nil {
			httpServer.Logger.Fatalf("Error while getting access token cache from Redis: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpServer.Options.RequestDeadline)
	defer cancel()

	if existsCacheCount == 0 {

		// TODO: Send session, user-agent and IPs
		authCliReq := &sso.AuthorizeClientRequest{
			ClientId:     query.ClientID,
			Scopes:       query.ScopesArray,
			RedirectUri:  query.RedirectURI,
			Subject:      query.SubjectID,
			ResponseType: query.ResponseType,
			State:        query.State,
			TenantId:     authnModel.Options.TenantID,
		}

		authCliRes, err := authzModel.Options.AuthorizationServiceClient.AuthorizeClient(ctx, authCliReq)

		if err != nil {
			httpServer.Logger.Errorf("Error while sending AccountService.AuthorizeClient request to gRPC server: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
			return false
		}

		if authCliRes.Status != nil && len(authCliRes.Status.Errors) > 0 {
			statusError := authCliRes.Status.Errors[0]
			hashParameters.Add("error", "server_error")
			hashParameters.Add("details", statusError.Message)
			if statusError.Slug == constants.UknownScopesSlg {
				hashParameters.Set("error", "invalid_scope")
			} else if statusError.Slug == constants.InvalidClientSlg {
				hashParameters.Set("error", "unauthorized_client")
			} else if statusError.Slug == constants.InvalidResponseTypeSlg {
				hashParameters.Set("error", "unsupported_response_type")
			}
			redirectURL.Fragment = hashParameters.Encode()
			http.Redirect(w, r, redirectURL.String(), http.StatusFound)
			return false
		}

		if len(authCliRes.UnauthorizedScopes) > 0 {
			replacers.Scopes = query.Scopes
			replacers.ClientName = authCliRes.ClientName
			grantNonce, err := nonceGenerator.New()
			if err != nil {
				httpServer.Logger.Errorf("Error when generating grant nonce: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(constants.InternalErrorResponse)
				return false
			}
			replacers.GrantNonce = grantNonce.Base32()

			redisSessID.WriteString(":grt:")
			redisSessID.WriteString(replacers.GrantNonce)

			grantID := redisSessID.String()
			grantData, _ := proto.Marshal(authCliReq)

			err = chc.Client.Set(grantID, grantData, authzModel.Options.GrantRequestDuration).Err()
			if err != nil {
				httpServer.Logger.Errorf("Error when setting grant cache nonce: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(constants.InternalErrorResponse)
				return false
			}
			return true
		}

	}
	if query.ResponseType == "code" {

		salt := make([]byte, 24)
		_, err = rand.Read(salt)
		if err != nil {
			httpServer.Logger.Fatalf("Error generating random bytes for authorization code: %v", err)
		}
		hash := sha1.New()
		hash.Write(salt)
		userAgent := r.Header.Get("User-Agent")
		hash.Write([]byte(userAgent))
		authorizationCode := hex.EncodeToString(hash.Sum(nil))

		eventReq := &sso.RegisterEventRequest{
			Data: map[string]string{
				"user_agent":         userAgent,
				"authorization_code": authorizationCode,
				"session_id":         query.SessionID,
				"account_id":         query.SubjectID,
				"client_id":          query.ClientID,
				"scopes":             query.Scopes,
				"response_type":      query.ResponseType,
				"redirect_uri":       query.RedirectURI,
			},
			Level:       sso.EventLevel_WARNING,
			IsSensitive: true,
			Message:     fmt.Sprintf("generated new authorization code for client %v: %v", query.ClientID, authorizationCode),
		}

		hashParameters.Add("code", authorizationCode)

		authCodeID, err := chc.CreateID(":authc:", authnModel.Options.TenantID, ':', query.ClientID, ':', authorizationCode)
		if err != nil {
			httpServer.Logger.Errorf("Error when generating authorization code cache ID: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
			return false
		}

		refreshTokenReq := &sso.NewRefreshTokenRequest{
			SessionId:   query.SessionID,
			Subject:     query.SubjectID,
			Scopes:      query.ScopesArray,
			RedirectUri: query.RedirectURI,
			ForceNew:    false,
		}

		refreshTokenReqBytes, err := proto.Marshal(refreshTokenReq)

		if err != nil {
			httpServer.Logger.Errorf("Error while encoding refresh token request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
			return false
		}

		err = chc.Client.Set(authCodeID.String(), refreshTokenReqBytes, authzModel.Options.AuthorizationCodeDuration).Err()

		if err != nil {
			httpServer.Logger.Errorf("Setting authorization code on Redis caused an error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
			return false
		}

		_, err = authnModel.Options.AccountServiceClient.RegisterEvent(ctx, eventReq)
		if err != nil {
			httpServer.Logger.Errorf("Error while registering event for authorization code: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
			return false
		}

	} else if query.ResponseType == "token" {

		nowNumericDate := jwt.NewNumericDate(time.Now())

		claims := &oauth2.TokenClaims{
			Claims: &jwt.Claims{
				Issuer:    authzModel.Options.Issuer,
				Subject:   query.SubjectID,
				Audience:  []string{query.ClientID},
				NotBefore: nowNumericDate,
				Expiry:    jwt.NewNumericDate(time.Now().Add(authzModel.Options.AccessTokenDuration)),
				IssuedAt:  nowNumericDate,
			},
			Scope:  query.Scopes,
			Tenant: authnModel.Options.TenantID,
		}

		tokenJWT := oauth2.GenerateJWT(oauth2jose, claims)

		accessToken, err := tokenJWT.CompactSerialize()

		if err != nil {
			httpServer.Logger.Errorf("Error while generating access token (implicit flow): %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
			return false
		}

		hashParameters.Add("access_token", accessToken)
		hashParameters.Add("token_type", "access_token")

		if existsCacheCount == 0 {
			err = chc.Client.Set(tokenCacheID.String(), query.SessionID, authzModel.Options.AccessTokenDuration).Err()
			if err != nil {
				httpServer.Logger.Errorf("Error while creating access token cache: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(constants.InternalErrorResponse)
				return false
			}
		}

	}

	redirectURL.Fragment = hashParameters.Encode()
	http.Redirect(w, r, redirectURL.String(), http.StatusFound)

	return false
}

func GrantScopesHandler(
	httpServer *httpserver.HTTPServer,
	authzModel *session.AuthorizationModel,
	authnModel *session.AuthenticationModel,
	chc *cache.Cache,
	w http.ResponseWriter,
	r *http.Request,
) {
	grantReq := &session.GrantScopesRequest{}
	if !httpserver.ReadJSONRequestBody(grantReq, httpServer, w, r) {
		return
	}
	sessCookie, err := r.Cookie(authnModel.SessionCookieKey)
	var sessID string
	if err == nil {
		sessID = sessCookie.Value
	} else {
		httpServer.Logger.Warnf("Unauthorized request to grant scopes: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(constants.UnauthorizedResponse)
		return
	}

	redisSessID, err := chc.CreateID(":sess:", sessID, ':', grantReq.Subject, ":grt:", grantReq.Nonce)
	if err != nil {
		httpServer.Logger.Errorf("Error when generating session cache ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	grantID := redisSessID.String()

	pipe := chc.Client.TxPipeline()

	getGrant := pipe.Get(grantID)
	pipe.Del(grantID)

	pipe.Exec()

	authCliReqBytes, err := getGrant.Bytes()

	if err != nil || getGrant.Err() != nil {
		httpServer.Logger.Warnf("Unauthorized request to grant scopes (invalid nonce): %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(constants.UnauthorizedResponse)
		return
	}

	authCliReq := &sso.AuthorizeClientRequest{}

	err = proto.Unmarshal(authCliReqBytes, authCliReq)
	if err != nil {
		httpServer.Logger.Errorf("Error while decoding authorization request (grant request): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	// TODO: Send session, user-agent and IPs
	grantScopesReq := &sso.GrantScopesRequest{
		ClientId: authCliReq.ClientId,
		Subject:  grantReq.Subject,
		Scopes:   authCliReq.Scopes,
		TenantId: authnModel.Options.TenantID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpServer.Options.RequestDeadline)
	defer cancel()

	grantScopesRes, err := authzModel.Options.AuthorizationServiceClient.GrantScopes(ctx, grantScopesReq)

	if err != nil {
		httpServer.Logger.Errorf("Error while sending grant scopes request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	if grantScopesRes.Status != nil && len(grantScopesRes.Status.Errors) > 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(grantScopesRes)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(constants.OkResponse)

}
