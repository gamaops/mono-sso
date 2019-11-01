package session

import (
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
)

type AuthenticationRequest struct {
	Identifier        string `json:"identifier"`
	Password          string `json:"password"`
	RememberMe        bool   `json:"remember_me"`
	RecaptchaResponse string `json:"recaptcha_response"`
}

type AuthenticationResponse struct {
	Challenge        string               `json:"challenge"`
	ActivationMethod sso.ActivationMethod `json:"activation_method"`
	Subject          string               `json:"subject"`
	Expiration       int64                `json:"expiration"`
	Name             string               `json:"name"`
}

type ActivateSessionRequest struct {
	ActivationCode string `json:"activation_code"`
	Subject        string `json:"subject"`
	Challenge      string `json:"challenge"`
}

type ActivateSessionResponse struct {
	Expiration int64 `json:"expiration"`
}

type GrantScopesRequest struct {
	Nonce   string `json:"nonce"`
	Subject string `json:"subject"`
	Granted bool   `json:"granted"`
}

type ExchangeResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    uint32 `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type OAuth2URLQuery struct {
	ClientID     string
	ResponseType string
	RedirectURI  string
	State        string
	Scopes       string
	ScopesBytes  []byte
	ScopesArray  []string
	SubjectID    string
	SessionID    string
}

type OAuth2DataForm struct {
	ClientID     string
	ClientSecret string
	GrantType    string
	Code         string
	RedirectURI  string
	RefreshToken string
}
