package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Every(time.Minute), rps),
	}
}

func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.limiter.Allow() {
			pkg.RespondWithError(w, http.StatusTooManyRequests, "rate is limited", errors.New("too many requests"))
			return
		}
		next(w, r)
	}
}
