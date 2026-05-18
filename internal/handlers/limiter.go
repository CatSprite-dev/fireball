package handlers

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
	rps     int
}

func NewRateLimiter(rps int) *RateLimiter {
	rl := RateLimiter{
		clients: make(map[string]*client),
		rps:     rps,
	}
	go rl.reap(3 * time.Minute)

	return &rl
}

func (rl *RateLimiter) reap(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, v := range rl.clients {
			if now.After(v.lastSeen.Add(interval)) {
				delete(rl.clients, k)
			}
		}
		rl.mu.Unlock()
	}
}

func getClientIP(r *http.Request) (string, error) {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0], nil
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri, nil
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, err := getClientIP(r)
		if err != nil {
			return
		}

		defer rl.mu.Unlock()
		rl.mu.Lock()
		v, ok := rl.clients[ip]
		if !ok {
			v = &client{
				limiter:  rate.NewLimiter(rate.Every(time.Minute), rl.rps),
				lastSeen: time.Now(),
			}
			rl.clients[ip] = v
		} else {
			v.lastSeen = time.Now()
		}

		if !v.limiter.Allow() {
			pkg.RespondWithError(w, http.StatusTooManyRequests, "rate is limited", errors.New("too many requests"))
			return
		}
		next(w, r)
	}
}
