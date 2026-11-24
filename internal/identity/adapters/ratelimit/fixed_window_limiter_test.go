package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func newTestLimiter(maxAttempts int, period time.Duration) (*FixedWindowLimiter, *fakeClock) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	return newFixedWindowLimiter(maxAttempts, period, clock.Now), clock
}

func TestFixedWindowLimiter_AllowsUpToMaxAttemptsThenRejects(t *testing.T) {
	limiter, clock := newTestLimiter(3, time.Minute)
	ctx := context.Background()

	for i := range 3 {
		allowed, retryAfter := limiter.Allow(ctx, "user@example.com")
		assert.True(t, allowed, "attempt %d", i+1)
		assert.Zero(t, retryAfter)
	}

	clock.Advance(20 * time.Second)
	allowed, retryAfter := limiter.Allow(ctx, "user@example.com")
	assert.False(t, allowed)
	assert.Equal(t, 40*time.Second, retryAfter)
}

func TestFixedWindowLimiter_ResetsAfterWindow(t *testing.T) {
	limiter, clock := newTestLimiter(1, time.Minute)
	ctx := context.Background()

	allowed, _ := limiter.Allow(ctx, "user@example.com")
	assert.True(t, allowed)
	allowed, _ = limiter.Allow(ctx, "user@example.com")
	assert.False(t, allowed)

	clock.Advance(time.Minute)
	allowed, retryAfter := limiter.Allow(ctx, "user@example.com")
	assert.True(t, allowed)
	assert.Zero(t, retryAfter)
}

func TestFixedWindowLimiter_KeysAreIndependent(t *testing.T) {
	limiter, _ := newTestLimiter(1, time.Minute)
	ctx := context.Background()

	allowed, _ := limiter.Allow(ctx, "a@example.com")
	assert.True(t, allowed)
	allowed, _ = limiter.Allow(ctx, "a@example.com")
	assert.False(t, allowed)

	allowed, _ = limiter.Allow(ctx, "b@example.com")
	assert.True(t, allowed)
}

func TestFixedWindowLimiter_SweepsExpiredWindows(t *testing.T) {
	limiter, clock := newTestLimiter(1, time.Minute)
	ctx := context.Background()

	limiter.Allow(ctx, "a@example.com")
	limiter.Allow(ctx, "b@example.com")
	assert.Len(t, limiter.windows, 2)

	clock.Advance(time.Minute)
	limiter.Allow(ctx, "c@example.com")
	assert.Len(t, limiter.windows, 1)
	assert.Contains(t, limiter.windows, "c@example.com")
}

func TestFixedWindowLimiter_ConcurrentAttemptsRespectLimit(t *testing.T) {
	limiter, _ := newTestLimiter(5, time.Minute)
	ctx := context.Background()

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allowed int
	)
	for range 50 {
		wg.Go(func() {
			if ok, _ := limiter.Allow(ctx, "user@example.com"); ok {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	assert.Equal(t, 5, allowed)
}

func TestNewFixedWindowLimiter_DisabledWhenNotPositive(t *testing.T) {
	for _, tc := range []struct {
		name        string
		maxAttempts int
		period      time.Duration
	}{
		{name: "zero attempts", maxAttempts: 0, period: time.Minute},
		{name: "zero period", maxAttempts: 5, period: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewFixedWindowLimiter(tc.maxAttempts, tc.period)
			for range 100 {
				allowed, retryAfter := limiter.Allow(context.Background(), "user@example.com")
				assert.True(t, allowed)
				assert.Zero(t, retryAfter)
			}
		})
	}
}
