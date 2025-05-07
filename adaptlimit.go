package adaptlimit

import (
	"context"
	"sync"
	"time"

	"github.com/estavadormir/adaptlimit/config"
	"github.com/estavadormir/adaptlimit/metrics"
)

type AdaptLimiter interface {
	Allow(key string) bool

	Wait(ctx context.Context, key string) error

	Done(key string, success bool, responseTime time.Duration)

	Close() error
}

func New(cfg *config.Config) AdaptLimiter {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	l := &limiter{
		config:         cfg,
		metrics:        metrics.NewCollector(cfg.MetricsInterval),
		limits:         make(map[string]*keyLimit),
		adjustInterval: cfg.AdjustInterval,
	}

	l.startAdjuster()

	return l
}

type limiter struct {
	config         *config.Config
	metrics        *metrics.Collector
	limits         map[string]*keyLimit
	adjustInterval time.Duration
	mu             sync.RWMutex
	closed         bool
	adjusterDone   chan struct{}
}

type keyLimit struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefill     time.Time
	successCount   int64
	failureCount   int64
	responseTimeMs int64
	requestCount   int64
	mu             sync.Mutex
}

func (l *limiter) Allow(key string) bool {
	if l.closed {
		return false
	}

	limit := l.getOrCreateLimit(key)

	limit.mu.Lock()
	defer limit.mu.Unlock()

	l.refillTokens(limit)

	if limit.tokens >= 1 {
		limit.tokens--
		limit.requestCount++
		return true
	}

	return false
}

func (l *limiter) Wait(ctx context.Context, key string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if l.Allow(key) {
				return nil
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (l *limiter) Done(key string, success bool, responseTime time.Duration) {
	if l.closed {
		return
	}

	limit := l.getOrCreateLimit(key)

	limit.mu.Lock()
	defer limit.mu.Unlock()

	if success {
		limit.successCount++
	} else {
		limit.failureCount++
	}

	limit.responseTimeMs += responseTime.Milliseconds()
}

func (l *limiter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	l.closed = true
	close(l.adjusterDone)
	return nil
}

func (l *limiter) getOrCreateLimit(key string) *keyLimit {
	l.mu.RLock()
	limit, ok := l.limits[key]
	l.mu.RUnlock()

	if ok {
		return limit
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	limit, ok = l.limits[key]
	if ok {
		return limit
	}

	limit = &keyLimit{
		tokens:     float64(l.config.InitialLimit),
		maxTokens:  float64(l.config.InitialLimit),
		refillRate: float64(l.config.InitialLimit) / float64(l.config.Interval.Seconds()),
		lastRefill: time.Now(),
	}

	l.limits[key] = limit
	return limit
}

func (l *limiter) refillTokens(limit *keyLimit) {
	now := time.Now()
	elapsed := now.Sub(limit.lastRefill).Seconds()
	limit.lastRefill = now

	newTokens := limit.refillRate * elapsed
	limit.tokens = min(limit.tokens+newTokens, limit.maxTokens)
}

func (l *limiter) startAdjuster() {
	l.adjusterDone = make(chan struct{})

	go func() {
		ticker := time.NewTicker(l.adjustInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				l.adjustLimits()
			case <-l.adjusterDone:
				return
			}
		}
	}()
}

func (l *limiter) adjustLimits() {
	cpuLoad := l.metrics.CPULoad()
	memLoad := l.metrics.MemoryLoad()

	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, limit := range l.limits {
		limit.mu.Lock()

		if limit.requestCount < 10 {
			limit.mu.Unlock()
			continue
		}

		total := limit.successCount + limit.failureCount
		errorRate := 0.0
		if total > 0 {
			errorRate = float64(limit.failureCount) / float64(total)
		}

		avgResponseTime := 0.0
		if total > 0 {
			avgResponseTime = float64(limit.responseTimeMs) / float64(total)
		}

		adjustFactor := 1.0

		if cpuLoad > l.config.HighLoadThreshold {
			adjustFactor *= 0.8 // Reduce by 20% under high CPU load
		} else if cpuLoad < l.config.LowLoadThreshold {
			adjustFactor *= 1.2 // Increase by 20% under low CPU load
		}

		if memLoad > l.config.HighLoadThreshold {
			adjustFactor *= 0.9 // Reduce by 10% under high memory load
		}

		if errorRate > l.config.HighErrorThreshold {
			adjustFactor *= 0.7 // Reduce by 30% under high error rate
		} else if errorRate < l.config.LowErrorThreshold {
			adjustFactor *= 1.1 // Increase by 10% under low error rate
		}

		targetResponseTime := float64(l.config.TargetResponseTime.Milliseconds())
		if avgResponseTime > 0 && targetResponseTime > 0 {
			responseTimeFactor := targetResponseTime / avgResponseTime
			responseTimeFactor = max(0.8, min(responseTimeFactor, 1.2))
			adjustFactor *= responseTimeFactor
		}

		newRefillRate := limit.refillRate * adjustFactor

		minRate := float64(l.config.MinLimit) / float64(l.config.Interval.Seconds())
		maxRate := float64(l.config.MaxLimit) / float64(l.config.Interval.Seconds())
		newRefillRate = max(minRate, min(newRefillRate, maxRate))

		limit.refillRate = newRefillRate
		limit.maxTokens = newRefillRate * float64(l.config.Interval.Seconds())

		limit.successCount = 0
		limit.failureCount = 0
		limit.responseTimeMs = 0
		limit.requestCount = 0

		limit.mu.Unlock()
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
