package main

import (
	"context"
	stdlog "log"
	"os"
	"reflect"
	"strings"
	"time"

	"math/rand"

	"github.com/gamaops/mono-sso/pkg/cache"
	"github.com/gamaops/mono-sso/pkg/datastore"
	"github.com/gamaops/mono-sso/pkg/oauth2"
	"github.com/go-redis/redis"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
)

var log = logrus.New()
var enableSessionValidation bool
var ServiceCache *cache.Cache
var ServiceDatastore *datastore.Datastore
var ServiceOAuth2Jose *oauth2.OAuth2Jose

func setup() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	// Postgres information
	viper.SetDefault("postgresUri", "")
	viper.BindEnv("postgresUri", "SSO_POSTGRES_URI")
	viper.SetDefault("postgresMaxConn", "5")
	viper.BindEnv("postgresMaxConn", "SSO_POSTGRES_MAX_CONN")

	// SSO Manager app setup information
	viper.SetDefault("clientAppName", "SSO Manager")
	viper.BindEnv("clienAppName", "SSO_CLIENT_APP_NAME")
	viper.SetDefault("clientAppRedirectUris", "https://localhost:3230/sign-in")
	viper.BindEnv("clientAppRedirectUris", "SSO_CLIENT_APP_REDIRECT_URIS")
	viper.SetDefault("adminAccountName", "SSO Administrator")
	viper.BindEnv("adminAccountName", "SSO_ADMIN_ACCOUNT_NAME")
	viper.SetDefault("adminAccountIdentifier", "sso_admin")
	viper.BindEnv("adminAccountIdentifier", "SSO_ADMIN_ACCOUNT_IDENTIFIER")
	viper.SetDefault("adminAccountPassword", "sso#mono@6014")
	viper.BindEnv("adminAccountPassword", "SSO_ADMIN_ACCOUNT_PASSWORD")
	viper.SetDefault("adminTenant", "appk20hhh83hag8fbk1rad7o5qkq")
	viper.BindEnv("adminTenant", "SSO_ADMIN_TENANT")
	viper.SetDefault("setupTimeout", "10s")
	viper.BindEnv("setupTimeout", "SSO_SETUP_TIMEOUT")

	// SSO Manager session
	viper.SetDefault("enableSessionValidation", "true")
	viper.BindEnv("enableSessionValidation", "SSO_ENABLE_SESSION_VALIDATION")
	viper.SetDefault("sessionAudience", "")
	viper.BindEnv("sessionAudience", "SSO_SESSION_AUDIENCE")
	viper.SetDefault("sessionPastToleration", "-15s")
	viper.BindEnv("sessionPastToleration", "SSO_SESSION_PAST_TOLERATION")
	viper.SetDefault("jwksUrl", "https://localhost:3230/.well-known/jwks.json")
	viper.BindEnv("jwksUrl", "SSO_JWKS_URL")

	// gRPC Server information
	viper.SetDefault("grpcListen", "0.0.0.0:3232")
	viper.BindEnv("grpcListen", "SSO_GRPC_LISTEN")
	viper.SetDefault("grpcMaxKeepAlive", "2m")
	viper.BindEnv("grpcMaxKeepAlive", "SSO_GRPC_KEEP_ALIVE")

	// Redis information
	cache.SetupViper()

	// Logging
	viper.SetDefault("prettyLog", "true")
	viper.BindEnv("prettyLog", "SSO_PRETTY_LOG")
	viper.SetDefault("logLevel", "debug")
	viper.BindEnv("logLevel", "SSO_LOG_LEVEL")

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

	enableSessionValidation = viper.GetBool("enableSessionValidation")

	if enableSessionValidation {
		ServiceOAuth2Jose = &oauth2.OAuth2Jose{
			Options: &oauth2.Options{
				JWKSURL: viper.GetString("jwksUrl"),
			},
		}
		err := oauth2.LoadOAuth2JoseFromURL(ServiceOAuth2Jose)
		if err != nil {
			log.Fatalf("Error when loading JWK from URL: %v", err)
		}
		setupClientServer()
	}

	log.SetOutput(os.Stdout)

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

	ServiceDatastore = &datastore.Datastore{
		Options: &datastore.Options{
			PostgresURI:    viper.GetString("postgresUri"),
			MaxConnections: viper.GetInt("postgresMaxConn"),
			Validator:      validator.New(),
		},
		Logger: log,
	}

	ServiceDatastore.Options.Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("setupTimeout"))
	defer cancel()

	err := datastore.StartDatastore(ctx, ServiceDatastore)

	if err != nil {
		log.Fatalf("Error starting datastore: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	err = registerSSOManagerApp(ctx)
	if err != nil {
		log.Fatalf("Error setting up administration assets: %v", err)
	}

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
