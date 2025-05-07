package circuit

import (
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Breaker struct {
	failureThreshold int
	resetTimeout     time.Duration
	halfOpenMax      int

	failures        int
	state           State
	lastStateChange time.Time
	halfOpenCount   int

	mu sync.RWMutex
}

func NewBreaker(failureThreshold int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		halfOpenMax:      1,
		state:            StateClosed,
		lastStateChange:  time.Now(),
	}
}

func (b *Breaker) Allow() bool {
	b.mu.RLock()

	now := time.Now()

	switch b.state {
	case StateClosed:
		b.mu.RUnlock()
		return true
	case StateOpen:
		if now.Sub(b.lastStateChange) > b.resetTimeout {
			b.mu.RUnlock()
			b.mu.Lock()

			if b.state == StateOpen && time.Now().Sub(b.lastStateChange) > b.resetTimeout {
				b.setState(StateHalfOpen)
				b.halfOpenCount = 0
				b.mu.Unlock()
				return true
			}

			result := b.state == StateClosed
			b.mu.Unlock()
			return result
		}
		b.mu.RUnlock()
		return false
	case StateHalfOpen:
		allowed := b.halfOpenCount < b.halfOpenMax
		if allowed {
			b.mu.RUnlock()
			b.mu.Lock()
			if b.state == StateHalfOpen && b.halfOpenCount < b.halfOpenMax {
				b.halfOpenCount++
				b.mu.Unlock()
				return true
			}
			b.mu.Unlock()
			return false
		}
		b.mu.RUnlock()
		return false
	default:
		b.mu.RUnlock()
		return false
	}
}

func (b *Breaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		b.failures = 0
	case StateHalfOpen:
		b.setState(StateClosed)
		b.failures = 0
	}
}

func (b *Breaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		b.failures++
		if b.failures >= b.failureThreshold {
			b.setState(StateOpen)
		}
	case StateHalfOpen:
		b.setState(StateOpen)
	case StateOpen:
		b.lastStateChange = time.Now()
	}
}

func (b *Breaker) State() State {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

func (b *Breaker) setState(state State) {
	b.state = state
	b.lastStateChange = time.Now()
}

func (b *Breaker) WithHalfOpenMax(max int) *Breaker {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.halfOpenMax = max
	return b
}

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}
