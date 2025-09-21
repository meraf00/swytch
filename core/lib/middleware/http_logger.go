package middleware

import (
	"context"
	"maps"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/meraf00/swytch/core/lib/logger"
)

type startTimeKey string
type idKey string

const (
	reqStartTimeKey startTimeKey = "reqStartTime"
	reqIDKey        idKey        = "reqID"
)

type loggerKey struct{}

type LoggerOptions struct {
	IgnorePaths  []string
	CustomProps  func(*http.Request, *http.Response) map[string]any
	GenRequestID func() string
}

func DefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		IgnorePaths: []string{"/favicon.ico"},
		GenRequestID: func() string {
			return uuid.New().String()
		},
	}
}

func HTTPLoggerMiddleware(log logger.Log, opts LoggerOptions) mux.MiddlewareFunc {
	if opts.GenRequestID == nil {
		opts.GenRequestID = DefaultLoggerOptions().GenRequestID
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip ignored paths
			if slices.Contains(opts.IgnorePaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Start time and request ID
			start := time.Now()
			reqID := opts.GenRequestID()
			ctx := context.WithValue(r.Context(), reqStartTimeKey, start)
			ctx = context.WithValue(ctx, reqIDKey, reqID)

			logger := log.WithFields(map[string]any{
				"reqID":     reqID,
				"method":    r.Method,
				"path":      r.URL.Path,
				"ip":        r.RemoteAddr,
				"userAgent": r.UserAgent(),
			})

			ctx = context.WithValue(ctx, loggerKey{}, logger)
			r = r.WithContext(ctx)

			// Wrap ResponseWriter to capture status
			rw := &responseWriter{w, http.StatusOK}

			defer func() {
				// Duration
				duration := time.Since(start)

				// Base fields
				fields := map[string]any{
					"status":     rw.status,
					"durationMs": duration.Milliseconds(),
				}

				// Add custom props if provided
				if opts.CustomProps != nil {
					maps.Copy(fields, opts.CustomProps(r, rw.response()))
				}

				// Create request-scoped logger
				requestLogger := logger.HTTP().WithFields(fields)

				// Log based on status code
				switch {
				case rw.status >= 500:
					requestLogger.Error("server error")
				case rw.status >= 400:
					requestLogger.Warn("client error")
				case rw.status >= 300:
					requestLogger.Info("redirection")
				default:
					requestLogger.Info("request completed")
				}
			}()

			// Process request
			next.ServeHTTP(rw, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) response() *http.Response {
	return &http.Response{
		StatusCode: rw.status,
	}
}

// Context helpers
func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(reqIDKey).(string); ok {
		return id
	}
	return ""
}

func GetRequestStartTime(r *http.Request) time.Time {
	if start, ok := r.Context().Value(reqStartTimeKey).(time.Time); ok {
		return start
	}
	return time.Time{}
}

func GetLogger(ctx context.Context) *logger.Log {
	if l, ok := ctx.Value(loggerKey{}).(*logger.Log); ok {
		return l
	}
	return nil
}
