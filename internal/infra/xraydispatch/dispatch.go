package xraydispatch

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

func xrayCore() string {
	return fmt.Sprintf("Xray-%s-%s", runtime.GOOS, runtime.GOARCH)
}

type XrayDispatcher struct {
	bin          string
	acceptStream io.Writer
	errorStream  io.Writer
}

func NewXrayDispatcher() *XrayDispatcher {
	return &XrayDispatcher{
		bin: xrayCore(),
	}
}

func (xd *XrayDispatcher) Run(ctx context.Context, conf json.RawMessage) error {
	var cfg map[string]any
	if err := json.Unmarshal(conf, &cfg); err != nil {
		panic(err)
	}

	cfg["stats"] = initLogging("info")
	cfg["stats"] = initStats()
	cfg["api"] = initApiObject()

	xrayCmd := exec.CommandContext(ctx, xd.bin)

	stdin, err := xrayCmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	defer stdin.Close()

	stdout, err := xrayCmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer stdout.Close()

	err = json.NewEncoder(stdin).Encode(cfg)
	if err != nil {
		panic(err)
	}

	if err := xrayCmd.Start(); err != nil {
		panic(err)
	}

	xrayOut := bufio.NewScanner(stdout)

	scans := make(chan struct{})

	go func() {
		defer close(scans)

		for xrayOut.Scan() {
			scans <- struct{}{}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-scans:
			err := xd.streamLog(xrayOut.Bytes())
			if err != nil {
				return err
			}
		}
	}
}

func (xd *XrayDispatcher) streamLog(p []byte) error {
	if bytes.Contains(p, []byte("accept")) {
		_, err := xd.acceptStream.Write(p)
		return err
	}

	_, err := xd.errorStream.Write(p)
	return err
}
