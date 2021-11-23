package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

// requestIDHeader is a X-RequestID header.
const requestIDHeader = "X-RequestID"

// LoggerMiddleware is a middleware that will log requestID taken from HTTP header (X-RequestID) and request time.
// If there is no request ID it will generate one.
func LoggerMiddleware(log *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return requestLogger(log)(next)
	}
}

// requestLogger returns a logger handler.
func requestLogger(log *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			log.Infof("%s: %s\n", time.Now(), requestID)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
