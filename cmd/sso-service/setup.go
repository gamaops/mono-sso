package main

import (
	"context"
	stdlog "log"
	"os"
	"time"

	"math/rand"

	"github.com/gamaops/mono-sso/pkg/cache"
	"github.com/gamaops/mono-sso/pkg/datastore"
	"github.com/go-redis/redis"
	logrus "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var log = logrus.New()
var ServiceCache *cache.Cache
var ServiceDatastore *datastore.Datastore

func setup() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	// MongoDB information
	viper.SetDefault("mongodbUri", "")
	viper.BindEnv("mongodbUri", "SSO_MONGODB_URI")
	viper.SetDefault("mongodbConnectTimeout", "15s")
	viper.BindEnv("mongodbConnectTimeout", "SSO_MONGODB_CONNECT_TIMEOUT")
	viper.SetDefault("mongodbDatabase", "sso")
	viper.BindEnv("mongodbDatabase", "SSO_MONGODB_DATABASE")
	viper.SetDefault("mongodbShutdownTimeout", "5s")
	viper.BindEnv("mongodbShutdownTimeout", "SSO_MONGODB_SHUTDOWN_TIMEOUT")

	// gRPC Server information
	viper.SetDefault("grpcListen", "0.0.0.0:3231")
	viper.BindEnv("grpcListen", "SSO_GRPC_LISTEN")
	viper.SetDefault("grpcMaxKeepAlive", "2m")
	viper.BindEnv("grpcListen", "SSO_GRPC_LISTEN")

	// Redis information
	viper.SetDefault("redisPrefix", "sso")
	viper.BindEnv("redisPrefix", "SSO_REDIS_PREFIX")
	viper.SetDefault("redisSentinel", "false")
	viper.BindEnv("redisSentinel", "SSO_REDIS_SENTINEL")
	viper.SetDefault("redisNodes", "")
	viper.BindEnv("redisNodes", "SSO_REDIS_NODES")
	viper.SetDefault("redisPassword", "")
	viper.BindEnv("redisPassword", "SSO_REDIS_PASSWORD")
	viper.SetDefault("redisDb", "0")
	viper.BindEnv("redisDb", "SSO_REDIS_DB")
	viper.SetDefault("redisMaster", "")
	viper.BindEnv("redisMaster", "SSO_REDIS_MASTER")
	viper.SetDefault("redisMaxPoolSize", "5")
	viper.BindEnv("redisMaxPoolSize", "SSO_REDIS_MAX_POOL_SIZE")
	viper.SetDefault("redisMinPoolSize", "1")
	viper.BindEnv("redisMinPoolSize", "SSO_REDIS_MIN_POOL_SIZE")
	viper.SetDefault("redisNode", "127.0.0.1:6379")
	viper.BindEnv("redisNode", "SSO_REDIS_NODE")

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
			MongoDBURI:    viper.GetString("mongodbUri"),
			MongoDatabase: viper.GetString("mongodbDatabase"),
		},
		Logger: log,
	}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("mongodbConnectTimeout"))
	defer cancel()
	err := datastore.StartDatastore(ctx, ServiceDatastore)

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
