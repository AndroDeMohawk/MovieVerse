package middleware

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter ограничивает количество запросов от одного IP за заданное окно времени.
func RateLimiter(rdb *redis.Client, limit int, window time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			clientIP := r.Header.Get("X-Forwarded-For")
			if clientIP == "" {
				clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
				if clientIP == "" {
					clientIP = r.RemoteAddr
				}
			}

			now := time.Now().Unix()
			windowSec := int64(window.Seconds())
			currentWindow := now / windowSec

			key := fmt.Sprintf("ratelimit:%s:%d", clientIP, currentWindow)

			pipe := rdb.Pipeline()
			incr := pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, window+time.Second)
			_, err := pipe.Exec(ctx)

			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			count, err := incr.Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count > int64(limit) {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				http.Error(w, "Rate limit exceeded. Too many requests.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
