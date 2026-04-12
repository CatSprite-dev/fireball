package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CatSprite-dev/fireball/internal/handlers"
)

func TestRateLimiter(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/login", nil)

	dummy := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	testLimiter := handlers.NewRateLimiter(2)
	next := http.HandlerFunc(testLimiter.Middleware(dummy))

	next(w, r)
	respCode := w.Result().StatusCode
	if respCode != 200 {
		t.Errorf("unexpected error: %v", respCode)
	}

	w = httptest.NewRecorder()
	next(w, r)
	respCode = w.Result().StatusCode
	if respCode != 200 {
		t.Errorf("unexpected error: %v", respCode)
	}

	w = httptest.NewRecorder()
	next(w, r)
	respCode = w.Result().StatusCode
	if respCode != 429 {
		t.Errorf("expected error, got: %v", respCode)
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	w := httptest.NewRecorder()
	r1 := httptest.NewRequest("POST", "/api/login", nil)
	r1.RemoteAddr = "192.0.2.2:1234"

	r2 := httptest.NewRequest("POST", "/api/login", nil)
	r2.RemoteAddr = "234.1.0.0:1234"

	dummy := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	testLimiter := handlers.NewRateLimiter(2)
	next := http.HandlerFunc(testLimiter.Middleware(dummy))

	next(w, r1)
	respCode := w.Result().StatusCode
	if respCode != 200 {
		t.Errorf("unexpected error: %v", respCode)
	}

	w = httptest.NewRecorder()
	next(w, r1)
	respCode = w.Result().StatusCode
	if respCode != 200 {
		t.Errorf("unexpected error: %v", respCode)
	}

	w = httptest.NewRecorder()
	next(w, r1)
	respCode = w.Result().StatusCode
	if respCode != 429 {
		t.Errorf("expected error, got: %v", respCode)
	}

	w = httptest.NewRecorder()
	next(w, r2)
	respCode = w.Result().StatusCode
	if respCode != 200 {
		t.Errorf("unexpected error: %v", respCode)
	}
}
