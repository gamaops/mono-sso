package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gamaops/mono-sso/pkg/cache"
	"github.com/gamaops/mono-sso/pkg/constants"
	httpserver "github.com/gamaops/mono-sso/pkg/http-server"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"github.com/gamaops/mono-sso/pkg/oauth2"
	"github.com/gamaops/mono-sso/pkg/session"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"github.com/square/go-jose/v3/jwt"
)

func ExchangeHandler(
	httpServer *httpserver.HTTPServer,
	authzModel *session.AuthorizationModel,
	oauth2jose *oauth2.OAuth2Jose,
	chc *cache.Cache,
	w http.ResponseWriter,
	r *http.Request,
) {
	r.ParseForm()
	w.Header().Set("Content-Type", "application/json")
	dataForm := session.NewOAuth2DataForm(r)

	if len(dataForm.ClientID) != 24 || dataForm.GrantType != "authorization_code" || len(dataForm.Code) != 40 || len(dataForm.RedirectURI) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(constants.InvalidExchangeResponse)
		return
	}

	authCodeID, err := chc.CreateID(":authc:", dataForm.ClientID, ':', dataForm.Code)
	if err != nil {
		httpServer.Logger.Errorf("Error when generating authorization code cache ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	authCodeIDStr := authCodeID.String()

	pipe := chc.Client.TxPipeline()

	getCode := pipe.Get(authCodeIDStr)
	pipe.Del(authCodeIDStr)

	pipe.Exec()

	err = getCode.Err()

	if err != nil {
		if err != redis.Nil {
			httpServer.Logger.Errorf("Error while getting authorization code from Redis: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(constants.InternalErrorResponse)
			return
		}
		w.WriteHeader(http.StatusForbidden)
		w.Write(constants.UnauthorizedExchangeResponse)
		return
	}

	refreshTokenReqBytes, err := getCode.Bytes()

	if err != nil {
		httpServer.Logger.Errorf("Error while getting refresh token bytes to exchange authorization code: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	refreshTokenReq := &sso.NewRefreshTokenRequest{}

	proto.Unmarshal(refreshTokenReqBytes, refreshTokenReq)

	if refreshTokenReq.RedirectUri != dataForm.RedirectURI {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(constants.InvalidExchangeResponse)
		return
	}

	refreshTokenReq.ClientId = dataForm.ClientID
	refreshTokenReq.ClientSecret = dataForm.ClientSecret
	refreshTokenReq.AuthorizationCode = dataForm.Code
	refreshTokenReq.Duration = authzModel.Options.RefreshTokenDuration.String()

	ctx, cancel := context.WithTimeout(context.Background(), httpServer.Options.RequestDeadline)
	defer cancel()
	refreshTokenRes, err := authzModel.Options.AuthorizationServiceClient.NewRefreshToken(ctx, refreshTokenReq)

	if err != nil {
		httpServer.Logger.Errorf("Error while requesting new refresh token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	if refreshTokenRes.Status != nil && len(refreshTokenRes.Status.Errors) > 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(refreshTokenRes)
		return
	}

	nowNumericDate := jwt.NewNumericDate(time.Now())

	claims := &oauth2.TokenClaims{
		Claims: &jwt.Claims{
			Issuer:    authzModel.Options.Issuer,
			Subject:   refreshTokenReq.Subject,
			Audience:  []string{refreshTokenReq.ClientId},
			ID:        refreshTokenRes.RefreshTokenId,
			NotBefore: nowNumericDate,
			Expiry:    jwt.NewNumericDate(time.Unix(refreshTokenRes.ExpiresAt, 0)),
			IssuedAt:  nowNumericDate,
		},
		Scope: strings.Join(refreshTokenReq.Scopes, " "),
	}

	res := &session.ExchangeResponse{
		TokenType: "refresh_token",
		ExpiresIn: uint32(authzModel.Options.AccessTokenDuration.Seconds()),
	}

	tokenJWT := oauth2.GenerateJWT(oauth2jose, claims)

	tokenJWTStr, err := tokenJWT.CompactSerialize()

	if err != nil {
		httpServer.Logger.Errorf("Error when generating refresh token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	res.RefreshToken = tokenJWTStr

	claims.Expiry = jwt.NewNumericDate(time.Now().Add(authzModel.Options.AccessTokenDuration))

	claims.ID = ""

	tokenJWT = oauth2.GenerateJWT(oauth2jose, claims)

	tokenJWTStr, err = tokenJWT.CompactSerialize()
	if err != nil {
		httpServer.Logger.Errorf("Error when generating access token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	res.AccessToken = tokenJWTStr

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)

}

func RefreshTokenHandler(
	httpServer *httpserver.HTTPServer,
	authzModel *session.AuthorizationModel,
	oauth2jose *oauth2.OAuth2Jose,
	w http.ResponseWriter,
	r *http.Request,
) {
	r.ParseForm()
	w.Header().Set("Content-Type", "application/json")
	dataForm := session.NewOAuth2DataForm(r)

	if len(dataForm.ClientID) != 24 || dataForm.GrantType != "refresh_token" || len(dataForm.RefreshToken) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(constants.InvalidRefreshResponse)
		return
	}

	claims, err := oauth2.DecodeJWT(oauth2jose, dataForm.RefreshToken)

	if err != nil || !claims.Claims.Audience.Contains(dataForm.ClientID) || claims.Claims.Issuer != authzModel.Options.Issuer {
		httpServer.Logger.Warnf("Invalid refresh token to generate access token: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(constants.InvalidRefreshResponse)
		return
	}

	refreshTokenReq := &sso.NewRefreshTokenRequest{
		Subject:      claims.Claims.Subject,
		ForceNew:     false,
		ClientSecret: dataForm.ClientSecret,
		ClientId:     dataForm.ClientID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpServer.Options.RequestDeadline)
	defer cancel()

	refreshTokenRes, err := authzModel.Options.AuthorizationServiceClient.NewRefreshToken(ctx, refreshTokenReq)

	if err != nil {
		httpServer.Logger.Errorf("Invalid request to refresh token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(constants.InternalErrorResponse)
		return
	}

	if refreshTokenRes.Status != nil && len(refreshTokenRes.Status.Errors) > 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(refreshTokenRes)
		return
	}

	if refreshTokenRes.RefreshTokenId != claims.ID {
		w.WriteHeader(http.StatusForbidden)
		w.Write(constants.InvalidRefreshResponse)
		return
	}

	nowNumericDate := jwt.NewNumericDate(time.Now())

	claims.Claims.ID = ""
	claims.Claims.NotBefore = nowNumericDate
	claims.Claims.IssuedAt = nowNumericDate
	claims.Claims.Expiry = jwt.NewNumericDate(time.Now().Add(authzModel.Options.AccessTokenDuration))

	res := &session.ExchangeResponse{
		TokenType: "access_token",
		ExpiresIn: uint32(authzModel.Options.AccessTokenDuration.Seconds()),
	}

	tokenJWT := oauth2.GenerateJWT(oauth2jose, claims)

	tokenJWTStr, err := tokenJWT.CompactSerialize()

	res.AccessToken = tokenJWTStr

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)

}
