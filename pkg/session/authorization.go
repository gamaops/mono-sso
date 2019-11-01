package session

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gamaops/mono-sso/pkg/cache"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
)

type AuthorizationOptions struct {
	GrantRequestDuration       time.Duration
	AuthorizationCodeDuration  time.Duration
	AccessTokenDuration        time.Duration
	RefreshTokenDuration       time.Duration
	Issuer                     string
	AuthorizationServiceClient sso.AuthorizationServiceClient
}

type AuthorizationModel struct {
	Options *AuthorizationOptions
}

func IsOAuth2ValidResponseType(responseType string) bool {
	if responseType == "code" || responseType == "token" {
		return true
	}
	return false
}

func GenerateTokenCacheID(chc *cache.Cache, query *OAuth2URLQuery) (*strings.Builder, error) {

	hash := sha1.New()
	hash.Write(query.ScopesBytes)
	scopesHash := hex.EncodeToString(hash.Sum(nil))

	return chc.CreateID(":tkn:", query.ClientID, ':', query.SubjectID, ':', scopesHash)

}

func NewOAuth2URLQuery(r *http.Request) *OAuth2URLQuery {
	urlQuery := r.URL.Query()
	scopes := urlQuery.Get("scopes")
	return &OAuth2URLQuery{
		ClientID:     urlQuery.Get("client_id"),
		ResponseType: urlQuery.Get("response_type"),
		RedirectURI:  urlQuery.Get("redirect_uri"),
		State:        urlQuery.Get("state"),
		Scopes:       scopes,
		ScopesBytes:  []byte(scopes),
		ScopesArray:  strings.Fields(scopes),
	}
}

func NewOAuth2DataForm(r *http.Request) *OAuth2DataForm {
	r.ParseForm()
	values := r.Form
	return &OAuth2DataForm{
		ClientID:     values.Get("client_id"),
		GrantType:    values.Get("grant_type"),
		RedirectURI:  values.Get("redirect_uri"),
		ClientSecret: values.Get("client_secret"),
		RefreshToken: values.Get("refresh_token"),
		Code:         values.Get("code"),
	}
}

func (q *OAuth2URLQuery) IsValidAuthorizationRequest() bool {
	if len(q.SessionID) == 0 || len(q.SubjectID) != 24 || len(q.ClientID) != 24 || len(q.RedirectURI) <= 8 || !IsOAuth2ValidResponseType(q.ResponseType) {
		return false
	}
	return true
}
