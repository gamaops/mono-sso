package session

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gamaops/mono-sso/pkg/constants"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-common"
	"github.com/gamaops/mono-sso/pkg/oauth2"
	"github.com/square/go-jose/v3/jwt"
)

type SessionRequestValidator struct {
	TimestampPastToleration time.Duration
	RequiredScopes          [][]string
	AllowedTenants          []string
	Audience                string
	Jose                    *oauth2.OAuth2Jose
	requiredScopesErr       error
	allowedTenantsMap       map[string]bool
}

func (s *SessionRequestValidator) Load() {
	scopesGroup := make([]string, len(s.RequiredScopes))
	for i, scopes := range s.RequiredScopes {
		scopesGroup[i] = strings.Join(scopes, ", ")
	}
	s.requiredScopesErr = &ErrRequiredScopes{
		Scopes: strings.Join(scopesGroup, " or "),
	}
	s.allowedTenantsMap = make(map[string]bool, len(s.AllowedTenants))
	for _, tenant := range s.AllowedTenants {
		s.allowedTenantsMap[tenant] = true
	}
}

var ErrInvalidAudience = errors.New("invalid token audience")
var ErrInvalidTimestamp = errors.New("invalid timestamp")
var ErrInvalidRequestSession = errors.New("invalid request session")
var ErrInvalidTenant = errors.New("invalid request session tenant")

// TODO: Validate locales
type ErrRequiredScopes struct {
	Scopes string
}

func (e *ErrRequiredScopes) Error() string {
	return fmt.Sprintf("required scopes to complete request: %v", e.Scopes)
}

func (s *SessionRequestValidator) ValidateSession(req *sso.RequestSession) (*oauth2.TokenClaims, error) {

	if req == nil {
		return nil, ErrInvalidRequestSession
	}

	now := time.Now()
	if req.Timestamp > now.Unix() || (s.TimestampPastToleration < 0 && req.Timestamp < now.Add(s.TimestampPastToleration).Unix()) {
		return nil, ErrInvalidTimestamp
	}

	claims, err := oauth2.DecodeJWT(s.Jose, req.AccessToken)

	if err != nil {
		return nil, err
	}

	if !s.allowedTenantsMap[claims.Tenant] {
		return claims, ErrInvalidTenant
	}

	if !claims.Audience.Contains(s.Audience) {
		return claims, ErrInvalidAudience
	}

	scopes := strings.Fields(claims.Scope)
	scopesMap := make(map[string]bool, len(scopes))
	for _, scope := range scopes {
		scopesMap[scope] = true
	}

	var invalid bool
	for _, requiredScopes := range s.RequiredScopes {
		invalid = false
		for _, scope := range requiredScopes {
			if !scopesMap[scope] {
				invalid = true
				break
			}
		}
		if !invalid {
			break
		}
	}

	if invalid {
		return claims, s.requiredScopesErr
	}

	return claims, nil

}

func (s *SessionRequestValidator) ParseSessionErrorToStatus(err error, status *sso.ResponseStatus) *sso.ResponseStatus {

	if status == nil {
		status = &sso.ResponseStatus{}
	}

	slug := constants.InternalErrorMsg

	switch err {
	case jwt.ErrExpired,
		jwt.ErrInvalidAudience,
		jwt.ErrInvalidClaims,
		jwt.ErrInvalidID,
		jwt.ErrInvalidIssuer,
		jwt.ErrInvalidSubject,
		jwt.ErrInvalidContentType,
		jwt.ErrIssuedInTheFuture,
		jwt.ErrNotValidYet,
		jwt.ErrUnmarshalAudience,
		jwt.ErrUnmarshalNumericDate:
		slug = constants.InvalidTokenSlg
	case ErrInvalidAudience:
		slug = constants.InvalidAudienceSlg
	case ErrInvalidRequestSession:
		slug = constants.InvalidRequestSessionSlg
	case ErrInvalidTimestamp:
		slug = constants.InvalidTimestampSlg
	case ErrInvalidTenant:
		slug = constants.InvalidTenantSlg
	case s.requiredScopesErr:
		slug = constants.RequiredScopesSlg
	}

	status.Errors = append(status.Errors, &sso.ResponseStatus_Error{
		Slug:    slug,
		Message: err.Error(),
	})

	return status

}
