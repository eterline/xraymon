// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package xraycommon

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/eterline/xraymon/internal/domain"
)

func xrayCore() string {
	return filepath.Join(
		"cores",
		fmt.Sprintf("Xray-%s-%s", runtime.GOOS, runtime.GOARCH),
	)
}

type XrayDispatcher struct {
	bin          string
	acceptStream io.Writer
	errorStream  io.Writer
}

func NewXrayDispatcher(accept, err io.Writer) *XrayDispatcher {
	return &XrayDispatcher{
		bin:          xrayCore(),
		acceptStream: accept,
		errorStream:  err,
	}
}

func (xd *XrayDispatcher) Run(ctx context.Context, conf domain.CoreConfiguration, level string) error {

	conf["log"] = structToRawJSON(initLogging(level))
	conf["stats"] = structToRawJSON(initStats())
	conf["api"] = structToRawJSON(initApiObject())

	cmd := exec.CommandContext(ctx, xd.bin)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return err
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return err
	}

	if err := json.NewEncoder(stdin).Encode(conf); err != nil {
		stdin.Close()
		return err
	}
	stdin.Close()

	lines := make(chan []byte)
	go streamLines(stdout, lines)

	for {
		select {
		case <-ctx.Done():
			return cmd.Process.Kill()
		case line, ok := <-lines:
			if !ok {
				return cmd.Wait()
			}
			if err := xd.streamLog(line); err != nil {
				return err
			}
		}
	}
}

// ----------------- Helpers -----------------

func streamLines(r io.Reader, out chan<- []byte) {
	sc := bufio.NewScanner(r)
	defer close(out)

	for sc.Scan() {
		b := append([]byte(nil), sc.Bytes()...) // copy
		out <- b
	}
}

func (xd *XrayDispatcher) streamLog(p []byte) error {
	if bytes.Contains(p, []byte("accepted")) {
		_, err := xd.acceptStream.Write(p)
		return err
	}
	_, err := xd.errorStream.Write(p)
	return err
}

func structToRawJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(b)
}
