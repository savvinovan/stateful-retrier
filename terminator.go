package main

import "time"

// Terminator is a struct that encapsulates the conditions under which a task should be terminated.
// Every retryable task should initialize own Terminator.
type Terminator struct {
	maxRetries int
	TTL        time.Duration
	startTime  time.Time
}

func NewTerminator(maxRetries int, ttl time.Duration, startTime time.Time) *Terminator {
	return &Terminator{
		maxRetries: maxRetries,
		TTL:        ttl,
		startTime:  startTime,
	}
}

func (t *Terminator) ShouldTerminate(retryCount int) bool {
	if t.maxRetries > 0 && retryCount >= t.maxRetries {
		return true
	}
	if t.TTL > 0 && time.Since(t.startTime) >= t.TTL {
		return true
	}
	return false
}
