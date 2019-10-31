package httpserver

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"sync"

	"time"

	"github.com/sirupsen/logrus"
)

type Options struct {
	HttpBind        string
	PrivateKeyPath  string
	CertificatePath string
	ShutdownTimeout time.Duration
}

type HTTPServer struct {
	Options      *Options
	Server       *http.Server
	WaitShutdown sync.WaitGroup
	Logger       *logrus.Logger
}

func StartServer(httpServer *HTTPServer) {
	httpServer.WaitShutdown = sync.WaitGroup{}
	httpServer.Server = &http.Server{
		Addr:              httpServer.Options.HttpBind,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    30720,
	}

	go func() {
		if len(httpServer.Options.PrivateKeyPath) > 0 && len(httpServer.Options.CertificatePath) > 0 {
			httpServer.Server.TLSConfig = &tls.Config{
				PreferServerCipherSuites: true,
				MinVersion:               13,
				ClientAuth:               tls.NoClientCert,
				Renegotiation:            tls.RenegotiateNever,
			}
			err := httpServer.Server.ListenAndServeTLS(httpServer.Options.CertificatePath, httpServer.Options.PrivateKeyPath)
			if err != nil && err != http.ErrServerClosed {
				httpServer.Logger.Fatalf("Error starting server (https enabled): %v", err)
			}
			return
		}
		err := httpServer.Server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			httpServer.Logger.Fatalf("Error starting server: %v", err)
		}
	}()

	httpServer.WaitShutdown.Add(1)

}

type HTTPClientIPs struct {
	ClientIP     string
	SourceIP     string
	ForwardedIPs []string
}

func ClientIPsFromRequest(r *http.Request) (*HTTPClientIPs, error) {
	cIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}
	if cIP == "::1" {
		cIP = "127.0.0.1"
	}
	clientIPs := &HTTPClientIPs{
		ClientIP: cIP,
	}
	fIP := r.Header.Get("X-Forwarded-For")
	if len(fIP) > 0 {
		clientIPs.ForwardedIPs = strings.Split(fIP, ", ")
		if clientIPs.ForwardedIPs[0] == "::1" {
			clientIPs.ForwardedIPs[0] = "127.0.0.1"
		}
		clientIPs.SourceIP = clientIPs.ForwardedIPs[0]
		return clientIPs, nil
	}
	clientIPs.ForwardedIPs = []string{}
	clientIPs.SourceIP = cIP
	return clientIPs, nil
}

func StopServer(httpServer *HTTPServer) {
	httpServer.Logger.Warn("Stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), httpServer.Options.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Server.Shutdown(ctx); err != nil {
		httpServer.Logger.Errorf("Error while stopping server: %v", err)
	}
	httpServer.WaitShutdown.Done()
}
