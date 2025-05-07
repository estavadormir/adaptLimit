package adaptlimit

import (
	"testing"
	"time"

	"github.com/estavadormir/adaptlimit/config"
)

func TestRateLimiterBasic(t *testing.T) {
	cfg := config.DefaultConfig().
		WithInitialLimit(10).
		WithMinLimit(5).
		WithMaxLimit(20).
		WithInterval(time.Second)

	limiter := New(cfg)
	defer limiter.Close()

	key := "test-key"

	for i := range 10 {
		if !limiter.Allow(key) {
			t.Errorf("Request %d should be allowed, but was denied", i)
		}
	}

	if limiter.Allow(key) {
		t.Errorf("Request 11 should be denied, but was allowed")
	}

	time.Sleep(time.Second)

	if !limiter.Allow(key) {
		t.Errorf("Request after refill should be allowed, but was denied")
	}
}

func TestRateLimiterWait(t *testing.T) {
	cfg := config.DefaultConfig().
		WithInitialLimit(2).
		WithInterval(time.Millisecond * 50)

	limiter := New(cfg)
	defer limiter.Close()

	key := "test-key-wait"

	if !limiter.Allow(key) {
		t.Fatalf("First request should be allowed")
	}
	if !limiter.Allow(key) {
		t.Fatalf("Second request should be allowed")
	}

	if limiter.Allow(key) {
		t.Errorf("Should not allow request after tokens are depleted")
	}

	time.Sleep(time.Millisecond * 70)

	if !limiter.Allow(key) {
		t.Errorf("Should allow request after waiting for token refill")
	}
}

func TestRateLimiterDone(t *testing.T) {
	cfg := config.DefaultConfig().
		WithInitialLimit(10).
		WithAdjustInterval(time.Millisecond * 100)

	limiter := New(cfg)
	defer limiter.Close()

	key := "test-key-done"

	for range 20 {
		if limiter.Allow(key) {
			limiter.Done(key, false, time.Millisecond*500)
		}
	}

	time.Sleep(time.Millisecond * 200)

	allowed := 0
	for range 20 {
		if limiter.Allow(key) {
			allowed++
		}
	}

	if allowed >= 10 {
		t.Errorf("Limit should have decreased below 10, but allowed %d requests", allowed)
	}
}
