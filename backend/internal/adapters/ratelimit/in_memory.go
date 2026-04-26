package ratelimit

import (
	"context"
	"sync"
	"time"
)

// InMemory is a simple per-key token bucket. Suitable for single-instance
// deployments. For multi-instance, swap with a Redis-backed adapter.
type InMemory struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity int
	window   time.Duration
	now      func() time.Time
}

type bucket struct {
	tokens     int
	lastRefill time.Time
}

// New returns a limiter that allows `capacity` actions per `window`,
// independently per key. capacity must be >= 1, window > 0.
func New(capacity int, window time.Duration) *InMemory {
	return &InMemory{
		buckets:  make(map[string]*bucket),
		capacity: capacity,
		window:   window,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (l *InMemory) Allow(_ context.Context, key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.capacity, lastRefill: now}
		l.buckets[key] = b
	}

	// Refill: 1 token per (window/capacity) elapsed.
	tokensPerSec := float64(l.capacity) / l.window.Seconds()
	elapsed := now.Sub(b.lastRefill).Seconds()
	add := int(elapsed * tokensPerSec)
	if add > 0 {
		b.tokens += add
		if b.tokens > l.capacity {
			b.tokens = l.capacity
		}
		b.lastRefill = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true, 0
	}
	// retryAfter = time needed to earn one token.
	retry := time.Duration(float64(time.Second) / tokensPerSec)
	return false, retry
}
