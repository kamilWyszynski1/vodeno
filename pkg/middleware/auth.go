package middleware

import (
	"net/http"
)

const (
	xTokenHeader = "X-Token" // X-Token header.
	testToken    = "test"    // only valid value for xTokenHeader.
)

// AuthenticationMiddleware will mock authentication.
// It expects an X-Token HTTP header.
// It returns 401 on X-Token different that testToken.
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xToken := r.Header.Get(xTokenHeader)
		if xToken != testToken {
			authFailed(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authFailed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
}
