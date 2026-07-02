package main

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type bucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func newBucket(maxTokens, refillRate float64) *bucket {
	return &bucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (b *bucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RateLimiter tracks one bucket per IP per route.
// A background goroutine evicts buckets idle for >10 minutes to prevent unbounded memory growth.
type RateLimiter struct {
	buckets map[string]*bucket
	mu      sync.Mutex
	max     float64
	refill  float64
}

func NewRateLimiter(maxTokens, refillPerSecond float64) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*bucket),
		max:     maxTokens,
		refill:  refillPerSecond,
	}
	go rl.sweepStale()
	return rl
}

// sweepStale runs every 5 minutes and removes bucket entries that have had no
// activity for more than 10 minutes, preventing unbounded memory growth.
func (rl *RateLimiter) sweepStale() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		rl.mu.Lock()
		for key, b := range rl.buckets {
			b.mu.Lock()
			if b.lastRefill.Before(cutoff) {
				delete(rl.buckets, key)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip, route string) bool {
	key := ip + ":" + route
	rl.mu.Lock()
	b, exists := rl.buckets[key]
	if !exists {
		b = newBucket(rl.max, rl.refill)
		rl.buckets[key] = b
	}
	rl.mu.Unlock()
	return b.allow()
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		// strip port number so all requests from same machine share a bucket
		if i := strings.LastIndex(ip, ":"); i != -1 {
			ip = ip[:i]
		}
		if !rl.Allow(ip, r.URL.Path) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
