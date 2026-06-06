package main

import (
	"net/http"
	"sync"
	"time"
	"strings"
)

type bucket struct {
	tokens   float64
	maxTokens float64
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

// RateLimiter tracks one bucket per IP per route
type RateLimiter struct {
	buckets map[string]*bucket
	mu      sync.Mutex
	max     float64
	refill  float64
}

func NewRateLimiter(maxTokens, refillPerSecond float64) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		max:     maxTokens,
		refill:  refillPerSecond,
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