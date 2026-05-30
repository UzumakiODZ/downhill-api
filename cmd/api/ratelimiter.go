package cmd

import (
	"net/http"
	"strconv"
	"time"
	"github.com/go-redis/redis_rate/v10"
)

func RateLimit(limiter *redis_rate.Limiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            
			userIP := GetRemoteAddr(r)
            res, err := limiter.Allow(r.Context(), userIP, redis_rate.PerMinute(10))
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            h := w.Header()
            h.Set("RateLimit-Remaining", strconv.Itoa(res.Remaining))

            if res.Allowed == 0 {
                seconds := int(res.RetryAfter / time.Second)
                h.Set("RateLimit-RetryAfter", strconv.Itoa(seconds))

                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}