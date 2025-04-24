package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/google/uuid"
)

type key string

const requestIDKey key = "x-request-id"

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			logger.Info("Incoming request",
				zap.String("method", r.Method),
				zap.String("url", r.URL.Path),
				zap.String("request_id", requestID),
			)

			// Обернем writer, если хочешь логировать ещё и ответ
			// Можно добавить log response status в будущем

			next.ServeHTTP(w, r)
		})
	}
}

// Получение ID из контекста
func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(requestIDKey).(string); ok {
		return val
	}
	return ""
}
