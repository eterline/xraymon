// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.
package xraycommon

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/eterline/xraymon/internal/domain"
)

var (
	accessReg = regexp.MustCompile(
		`from (?P<client>[^ ]+)\s+accepted\s+(?P<target>[^ ]+)\s+\` +
			`[(?P<inbound>[^ ]+)\s*(?:>>|=>|->)\s*(?P<outbound>[^ ]+)\]` +
			`(?:\s+email:\s*(?P<email>[A-Za-z0-9._\-]+))?`,
	)

	coreLineReg = regexp.MustCompile(`^\S+\s+\S+\s+\[([A-Za-z]+)\]\s+(.+)$`)
)

// ========================

type basicLogger struct {
	logger *slog.Logger

	fileMu sync.Mutex
	path   string
	file   *os.File
}

func (bl *basicLogger) Close() error {
	if bl.file != nil {
		return bl.file.Close()
	}
	return nil
}

func newBasicLogger(path string) (*basicLogger, error) {
	bl := &basicLogger{
		path: path,
	}

	if err := bl.newLog(false); err != nil {
		return nil, err
	}

	return bl, nil
}

func (bl *basicLogger) newLog(rotate bool) error {
	bl.fileMu.Lock()
	defer bl.fileMu.Unlock()

	if bl.file != nil {
		bl.file.Close()
	}

	flag := os.O_CREATE | os.O_WRONLY
	if rotate {
		flag |= os.O_TRUNC // rewrite file
	} else {
		flag |= os.O_APPEND // append file
	}

	f, err := os.OpenFile(bl.path, flag, 0644)
	if err != nil {
		return err
	}

	handler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	bl.file = f
	bl.logger = slog.New(handler)

	return nil
}

func (bl *basicLogger) Rotate() error {
	return bl.newLog(true)
}

// ========================

type coreLogger struct {
	*basicLogger
	lastLine   *bytes.Buffer
	lastLineMu sync.RWMutex
}

func NewCoreLogger(path string) (*coreLogger, error) {
	base, err := newBasicLogger(path)
	if err != nil {
		return nil, err
	}

	cl := &coreLogger{
		basicLogger: base,
		lastLine:    &bytes.Buffer{},
	}

	return cl, nil
}

func (cl *coreLogger) updateLast(line []byte) {
	if len(line) == 0 {
		return
	}

	cl.lastLineMu.Lock()
	defer cl.lastLineMu.Unlock()

	cl.lastLine.Reset()
	cl.lastLine.Write(line)
}

func (cl *coreLogger) Write(p []byte) (int, error) {
	cl.updateLast(p)

	log, ok := parseCoreLine(p)
	if !ok {
		return len(p), nil
	}

	cl.logger.Info(
		"core log",
		"core_level", log.Level,
		"data", log.Payload,
	)

	return len(p), nil
}

func (cl *coreLogger) LastLog() string {
	cl.lastLineMu.RLock()
	defer cl.lastLineMu.RUnlock()
	return cl.lastLine.String()
}

// ========================

type accessLogger struct {
	*basicLogger
}

func NewAccessLogger(path string) (*accessLogger, error) {
	base, err := newBasicLogger(path)
	if err != nil {
		return nil, err
	}

	cl := &accessLogger{
		basicLogger: base,
	}

	return cl, nil
}

func (al *accessLogger) Write(p []byte) (int, error) {
	log, ok := parseAccess(p)
	if !ok {
		return len(p), nil
	}

	al.logger.Info(
		"new connection",
		"client", log.Client,
		"target", log.Target,
		"proto", log.getProto(),
		"inbound", log.Inbound,
		"outbound", log.Outbound,
		"user", log.getEmail(),
	)

	return len(p), nil
}

func (l *accessLogger) readAllConnections() ([]domain.ConnectionMetadata, error) {
	l.fileMu.Lock()
	defer l.fileMu.Unlock()

	f, err := os.Open(l.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var list []domain.ConnectionMetadata

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := scanner.Bytes()

		var obj map[string]string
		if json.Unmarshal(raw, &obj) != nil {
			continue
		}

		m := domain.ConnectionMetadata{
			Client:   obj["client"],
			Server:   obj["target"],
			Proto:    obj["proto"],
			Inbound:  obj["inbound"],
			Outbound: obj["outbound"],
			User:     obj["user"],
		}

		list = append(list, m)
	}

	return list, nil
}

func (al *accessLogger) LastConnections(ctx context.Context, n int) ([]domain.ConnectionMetadata, error) {
	if n <= 0 {
		return al.readAllConnections()
	}

	al.fileMu.Lock()
	defer al.fileMu.Unlock()

	f, err := os.Open(al.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	if size == 0 {
		return nil, nil
	}

	const bufSize = 64 * 1024
	buf := make([]byte, bufSize)

	var (
		lines    = make([][]byte, 0, n)
		leftover []byte
		pos      = size
	)

	for pos > 0 && len(lines) < n {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		readSize := int64(bufSize)
		if pos < readSize {
			readSize = pos
		}
		pos -= readSize

		_, err := f.ReadAt(buf[:readSize], pos)
		if err != nil && err != io.EOF {
			return nil, err
		}

		chunk := buf[:readSize]

		for i := len(chunk) - 1; i >= 0 && len(lines) < n; i-- {
			if chunk[i] == '\n' {
				line := append([]byte{}, chunk[i+1:]...)
				line = append(line, leftover...)

				if len(line) > 0 && json.Valid(line) {
					lines = append(lines, bytes.TrimSpace(line))
				}

				leftover = leftover[:0]
				chunk = chunk[:i]
			}
		}

		leftover = append(chunk, leftover...)
	}

	if len(leftover) > 0 && len(lines) < n && json.Valid(leftover) {
		lines = append(lines, bytes.TrimSpace(leftover))
	}

	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		lines[i], lines[j] = lines[j], lines[i]
	}

	result := make([]domain.ConnectionMetadata, 0, len(lines))
	for _, raw := range lines {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var obj map[string]string
		if err := json.Unmarshal(raw, &obj); err != nil {
			continue
		}

		result = append(result, domain.ConnectionMetadata{
			Client:   obj["client"],
			Server:   obj["target"],
			Proto:    obj["proto"],
			Inbound:  obj["inbound"],
			Outbound: obj["outbound"],
			User:     obj["user"],
		})
	}

	return result, nil
}

// ========================

type coreLineFields struct {
	Level   string
	Payload string
}

func parseCoreLine(line []byte) (meta coreLineFields, ok bool) {
	match := coreLineReg.FindSubmatch(line)
	if len(match) != 3 {
		return
	}

	meta.Level = string(match[1])
	meta.Payload = string(match[2])

	return meta, true
}

// ======

type accessFields struct {
	Client   string
	Target   string
	Inbound  string
	Outbound string
	Email    string
}

func (af *accessFields) getProto() (proto string) {
	switch {
	case strings.HasPrefix(af.Target, "udp"):
		return "udp"
	case strings.HasPrefix(af.Target, "tcp"):
		return "tcp"
	default:
		return "http"
	}
}

func (af *accessFields) getEmail() string {
	if af.Email == "" {
		return "UNKNOWN"
	}
	return af.Email
}

func parseAccess(line []byte) (meta accessFields, ok bool) {
	match := accessReg.FindSubmatch(line)
	if len(match) < 5 {
		return
	}

	meta.Client = string(match[1])
	meta.Target = string(match[2])
	meta.Inbound = string(match[3])
	meta.Outbound = string(match[4])

	if len(match) == 6 {
		meta.Email = string(match[5])
	}

	return meta, true
}
