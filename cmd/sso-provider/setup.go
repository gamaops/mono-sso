package main

import (
	stdlog "log"
	"net/http"
	"os"

	recaptcha "github.com/dpapathanasiou/go-recaptcha"
	"github.com/gamaops/mono-sso/pkg/cache"
	"github.com/gamaops/mono-sso/pkg/handlers"
	httpserver "github.com/gamaops/mono-sso/pkg/http-server"
	"github.com/gamaops/mono-sso/pkg/oauth2"
	"github.com/gamaops/mono-sso/pkg/session"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	logrus "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var log = logrus.New()
var ServiceCache *cache.Cache
var ServiceHTTPServer *httpserver.HTTPServer
var ServiceOAuth2Jose *oauth2.OAuth2Jose
var ServiceAuthenticationModel *session.AuthenticationModel
var ServiceAuthorizationModel *session.AuthorizationModel
var Router *mux.Router = mux.NewRouter()

func setup() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	// Namespace is used to prefix cookies and other things to avoid collision between two instances of SSO provider
	viper.SetDefault("namespace", "SSO")
	viper.BindEnv("namespace", "SSO_NAMESPACE")

	// Issuer to use as "iss" in JWTs
	viper.SetDefault("issuer", "accounts.savesafe.app")
	viper.BindEnv("issuer", "SSO_ISSUER")

	// Path with the index.html file, must end with /
	viper.SetDefault("templatePath", "./index.html")
	viper.BindEnv("templatePath", "SSO_TEMPLATE_PATH")

	// Recaptcha keys, the default keys are just for testing purpose
	viper.SetDefault("recaptchaSiteKey", "6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI")
	viper.BindEnv("recaptchaSiteKey", "SSO_RECAPTCHA_SITE_KEY")
	viper.SetDefault("recaptchaSecretKey", "6LeIxAcTAAAAAGG-vFI1TnRWxMZNFuojJ4WifJWe")
	viper.BindEnv("recaptchaSecretKey", "SSO_RECAPTCHA_SECRET_KEY")

	// Key used to provide JWK and sign JWTs
	viper.SetDefault("privateKeyPath", "")
	viper.BindEnv("privateKeyPath", "SSO_PRIVATE_KEY_PATH")
	viper.SetDefault("privateKeyPassword", "")
	viper.BindEnv("privateKeyPassword", "SSO_PRIVATE_KEY_PASSWORD")

	// SSO provider server
	viper.SetDefault("httpBind", "0.0.0.0:3230")
	viper.BindEnv("httpBind", "SSO_HTTP_BIND")
	viper.SetDefault("httpShutdownTimeout", "10000ms")
	viper.BindEnv("httpShutdownTimeout", "SSO_HTTP_SHUTDOWN_TIMEOUT")
	viper.SetDefault("requestDeadline", "2s")
	viper.BindEnv("requestDeadline", "SSO_HTTP_REQUEST_DEADLINE")
	viper.SetDefault("httpsPrivateKey", "")
	viper.BindEnv("httpsPrivateKey", "SSO_HTTPS_PRIVATE_KEY")
	viper.SetDefault("httpsCertificate", "")
	viper.BindEnv("httpsCertificate", "SSO_HTTPS_CERTIFICATE")

	// CORS
	viper.SetDefault("allowedOrigins", "https://localhost:3230")
	viper.BindEnv("allowedOrigins", "SSO_ALLOWED_ORIGINS")

	// SSO Service gRPC server address
	viper.SetDefault("grpcServerAddr", "127.0.0.1:3231")
	viper.BindEnv("grpcServerAddr", "SSO_GRPC_SERVER_ADDR")

	// Redis information
	cache.SetupViper()

	// Logging
	viper.SetDefault("prettyLog", "true")
	viper.BindEnv("prettyLog", "SSO_PRETTY_LOG")
	viper.SetDefault("logLevel", "debug")
	viper.BindEnv("logLevel", "SSO_LOG_LEVEL")

	// Session settings
	viper.SetDefault("rememberMeTimeout", "0s") // If this parameter is 0 the remember me option is disabled
	viper.BindEnv("rememberMeTimeout", "SSO_REMEMBER_ME_TIMEOUT")
	viper.SetDefault("ephemeralSessionDuration", "15m")
	viper.BindEnv("ephemeralSessionDuration", "SSO_EPHEMERAL_SESSION_DURATION")
	viper.SetDefault("mfaSessionTimeout", "3m") // Time to keep session when the SSO ask user for activation code
	viper.BindEnv("mfaSessionTimeout", "SSO_MFA_SESSION_TIMEOUT")
	viper.SetDefault("grantRequestTimeout", "3m") // Time to keep grant request cache
	viper.BindEnv("grantRequestTimeout", "SSO_GRANT_REQUEST_TIMEOUT")
	viper.SetDefault("authorizationCodeTimeout", "3m") // Time to keep authorization code
	viper.BindEnv("authorizationCodeTimeout", "SSO_AUTHORIZATION_CODE_TIMEOUT")
	viper.SetDefault("refreshTokenDuration", "60m")
	viper.BindEnv("refreshTokenDuration", "SSO_REFRESH_TOKEN_DURATION")
	viper.SetDefault("accessTokenDuration", "5m")
	viper.BindEnv("accessTokenDuration", "SSO_ACCESS_TOKEN_DURATION")
	viper.SetDefault("sessionCookieDomain", "")
	viper.BindEnv("sessionCookieDomain", "SSO_SESSION_COOKIE_DOMAIN")
	viper.SetDefault("sessionCookiePath", "/")
	viper.BindEnv("sessionCookiePath", "SSO_SESSION_COOKIE_PATH")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Unable to load configuration file: %v", err)
		}
	}

	if viper.GetBool("prettyLog") {
		log.SetFormatter(&logrus.TextFormatter{})
	} else {
		log.SetFormatter(&logrus.JSONFormatter{})
	}

	log.SetOutput(os.Stdout)

	recaptcha.Init(viper.GetString("recaptchaSecretKey"))

	ServiceOAuth2Jose = &oauth2.OAuth2Jose{
		Options: &oauth2.Options{
			PrivateKeyPath:     viper.GetString("privateKeyPath"),
			PrivateKeyPassword: viper.GetString("privateKeyPassword"),
		},
		Logger: log,
	}

	err := oauth2.SetupOAuth2Jose(ServiceOAuth2Jose)

	if err != nil {
		log.Fatalf("Error when setting up jose: %v", err)
	}

	ServiceCache = &cache.Cache{
		Options: &cache.Options{
			Prefix:      viper.GetString("redisPrefix"),
			Password:    viper.GetString("redisPassword"),
			DB:          viper.GetInt("redisDb"),
			Nodes:       viper.GetStringSlice("redisNodes"),
			Node:        viper.GetString("redisNode"),
			MasterName:  viper.GetString("redisMaster"),
			Sentinel:    viper.GetBool("redisSentinel"),
			MaxPoolSize: viper.GetInt("redisMaxPoolSize"),
			MinPoolSize: viper.GetInt("redisMinPoolSize"),
		},
		Logger: log,
	}

	cacheLogger := stdlog.New(logrus.StandardLogger().Writer(), "", 0)
	redis.SetLogger(cacheLogger)

	cache.StartCacheClient(ServiceCache)

	startGrpcClient()

	ServiceAuthorizationModel = &session.AuthorizationModel{
		Options: &session.AuthorizationOptions{
			GrantRequestDuration:       viper.GetDuration("grantRequestTimeout"),
			AuthorizationCodeDuration:  viper.GetDuration("authorizationCodeTimeout"),
			AccessTokenDuration:        viper.GetDuration("accessTokenDuration"),
			RefreshTokenDuration:       viper.GetDuration("refreshTokenDuration"),
			Issuer:                     viper.GetString("issuer"),
			AuthorizationServiceClient: authorizationServiceClient,
		},
	}

	ServiceAuthenticationModel = &session.AuthenticationModel{
		Options: &session.AuthenticationOptions{
			IndexTemplatePath:        viper.GetString("templatePath"),
			Namespace:                viper.GetString("namespace"),
			RememberMeDuration:       viper.GetDuration("rememberMeTimeout"),
			EphemeralSessionDuration: viper.GetDuration("ephemeralSessionDuration"),
			MFASessionDuration:       viper.GetDuration("mfaSessionTimeout"),
			SessionCookieDomain:      viper.GetString("sessionCookieDomain"),
			SessionCookiePath:        viper.GetString("sessionCookiePath"),
			AccountServiceClient:     accountServiceClient,
		},
		Logger: log,
	}

	err = session.SetupAuthenticationModel(ServiceAuthenticationModel)
	if err != nil {
		log.Fatalf("Error while setting up authentication model: %v", err)
	}

	ServiceHTTPServer = &httpserver.HTTPServer{
		Options: &httpserver.Options{
			HTTPBind:        viper.GetString("httpBind"),
			PrivateKeyPath:  viper.GetString("httpsPrivateKey"),
			CertificatePath: viper.GetString("httpsCertificate"),
			ShutdownTimeout: viper.GetDuration("httpShutdownTimeout"),
			RequestDeadline: viper.GetDuration("requestDeadline"),
		},
		Logger: log,
	}

	httpCors := cors.New(cors.Options{
		AllowedOrigins:   viper.GetStringSlice("allowedOrigins"),
		AllowCredentials: true,
		AllowedMethods:   []string{"POST", "GET"},
		Debug:            false,
	})

	Router.HandleFunc("/.well-known/jwks.json", handlers.JWKSHandler(ServiceOAuth2Jose))

	signInAuthenticateHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.AuthenticateHandler(ServiceHTTPServer, ServiceAuthenticationModel, ServiceCache, w, r)
	})
	Router.Handle("/sign-in/authenticate", httpCors.Handler(signInAuthenticateHandler))

	signInActivateHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.ActivateSessionHandler(ServiceHTTPServer, ServiceAuthenticationModel, ServiceCache, w, r)
	})
	Router.Handle("/sign-in/activate", httpCors.Handler(signInActivateHandler))

	Router.HandleFunc("/sign-in", func(w http.ResponseWriter, r *http.Request) {
		handlers.IndexHandler(
			ServiceHTTPServer,
			ServiceAuthorizationModel,
			ServiceAuthenticationModel,
			ServiceOAuth2Jose,
			ServiceCache,
			w,
			r,
		)
	})

	signInAuthorizeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.GrantScopesHandler(
			ServiceHTTPServer,
			ServiceAuthorizationModel,
			ServiceAuthenticationModel,
			ServiceCache,
			w,
			r,
		)
	})
	Router.Handle("/sign-in/authorize", httpCors.Handler(signInAuthorizeHandler))

	Router.HandleFunc("/sign-in/exchange", func(w http.ResponseWriter, r *http.Request) {
		handlers.ExchangeHandler(
			ServiceHTTPServer,
			ServiceAuthorizationModel,
			ServiceOAuth2Jose,
			ServiceCache,
			w,
			r,
		)
	})
	Router.HandleFunc("/sign-in/token", func(w http.ResponseWriter, r *http.Request) {
		handlers.RefreshTokenHandler(
			ServiceHTTPServer,
			ServiceAuthorizationModel,
			ServiceOAuth2Jose,
			w,
			r,
		)
	})
	http.Handle("/", Router)

	httpserver.StartServer(ServiceHTTPServer)

	log.SetLevel(logrus.DebugLevel)

	switch viper.GetString("logLevel") {
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	}

}
