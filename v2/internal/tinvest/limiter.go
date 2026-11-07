package tinvest

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	limiters map[group]*rate.Limiter
}

func NewLimiter() *Limiter {
	limiters := make(map[group]*rate.Limiter, len(groupLimit))
	for g, perMinute := range groupLimit {
		limiters[g] = rate.NewLimiter(rate.Every(time.Minute/time.Duration(perMinute)), perMinute)
	}
	return &Limiter{limiters: limiters}
}

func (l *Limiter) Wait(ctx context.Context, method string) error {
	g, ok := methodGroup[method]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownMethod, method)
	}
	return l.limiters[g].Wait(ctx)
}
