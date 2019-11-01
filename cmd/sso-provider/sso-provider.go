package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gamaops/mono-sso/pkg/cache"
	httpserver "github.com/gamaops/mono-sso/pkg/http-server"
)

var waitShutdown = sync.WaitGroup{}

func main() {

	waitShutdown.Add(1)

	setup()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Warnf("Stopping server, signal received: %v", sig)
		httpserver.StopServer(ServiceHTTPServer)
		cache.StopCacheClient(ServiceCache)
		waitShutdown.Done()
	}()

	waitShutdown.Wait()

}
