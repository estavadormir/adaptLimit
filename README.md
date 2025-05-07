# AdaptLimit 🚦

[![Go Report Card](https://goreportcard.com/badge/github.com/estavadormir/adaptlimit)](https://goreportcard.com/report/github.com/estavadormir/adaptlimit)
[![codecov](https://codecov.io/gh/estavadormir/adaptlimit/branch/main/graph/badge.svg)](https://codecov.io/gh/estavadormir/adaptlimit)
[![Go Reference](https://pkg.go.dev/badge/github.com/estavadormir/adaptlimit.svg)](https://pkg.go.dev/github.com/estavadormir/adaptlimit)
![Test and Coverage](https://github.com/estavadormir/adaptlimit/workflows/Test%20and%20Coverage/badge.svg)

A smarter rate limiter for Go that adapts on its own! Unlike regular rate limiters that use fixed limits, AdaptLimit watches how your system is doing and adjusts automatically.

## What Makes It Cool?

- **It adapts on its own** - Watches for:
  - How busy your server is
  - Whether requests are failing
  - How long responses take

- **Built-in circuit breaker** - Like a safety switch to prevent cascading failures

- **Easy to use** - Simple API that fits right into your Go code

- **Highly configurable** - Tweak it how you want, or just use the defaults

## Getting Started

Install it:

```bash
go get github.com/estavadormir/adaptlimit
```

Basic usage:

```go
package main

import (
	"github.com/estavadormir/adaptlimit"
)

func main() {
	// Create a limiter (it's smart right out of the box)
	limiter := adaptlimit.New(nil)
	defer limiter.Close()

	// Use it to check if a request should proceed
	if limiter.Allow("api-key") {
		// Handle the request
	} else {
		// Tell the client to slow down
	}

	// Let the limiter know how it went
	limiter.Done("api-key", true, time.Millisecond*50)
}
```

## Add It to Your Web Server

```go
// Quick middleware example
func rateLimitMiddleware(limiter adaptlimit.AdaptLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr

			if !limiter.Allow(clientIP) {
				http.Error(w, "Whoa there! Too many requests", http.StatusTooManyRequests)
				return
			}

			start := time.Now()
			success := true

			defer func() {
				limiter.Done(clientIP, success, time.Since(start))
			}()

			next.ServeHTTP(w, r)
		})
	}
}
```

## Customize It

Want to tweak settings? No problem:

```go
cfg := config.DefaultConfig().
	WithInitialLimit(200).                    // Start with this many requests/second
	WithInterval(time.Second).                // Per second rate limiting
	WithTargetResponseTime(time.Millisecond * 100)  // Aim for 100ms responses

limiter := adaptlimit.New(cfg)
```

## How It Works

Think of AdaptLimit like a smart bouncer at a club:

1. Starts with reasonable limits on how many people can enter
2. Watches how crowded it's getting inside
3. Notices if people inside are having a good time
4. Adjusts entry rate based on what it sees

When your server gets busy or slow, it automatically slows down incoming traffic. When things are running smoothly, it lets more through.

## License

MIT License - See LICENSE file for details
