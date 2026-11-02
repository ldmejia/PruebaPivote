package core

import (
	"sync"
	"time"
)

type Breaker struct {
	mu        sync.Mutex
	threshold int
	cooldown  time.Duration
	failures  int
	openUntil time.Time
	probing   bool
	now       func() time.Time
}

func NewBreaker(threshold int, cooldown time.Duration) *Breaker {
	return &Breaker{threshold: threshold, cooldown: cooldown, now: time.Now}
}

func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.failures < b.threshold {
		return true
	}
	if b.now().Before(b.openUntil) || b.probing {
		return false
	}
	b.probing = true
	return true
}

func (b *Breaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.probing = false
}

func (b *Breaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	b.probing = false
	if b.failures >= b.threshold {
		b.openUntil = b.now().Add(b.cooldown)
	}
}

func (b *Breaker) State() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch {
	case b.failures < b.threshold:
		return "closed"
	case b.now().Before(b.openUntil):
		return "open"
	default:
		return "half-open"
	}
}
