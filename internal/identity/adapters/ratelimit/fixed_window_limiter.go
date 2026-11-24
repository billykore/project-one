// Package ratelimit provides in-process login attempt limiters.
package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/billykore/project-one/internal/identity/ports"
)

// window tracks the attempts recorded for one key since start.
type window struct {
	start time.Time
	count int
}

// FixedWindowLimiter allows up to maxAttempts per key in each fixed window.
// State is held in process memory, so every API instance enforces its own
// limit and the counters reset when the process restarts.
type FixedWindowLimiter struct {
	mu          sync.Mutex
	maxAttempts int
	period      time.Duration
	now         func() time.Time
	windows     map[string]*window
	lastSweep   time.Time
}

// NewFixedWindowLimiter returns a limiter that accepts maxAttempts per key in
// every period. A non-positive maxAttempts or period disables limiting.
func NewFixedWindowLimiter(maxAttempts int, period time.Duration) ports.LoginAttemptLimiter {
	if maxAttempts <= 0 || period <= 0 {
		return unlimited{}
	}
	return newFixedWindowLimiter(maxAttempts, period, time.Now)
}

func newFixedWindowLimiter(maxAttempts int, period time.Duration, now func() time.Time) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		maxAttempts: maxAttempts,
		period:      period,
		now:         now,
		windows:     make(map[string]*window),
		lastSweep:   now(),
	}
}

// Allow records an attempt for key. It rejects the attempt once the key has
// used maxAttempts in its current window and reports the time until that
// window ends.
func (l *FixedWindowLimiter) Allow(_ context.Context, key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	w, ok := l.windows[key]
	if !ok || !now.Before(w.start.Add(l.period)) {
		l.windows[key] = &window{start: now, count: 1}
		return true, 0
	}
	if w.count < l.maxAttempts {
		w.count++
		return true, 0
	}
	return false, w.start.Add(l.period).Sub(now)
}

// sweep drops expired windows at most once per period so memory stays
// bounded by the keys seen in the last two periods.
func (l *FixedWindowLimiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.period {
		return
	}
	for key, w := range l.windows {
		if !now.Before(w.start.Add(l.period)) {
			delete(l.windows, key)
		}
	}
	l.lastSweep = now
}

// unlimited accepts every attempt; it is used when limiting is disabled.
type unlimited struct{}

func (unlimited) Allow(context.Context, string) (bool, time.Duration) {
	return true, 0
}
