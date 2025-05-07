package circuit_test

import (
	"testing"
	"time"

	"github.com/estavadormir/adaptlimit/circuit"
)

func TestCircuitBreakerStates(t *testing.T) {
	breaker := circuit.NewBreaker(2, time.Millisecond*100)

	if breaker.State() != circuit.StateClosed {
		t.Errorf("Initial state should be CLOSED, got %v", breaker.State())
	}

	breaker.Failure()
	if breaker.State() != circuit.StateClosed {
		t.Errorf("State after 1 failure should still be CLOSED, got %v", breaker.State())
	}

	breaker.Failure()
	if breaker.State() != circuit.StateOpen {
		t.Errorf("State after 2 failures should be OPEN, got %v", breaker.State())
	}

	time.Sleep(time.Millisecond * 150)

	if !breaker.Allow() {
		t.Errorf("First request after timeout should be allowed")
	}

	breaker.Success()
	if breaker.State() != circuit.StateClosed {
		t.Errorf("State after success in half-open should be CLOSED, got %v", breaker.State())
	}
}
