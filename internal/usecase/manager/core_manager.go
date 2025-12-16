package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/eterline/xraymon/internal/domain"
	"github.com/eterline/xraymon/internal/infra/log"
)

type CoreRunner interface {
	Run(ctx context.Context, conf domain.CoreConfiguration, level string) error
}

type LastLogger interface {
	LastLog() string
}

type CoreManager struct {
	rootCtx context.Context

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	loader domain.ConfigLoader
	dsp    CoreRunner
	level  string

	restartCh chan restartType
	closed    bool

	lastLine      LastLogger
	working       bool
	lastStartTime time.Time
	crashRestarts int // число рестартов подряд после краша
}

type restartType int

const (
	restartManual restartType = iota
	restartCrash
)

func NewCoreManager(ctx context.Context, dsp CoreRunner, loader domain.ConfigLoader, last LastLogger, level string) *CoreManager {
	m := &CoreManager{
		dsp:       dsp,
		loader:    loader,
		level:     level,
		rootCtx:   ctx,
		lastLine:  last,
		restartCh: make(chan restartType, 1),
	}

	go m.loop()

	return m
}

func (m *CoreManager) Start() error {
	return m.Restart()
}

func (m *CoreManager) Restart() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("core manager closed")
	}

	select {
	case m.restartCh <- restartManual:
	default:
	}

	return nil
}

func (m *CoreManager) loop() {
	log := log.MustLoggerFromContext(m.rootCtx)

	for {
		select {
		case <-m.rootCtx.Done():
			log.Info("core manager exit")
			if m.cancel != nil {
				m.cancel()
			}
			return

		case t := <-m.restartCh:
			if t == restartCrash {
				delay, t := m.handleCrashDelay()
				log.Error("core crash", "restart_in", int(delay.Seconds()))
				<-t
			}

			m.performRestart(t)
		}
	}
}

func (m *CoreManager) handleCrashDelay() (delay time.Duration, next <-chan time.Time) {
	m.mu.Lock()
	n := m.crashRestarts
	m.mu.Unlock()
	delay = time.Duration(10+5*n) * time.Second
	return delay, time.After(delay)
}

func (m *CoreManager) performRestart(t restartType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}

	if t == restartManual {
		m.crashRestarts = 0
	}

	if m.cancel != nil {
		m.cancel()
	}

	cfg, err := m.loader.LoadConfig()
	if err != nil {
		log.MustLoggerFromContext(m.rootCtx).Error("config load failed", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(m.rootCtx)
	m.ctx = ctx
	m.cancel = cancel

	m.working = true
	m.lastStartTime = time.Now()

	go func() {
		log := log.MustLoggerFromContext(ctx)

		log.Info("run core", "log_level", m.level)

		err := m.dsp.Run(ctx, cfg, m.level)
		if err != nil {
			log.Error("core crashed", "error", err)

			m.mu.Lock()
			m.working = false
			m.crashRestarts++
			m.mu.Unlock()

			select {
			case m.restartCh <- restartCrash:
			default:
			}
			return
		}

		m.mu.Lock()
		m.working = true
		m.crashRestarts = 0
		m.mu.Unlock()
	}()
}

func (m *CoreManager) Status() domain.CoreStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	var wt time.Duration
	if m.working {
		wt = time.Since(m.lastStartTime)
	}

	fmt.Println(wt)

	return domain.CoreStatus{
		Working:     m.working,
		LastLog:     m.lastLine.LastLog(),
		WorkingTime: wt,
	}
}
