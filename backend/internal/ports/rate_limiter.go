package ports

import (
	"context"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (allowed bool, retryAfter time.Duration)
}
