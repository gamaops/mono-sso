package main

import (
	"context"
	stdlog "log"
	"os"
	"time"

	"math/rand"

	"github.com/gamaops/gamago/pkg/id"
	"github.com/gamaops/mono-sso/pkg/cache"
	"github.com/gamaops/mono-sso/pkg/datastore"
	"github.com/go-redis/redis"
	logrus "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var log = logrus.New()
var ServiceCache *cache.Cache
var ServiceDatastore *datastore.Datastore
var tokenIDGenerator *id.IDGenerator

func setup() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	// Postgres information
	viper.SetDefault("postgresUri", "")
	viper.BindEnv("postgresUri", "SSO_POSTGRES_URI")
	viper.SetDefault("postgresMaxConn", "5")
	viper.BindEnv("postgresMaxConn", "SSO_POSTGRES_MAX_CONN")

	// gRPC Server information
	viper.SetDefault("grpcListen", "0.0.0.0:3231")
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

	log.SetOutput(os.Stdout)

	var err error
	tokenIDGenerator, err = id.NewIDGenerator(10)

	if err != nil {
		log.Fatalf("Error initializing ID generator: %v", err)
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

	ServiceDatastore = &datastore.Datastore{
		Options: &datastore.Options{
			PostgresURI:    viper.GetString("postgresUri"),
			MaxConnections: viper.GetInt("postgresMaxConn"),
		},
		Logger: log,
	}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("mongodbConnectTimeout"))
	defer cancel()
	err = datastore.StartDatastore(ctx, ServiceDatastore)

	if err != nil {
		log.Fatalf("Error starting datastore: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

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
