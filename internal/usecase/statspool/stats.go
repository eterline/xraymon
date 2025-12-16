package statspool

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/eterline/xraymon/internal/domain"
)

type StatsProvider interface {
	Stats(ctx context.Context) ([]domain.StatsSnapshot, error)
}

type StatsPool struct {
	statsProvider StatsProvider
	pollInterval  time.Duration
	logger        *slog.Logger

	mu    sync.RWMutex
	cache []domain.StatsSnapshot
	done  context.CancelFunc
}

func NewStatsPool(stats StatsProvider, interval time.Duration, logger *slog.Logger) *StatsPool {
	return &StatsPool{
		statsProvider: stats,
		pollInterval:  interval,
		logger:        logger,
	}
}

func (p *StatsPool) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	p.done = cancel

	go func() {
		ticker := time.NewTicker(p.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				p.logger.Info("StatsPool stopped")
				return
			case <-ticker.C:
				p.collect(ctx)
			}
		}
	}()
}

func (p *StatsPool) Stop() {
	if p.done != nil {
		p.done()
	}
}

func (p *StatsPool) StatsNow(ctx context.Context) ([]domain.StatsSnapshot, error) {
	done := make(chan struct{})
	var snapshot []domain.StatsSnapshot

	go func() {
		p.mu.RLock()
		defer p.mu.RUnlock()
		snapshot = append([]domain.StatsSnapshot(nil), p.cache...)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return snapshot, nil
	}
}

func (p *StatsPool) collect(ctx context.Context) {
	snapshots, err := p.statsProvider.Stats(ctx)
	if err != nil {
		p.logger.Error("failed to collect stats", "error", err)
		return
	}

	p.mu.Lock()
	p.cache = snapshots
	p.mu.Unlock()

	p.logger.Debug("collected stats", "count", len(snapshots))
}
