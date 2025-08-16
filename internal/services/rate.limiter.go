package services

import (
	"SalaryAdvance/pkg/config"
	"sync"
	"time"
)

type LoginRateLimiter struct {
	attempts map[string]*LoginAttempt
	mutex    sync.Mutex
}

type LoginAttempt struct {
	Count       int
	LastAttempt time.Time
}

func NewLoginRateLimiter() *LoginRateLimiter {
	return &LoginRateLimiter{
		attempts: make(map[string]*LoginAttempt),
	}
}

func (r *LoginRateLimiter) CheckAndIncrement(email string) (bool, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	const maxAttempts = 5
	const windowDuration = 15 * time.Minute

	attempt, exists := r.attempts[email]
	currentTime := time.Now()

	if exists && currentTime.Sub(attempt.LastAttempt) > windowDuration {
		attempt.Count = 0
	}

	if !exists {
		attempt = &LoginAttempt{
			Count:       0,
			LastAttempt: currentTime,
		}
		r.attempts[email] = attempt
	}

	if attempt.Count >= maxAttempts {
		return false, config.ErrTooManyRequests
	}

	attempt.Count++
	attempt.LastAttempt = currentTime
	return true, nil
}

func (r *LoginRateLimiter) Reset(email string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.attempts, email)
}
