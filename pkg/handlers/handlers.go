package handlers

import (
	"net/http"

	"github.com/gamaops/mono-sso/pkg/oauth2"
)

func JWKSHandler(oauth2jose *oauth2.OAuth2Jose) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(oauth2jose.JWKS)
	}
}
