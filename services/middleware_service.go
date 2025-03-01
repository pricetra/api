package services

import (
	"context"
	"net/http"

	"github.com/pricetra/api/types"
)

// Extracts the value from "Authorization" header and stores
// it within the request context, with key `types.AuthorizationKey`
func (Service) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		r = r.WithContext(context.WithValue(
			r.Context(), 
			types.AuthorizationKey, 
			types.AuthorizationKeyType(authorization),
		))
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
