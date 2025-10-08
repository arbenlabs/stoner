package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/arbenlabs/stoner/logger"

	"github.com/gorilla/csrf"
	"golang.org/x/time/rate"
)

type Middleware struct {
	logger            *logger.Logger
	RateLimiter       *rate.Limiter
	MaxRequestSize    int64
	MaxHeaderSize     int64
	MaxFileUploadSize int64
	ReadTimeout       int
	WriteTimeout      int
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// NewMiddleware creates a new middleware.
func NewMiddleware(
	rateLimitRequestsPerSecond int,
	rateLimitBurst int,
	maxRequestSize int64,
	maxHeaderSize int64,
	maxFileUploadSize int64,
	readTimeout int,
	writeTimeout int,
) *Middleware {
	return &Middleware{
		RateLimiter:       rate.NewLimiter(rate.Limit(rateLimitRequestsPerSecond), rateLimitBurst),
		MaxRequestSize:    maxRequestSize,
		MaxHeaderSize:     maxHeaderSize,
		MaxFileUploadSize: maxFileUploadSize,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
	}
}

// RateLimit is a middleware that limits the number of requests per second.
func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.RateLimiter.Allow() {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		}
	})
}

// RequestSizeLimitMiddleware limits request body size
func (m *Middleware) RequestSizeLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check content length
			if r.ContentLength > m.MaxRequestSize {
				http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
				return
			}

			// Limit request body
			r.Body = http.MaxBytesReader(w, r.Body, m.MaxRequestSize)

			next.ServeHTTP(w, r)
		})
	}
}

// RequestTimeoutMiddleware adds request timeout
func (m *Middleware) RequestTimeout() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Duration(m.ReadTimeout)*time.Second)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggerMiddleware logs HTTP requests
func (m *Middleware) LogHTTRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		m.logger.LogHTTPRequest(
			r.Method,
			r.URL.Path,
			r.UserAgent(),
			r.RemoteAddr,
			r.Header.Get("Content-Type"),
		)

		m.logger.LogPerformance("http_request", duration, map[string]interface{}{
			"status_code": wrapped.statusCode,
			"method":      r.Method,
			"path":        r.URL.Path,
		})
	})
}

// CSRFMiddleware implements CSRF protection
func (m *Middleware) CSRFMiddleware(authKey []byte, secure bool) func(http.Handler) http.Handler {
	return csrf.Protect(authKey, csrf.Secure(secure))
}
