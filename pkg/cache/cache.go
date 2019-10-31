package cache

import (
	"errors"
	"strings"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

var ErrInvalidIDPart = errors.New("invalid cache key ID part, must be string, []byte or rune")

type Options struct {
	Prefix      string
	Password    string
	DB          int
	MasterName  string
	MaxPoolSize int
	MinPoolSize int
	Nodes       []string
	Node        string
	Sentinel    bool
}

type Cache struct {
	Options *Options
	Client  *redis.Client
	Logger  *logrus.Logger
}

func (c *Cache) CreateID(parts ...interface{}) (*strings.Builder, error) {
	id := &strings.Builder{}
	id.WriteString(c.Options.Prefix)
	var err error = nil
	for _, part := range parts {
		switch value := part.(type) {
		case string:
			_, err = id.WriteString(value)
		case rune:
			_, err = id.WriteRune(value)
		case []byte:
			_, err = id.Write(value)
		default:
			return nil, ErrInvalidIDPart
		}
		if err != nil {
			return nil, err
		}
	}
	return id, nil
}

func StartCacheClient(cache *Cache) {

	if cache.Options.Sentinel {
		cache.Client = redis.NewFailoverClient(&redis.FailoverOptions{
			Password:      cache.Options.Password,
			DB:            cache.Options.DB,
			MasterName:    cache.Options.MasterName,
			SentinelAddrs: cache.Options.Nodes,
			PoolSize:      cache.Options.MaxPoolSize,
			MinIdleConns:  cache.Options.MinPoolSize,
		})
	}

	cache.Client = redis.NewClient(&redis.Options{
		Addr:         cache.Options.Node,
		Password:     cache.Options.Password,
		DB:           cache.Options.DB,
		Network:      "tcp",
		PoolSize:     cache.Options.MaxPoolSize,
		MinIdleConns: cache.Options.MinPoolSize,
	})

}

func StopCacheClient(cache *Cache) {

	cache.Logger.Warn("Closing Redis client")

	err := cache.Client.Close()

	if err != nil {
		cache.Logger.Errorf("Error while closing Redis client: %v", err)
	}

	cache.Logger.Warn("Redis client closed")

}
