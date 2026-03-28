package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu       sync.Mutex
	lastCall time.Time
	interval time.Duration
}

func New() *Limiter {
	return &Limiter{
		interval: 1 * time.Second,
	}
}

func NewWithInterval(interval time.Duration) *Limiter {
	return &Limiter{
		interval: interval,
	}
}

func (l *Limiter) Wait() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.lastCall.IsZero() {
		elapsed := time.Since(l.lastCall)
		if elapsed < l.interval {
			time.Sleep(l.interval - elapsed)
		}
	}
	l.lastCall = time.Now()
}
