package log

import (
	"log/slog"
	"os"
	"sync"
)

type CoreLogger struct {
	mu sync.Mutex

	path   string
	logger *slog.Logger
	file   *os.File

	lastLine []byte
	lineMu   sync.RWMutex
}

func NewCoreLogger(path string) (*CoreLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &CoreLogger{
		path:     path,
		file:     file,
		logger:   slog.New(handler),
		lastLine: nil,
	}, nil
}

func (l *CoreLogger) Write(p []byte) (int, error) {
	l.lineMu.Lock()
	l.lastLine = append(l.lastLine[:0], p...)
	l.lineMu.Unlock()

	c, ok := parseCoreLine(p)
	if !ok {
		return len(p), nil
	}

	log := l.logger.With("payload", c.Payload)

	if c.IsError {
		log.Error("xray error occured")
	} else {
		log.Info("xray new log")
	}

	return len(p), nil
}

func (l *CoreLogger) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	handler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	l.file = f
	l.logger = slog.New(handler)

	return nil
}

func (l *CoreLogger) LastLine() string {
	l.lineMu.RLock()
	defer l.lineMu.RUnlock()
	return string(l.lastLine)
}
