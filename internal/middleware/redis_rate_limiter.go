package middleware

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redisStore  *redis.Client
	rpm         int
	script      *redis.Script
	behindProxy bool
}

func NewRateLimiter(store *redis.Client, rpm int, behindProxy bool) *RateLimiter {
	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local id = ARGV[4]

		redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
		local count = redis.call('ZCARD', key)
		if count >= limit then
			return 0
		end
		redis.call('ZADD', key, now, id)
		redis.call('EXPIRE', key, math.ceil(window / 1000))
		return 1
	`)
	return &RateLimiter{
		redisStore:  store,
		rpm:         rpm,
		script:      script,
		behindProxy: behindProxy,
	}
}

func getClientIP(r *http.Request, behindProxy bool) (string, error) {
	if behindProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			return strings.TrimSpace(ips[len(ips)-1]), nil
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return xri, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func (rl *RateLimiter) allow(ctx context.Context, ip string) (bool, error) {
	key := "ratelimit:" + ip
	now := time.Now().UnixMilli()
	window := int64(60 * 1000)
	id := uuid.New().String()

	result, err := rl.script.Run(ctx, rl.redisStore, []string{key}, now, window, rl.rpm, id).Result()
	if err != nil {
		return false, err
	}
	return result.(int64) == 1, nil
}

func (rl *RateLimiter) LimiterMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, err := getClientIP(r, rl.behindProxy)
		if err != nil {
			pkg.RespondWithError(w, http.StatusInternalServerError, "couldn't determine ip", err)
			return
		}

		allowed, err := rl.allow(r.Context(), ip)
		if err != nil {
			pkg.RespondWithError(w, http.StatusInternalServerError, "service unavailable", err)
			return
		}
		if !allowed {
			pkg.RespondWithError(w, http.StatusTooManyRequests, "rate is limited", errors.New("too many requests"))
			return
		}
		next(w, r)
	}
}
