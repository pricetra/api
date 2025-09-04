package services

import (
	"context"
	"net/http"

	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/types"
)

// Extracts the value from "Authorization" header and stores
// it within the request context, with key `types.AuthorizationKey`
func (s Service) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		bearer_token := types.AuthorizationKeyType(authorization)
		r = r.WithContext(context.WithValue(
			r.Context(), 
			types.AuthorizationKey, 
			bearer_token,
		))

		// If valid JWT token, store the user info in context with key `types.AuthUserKey`
		if user, err := s.VerifyJwt(r.Context(), bearer_token); err == nil && user != (gmodel.User{}) {
			r = r.WithContext(context.WithValue(
				r.Context(), 
				types.AuthUserKey, 
				user,
			))
		}
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
