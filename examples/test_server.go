package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/estavadormir/adaptlimit"
	"github.com/estavadormir/adaptlimit/config"
)

var (
	requestsTotal    int64
	requestsAllowed  int64
	requestsRejected int64
)

func main() {
	limiterConfig := config.DefaultConfig().
		WithInitialLimit(10). // Reduce from 100 to 10
		WithMinLimit(5).      // Reduce minimum as well
		WithMaxLimit(50).     // Reduce max from 500 to 50
		WithInterval(time.Second).
		WithAdjustInterval(time.Second * 2).          // More frequent adjustments
		WithTargetResponseTime(time.Millisecond * 50) // More strict target

	log.Printf("Starting with configuration: Initial=%d, Min=%d, Max=%d, Interval=%v, AdjustInterval=%v",
		limiterConfig.InitialLimit, limiterConfig.MinLimit, limiterConfig.MaxLimit,
		limiterConfig.Interval, limiterConfig.AdjustInterval)

	limiter := adaptlimit.New(limiterConfig)
	defer limiter.Close()

	//simulates all traffic coming from one client
	const fixedClientKey = "test-client"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleWithRateLimit(limiter, handleRoot, fixedClientKey, w, r)
	})
	http.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		handleWithRateLimit(limiter, handleSlow, fixedClientKey, w, r)
	})
	http.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		handleWithRateLimit(limiter, handleError, fixedClientKey, w, r)
	})
	http.HandleFunc("/metrics", handleMetrics)

	srv := &http.Server{
		Addr: ":8080",
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	go reportStats()

	log.Println("Starting test server on http://localhost:8080")
	log.Println("Available endpoints:")
	log.Println("  / - Normal endpoint (fast response)")
	log.Println("  /slow - Slow endpoint (simulates heavy processing)")
	log.Println("  /error - Error endpoint (simulates failures)")
	log.Println("  /metrics - Shows current metrics")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func handleWithRateLimit(limiter adaptlimit.AdaptLimiter, handler http.HandlerFunc, clientKey string, w http.ResponseWriter, r *http.Request) {
	//n of requests
	atomic.AddInt64(&requestsTotal, 1)

	if !limiter.Allow(clientKey) {
		atomic.AddInt64(&requestsRejected, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Rate limit exceeded",
		})
		return
	}

	atomic.AddInt64(&requestsAllowed, 1)

	start := time.Now()
	success := true

	defer func() {
		responseTime := time.Since(start)
		limiter.Done(clientKey, success, responseTime)

		log.Printf("Request completed: Success=%v, ResponseTime=%v", success, responseTime)
	}()

	handler(w, r)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	delay := parseDelayQuery(r, 0)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Normal response",
	})
}

func handleSlow(w http.ResponseWriter, r *http.Request) {
	delay := parseDelayQuery(r, 200)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Slow response (delayed %dms)", delay),
	})
}

func handleError(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	failureRate := 100
	failureRateParam := r.URL.Query().Get("rate")
	if failureRateParam != "" {
		if rate, err := strconv.Atoi(failureRateParam); err == nil {
			failureRate = rate
		}
	}

	if failureRate > 0 && (failureRate == 100 || (failureRate > 0 && failureRate > (time.Now().Nanosecond()%100))) {
		panic("simulated error")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Error endpoint returned success (lucky!)",
	})
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	total := atomic.LoadInt64(&requestsTotal)
	allowed := atomic.LoadInt64(&requestsAllowed)
	rejected := atomic.LoadInt64(&requestsRejected)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_requests":    total,
		"allowed_requests":  allowed,
		"rejected_requests": rejected,
		"rejection_rate":    fmt.Sprintf("%.2f%%", float64(rejected)/float64(total)*100),
	})
}

func parseDelayQuery(r *http.Request, defaultDelay int) int {
	delayParam := r.URL.Query().Get("delay")
	if delayParam != "" {
		if delay, err := strconv.Atoi(delayParam); err == nil {
			return delay
		}
	}
	return defaultDelay
}

func reportStats() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for range ticker.C {
		total := atomic.LoadInt64(&requestsTotal)
		allowed := atomic.LoadInt64(&requestsAllowed)
		rejected := atomic.LoadInt64(&requestsRejected)

		var rejectionRate float64
		if total > 0 {
			rejectionRate = float64(rejected) / float64(total) * 100
		}

		log.Printf("Stats: Total=%d, Allowed=%d, Rejected=%d, Rejection Rate=%.2f%%",
			total, allowed, rejected, rejectionRate)
	}
}
