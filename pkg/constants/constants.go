package constants

const (
	InvalidAccountSlg      string = "INVALID_ACCOUNT"
	InvalidActivationSlg   string = "INVALID_ACTIVATION"
	UknownScopesSlg        string = "UNKNOWN_SCOPES"
	InvalidClientSlg       string = "INVALID_CLIENT"
	InvalidResponseTypeSlg string = "INVALID_RESPONSE_TYPE"
	NoValidRefreshTokenSlg string = "NO_VALID_REFRESH_TOKEN"
)

const (
	InvalidAccountMsg      string = "identifier and/or password invalid"
	InvalidActivationMsg   string = "invalid activation request"
	UknownScopesMsg        string = "there're unknown scopes in authorization request"
	InvalidClientMsg       string = "invalid client id or redirect uri"
	InvalidResponseTypeMsg string = "this response type is unsupported by this type of client"
	NoValidRefreshTokenMsg string = "this client has no valid refresh token to this subject"
)
