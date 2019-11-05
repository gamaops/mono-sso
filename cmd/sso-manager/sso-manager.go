package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gamaops/mono-sso/pkg/cache"
	"github.com/gamaops/mono-sso/pkg/datastore"
	"github.com/spf13/viper"
)

var waitShutdown = sync.WaitGroup{}

func main() {

	waitShutdown.Add(1)

	setup()
	go startGrpcServer()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Warnf("Stopping server, signal received: %v", sig)
		stopGrpcServer()
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("mongodbShutdownTimeout"))
		defer cancel()
		datastore.StopDataStore(ctx, ServiceDatastore)
		cache.StopCacheClient(ServiceCache)
		waitShutdown.Done()
	}()

	waitShutdown.Wait()

}
