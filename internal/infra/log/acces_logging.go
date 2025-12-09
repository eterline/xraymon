package log

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"log/slog"

	"github.com/eterline/tunbee/internal/domain"
)

type AccessLogger struct {
	mu     sync.Mutex
	path   string
	logger *slog.Logger
	file   *os.File
}

func NewAccessLogger(path string) (*AccessLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &AccessLogger{
		path:   path,
		file:   file,
		logger: slog.New(handler),
	}, nil
}

func (l *AccessLogger) Write(p []byte) (int, error) {
	line := string(p)

	af, ok := parseAccess(line)
	if !ok {
		return len(p), nil
	}

	target, proto := af.accessProto()

	meta := domain.ConnectionMetadata{
		Client:   af.Client,
		Server:   target,
		Proto:    proto,
		Inbound:  af.Inbound,
		Outbound: af.Outbound,
		User:     af.getEmail(),
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.Info("access",
		slog.String("client", meta.Client),
		slog.String("server", meta.Server),
		slog.String("proto", meta.Proto),
		slog.String("inbound", meta.Inbound),
		slog.String("outbound", meta.Outbound),
		slog.String("user", meta.User),
	)

	return len(p), nil
}

func (l *AccessLogger) Rotate() error {
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

func (l *AccessLogger) Last(n int) ([]domain.ConnectionMetadata, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.Open(l.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var list []domain.ConnectionMetadata

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := scanner.Bytes()

		var obj map[string]any
		if json.Unmarshal(raw, &obj) != nil {
			continue
		}

		if obj["msg"] != "access" {
			continue
		}

		m := domain.ConnectionMetadata{
			Client:   getString(obj["client"]),
			Server:   getString(obj["server"]),
			Proto:    getString(obj["proto"]),
			Inbound:  getString(obj["inbound"]),
			Outbound: getString(obj["outbound"]),
			User:     getString(obj["user"]),
		}

		list = append(list, m)
	}

	if n == 0 || n >= len(list) {
		return list, nil
	}
	return list[len(list)-n:], nil
}

func getString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
