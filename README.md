# Mono SSO

Mono SSO is an OAuth 2 compliant SSO focused on scalability and performance to provide authorization for SCSs and microservices.

----------------

## Architecture

The project is divided in three parts:

* SSO Provider
* SSO Service
* SSO Manager

Each part has its own responsibilities and can be scaled/deployed according to your needs.

----------------

## SSO Provider

SSO Provider is responsible for providing the edge API to end-users and interact with resource owners/client apps. It depends on:

* Redis - store cache and session data
* SSO Service - consumes the SSO API through gRPC

It provides a main entry point that parses a HTML file to display the SSO sign-in page and there are some vars that it replaces on template:

* `.Scopes` - scopes required by client app, this var is only provided when an app requires the end-user to grant scopes
* `.ClientName` - client app name, this var is only provided when an app requires the end-user to grant scopes
* `.GrantNonce` - nonce to be passed to grant request, this var is only provided when an app requires the end-user to grant scopes
* `.RequireSignIn` - if the session is invalid or expired this value will be true indicating that the end-user must sign-in again to activate the current session

You can see the example of this template usage in [web/index.html](web/index.html). The configuration is done through environment vars:

Environment Var                  | Type     | Description                                                                                                                        | Default Value
---------------------------------|----------|------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------
`SSO_NAMESPACE`                  | string   | Namespace to prefix cookies                                                                                                        | `"SSO"`
`SSO_ISSUER`                     | string   | **iss** field in generated JWTs                                                                                                    | `"accounts.savesafe.app"`
`SSO_TEMPLATE_PATH`              | string   | Path to HTML template file                                                                                                         | `"./index.html"`
`SSO_RECAPTCHA_SITE_KEY`         | string   | Recaptcha site key (v2 or v3)                                                                                                      | `"6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI"`
`SSO_RECAPTCHA_SECRET_KEY`       | string   | Recaptcha secret key (v2 or v3)                                                                                                    | `"6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe"`
`SSO_PRIVATE_KEY_PATH`           | string   | Path to private key to sign JWTs (it'll be auto generated if it's empty)                                                           | `""`
`SSO_PRIVATE_KEY_PASSWORD`       | string   | Private key password, leave it empty if there is no password                                                                       | `""`
`SSO_HTTP_BIND`                  | string   | IP and port to bind HTTP(S) server                                                                                                 | `"0.0.0.0:3230"`
`SSO_HTTP_SHUTDOWN_TIMEOUT`      | duration | Maximum time to wait HTTP(S) server to shutdown                                                                                    | `"10000ms"`
`SSO_HTTP_REQUEST_DEADLINE`      | duration | Maximum duration of each HTTP request                                                                                              | `"2s"`
`SSO_HTTPS_PRIVATE_KEY`          | string   | Private key to enable HTTPS server, empty will provide a plain HTTP server                                                         | `""`
`SSO_HTTPS_CERTIFICATE`          | string   | Certificate to enable HTTPS server, empty will provide a plain HTTP server                                                         | `""`
`SSO_ALLOWED_ORIGINS`            | string   | CORS allowed origins                                                                                                               | `"https://localhost:3230"`
`SSO_GRPC_SERVER_ADDR`           | string   | Address of SSO Service                                                                                                             | `"127.0.0.1:3231"`
`SSO_PRETTY_LOG`                 | boolean  | Enable pretty log print, otherwise it'll be printed as JSON lines                                                                  | `"true"`
`SSO_LOG_LEVEL`                  | string   | Log level, can be: `debug, info, warn, error`                                                                                      | `"debug"`
`SSO_REMEMBER_ME_TIMEOUT`        | duration | If it's > 0 it'll be the session duration of a remember-me enabled session, otherwise all sessions will be considered as ephemeral | `"0s"`
`SSO_EPHEMERAL_SESSION_DURATION` | duration | Maximum duration of ephemeral sessions (normal sessions without remember-me)                                                       | `"15m"`
`SSO_MFA_SESSION_TIMEOUT`        | duration | Maximum duration of a MFA code                                                                                                     | `"3m"`
`SSO_GRANT_REQUEST_TIMEOUT`      | duration | Maximum duration of a grant nonce                                                                                                  | `"3m"`
`SSO_AUTHORIZATION_CODE_TIMEOUT` | duration | Maximum duration of a authorization code (authorization code OAuth2 flow)                                                          | `"3m"`
`SSO_REFRESH_TOKEN_DURATION`     | duration | Duration of refresh tokens                                                                                                         | `"60m"`
`SSO_ACCESS_TOKEN_DURATION`      | duration | Duration of access tokens                                                                                                          | `"5m"`
`SSO_SESSION_COOKIE_DOMAIN`      | duration | Domain specified when setting cookies, empty and it won't be specified                                                             | `""`
`SSO_SESSION_COOKIE_PATH`        | duration | Path specified when setting cookies                                                                                                | `"/"`
`SSO_REDIS_PREFIX`               | string   | Prefix to be used when setting keys in Redis                                                                                       | `"sso"`
`SSO_REDIS_SENTINEL`             | boolean  | Enables Redis Sentinel connection                                                                                                  | `"false"`
`SSO_REDIS_NODES`                | strings  | When using Redis Sentinel specify the nodes here                                                                                   | `""`
`SSO_REDIS_PASSWORD`             | string   | Redis password                                                                                                                     | `""`
`SSO_REDIS_DB`                   | int      | Redis database to pick                                                                                                             | `"0"`
`SSO_REDIS_MASTER`               | string   | Redis Sentinel master to use                                                                                                       | `""`
`SSO_REDIS_MAX_POOL_SIZE`        | int      | Maximum number of Redis connections on pool                                                                                        | `"5"`
`SSO_REDIS_MIN_POOL_SIZE`        | int      | Minimum number of Redis connections on pool                                                                                        | `"1"`
`SSO_REDIS_NODE`                 | int      | When using Redis standalone specify the address here                                                                               | `"127.0.0.1:6379"`