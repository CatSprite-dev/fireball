package tinvest

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLimiterUnknownMethod(t *testing.T) {
	l := NewLimiter()
	err := l.Wait(context.Background(), "NoSuchService/NoSuchMethod")
	if !errors.Is(err, ErrUnknownMethod) {
		t.Fatalf("want ErrUnknownMethod, got %v", err)
	}
}

func TestLimiterCanceledContext(t *testing.T) {
	l := NewLimiter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := l.Wait(ctx, "UsersService/GetAccounts")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

func TestEveryMethodHasLimit(t *testing.T) {
	for method, g := range methodGroup {
		if _, ok := groupLimit[g]; !ok {
			t.Errorf("method %s: group %q has no limit", method, g)
		}
	}
}

func TestLimiterWaitsWhenExhausted(t *testing.T) {
	l := NewLimiter()
	ctx := context.Background()

	for i := 0; i < limitAccounts; i++ {
		if err := l.Wait(ctx, "UsersService/GetAccounts"); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}

	start := time.Now()
	if err := l.Wait(ctx, "UsersService/GetAccounts"); err != nil {
		t.Fatal(err)
	}
	if waited := time.Since(start); waited < time.Second {
		t.Fatalf("waited %v, want about 1.2s", waited)
	}
}
