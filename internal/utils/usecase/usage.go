package usecase

import (
	"sync"
	"time"
)

type interavlLim struct {
	mu      sync.Mutex
	inter   time.Duration
	lastuse time.Time
}

func NewIntervalLimiter(interval time.Duration) *interavlLim {
	return &interavlLim{
		inter: interval,
	}
}

func (l *interavlLim) InLimits() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if now.Sub(l.lastuse) < l.inter {
		return false
	}

	l.lastuse = now
	return true
}
