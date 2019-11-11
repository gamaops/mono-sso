package constants

const (
	InvalidAccountSlg        string = "INVALID_ACCOUNT"
	InvalidActivationSlg     string = "INVALID_ACTIVATION"
	UknownScopesSlg          string = "UNKNOWN_SCOPES"
	UnauthorizedScopesSlg    string = "UNAUTHORIZED_SCOPES"
	InvalidClientSlg         string = "INVALID_CLIENT"
	InvalidResponseTypeSlg   string = "INVALID_RESPONSE_TYPE"
	NoValidRefreshTokenSlg   string = "NO_VALID_REFRESH_TOKEN"
	InternalErrorSlg         string = "INTERNAL_ERROR"
	InvalidRequestSlg        string = "INVALID_REQUEST"
	InvalidGrantSlg          string = "INVALID_GRANT"
	NotFoundSlg              string = "NOT_FOUND"
	VersionMismatchSlg       string = "VERSION_MISMATCH"
	InvalidAudienceSlg       string = "INVALID_AUDIENCE"
	InvalidRequestSessionSlg string = "INVALID_REQUEST_SESSION"
	InvalidTimestampSlg      string = "INVALID_TIMESTAMP"
	RequiredScopesSlg        string = "REQUIRED_SCOPES"
	InvalidTokenSlg          string = "INVALID_TOKEN"
	InvalidTenantSlg         string = "INVALID_TENANT"
)

const (
	InvalidAccountMsg      string = "identifier and/or password invalid"
	InvalidActivationMsg   string = "invalid activation request"
	UknownScopesMsg        string = "there're unknown scopes in authorization request"
	UnauthorizedScopesMsg  string = "there're unauthorized scopes in request"
	InvalidClientMsg       string = "invalid client id or redirect uri"
	InvalidResponseTypeMsg string = "this response type is unsupported by this type of client"
	NoValidRefreshTokenMsg string = "this client has no valid refresh token to this subject"
	InternalErrorMsg       string = "internal server error"
	InvalidGrantMsg        string = "invalid grant request"
	InvalidTenantMsg       string = "the tenant for this request is invalid"
)

var InvalidRecaptchaResponse []byte = []byte(`{"status":{"errors":[{"slug":"INVALID_RECAPTCHA","message":"invalid recaptcha response"}]}}`)
var InternalErrorResponse []byte = []byte(`{"status":{"errors":[{"slug":"INTERNAL_ERROR","message":"internal server error"}]}}`)
var InvalidPayloadResponse []byte = []byte(`{"status":{"errors":[{"slug":"INVALID_PAYLOAD","message":"invalid payload, check the data sent in request"}]}}`)
var UnauthorizedResponse []byte = []byte(`{"status":{"errors":[{"slug":"UNAUTHORIZED","message":"you must authenticate first"}]}}`)
var UnauthorizedExchangeResponse []byte = []byte(`{"status":{"errors":[{"slug":"UNAUTHORIZED_CODE","message":"your code is invalid, request the resource owner again for grant"}]}}`)
var InvalidExchangeResponse []byte = []byte(`{"status":{"errors":[{"slug":"INVALID_EXCHANGE","message":"your exchange request is invalid, something is wrong with your request parameters"}]}}`)
var InvalidActivationResponse []byte = []byte(`{"status":{"errors":[{"slug":"INVALID_ACTIVATION","message":"invalid activation request"}]}}`)
var InvalidRefreshResponse []byte = []byte(`{"status":{"errors":[{"slug":"INVALID_REFRESH","message":"invalid refresh token request"}]}}`)
var OkResponse []byte = []byte(`{"status":{"errors":[]}}`)
