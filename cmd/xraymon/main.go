package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/eterline/tunbee/internal/domain"
	"github.com/eterline/tunbee/internal/infra/log"
	"github.com/eterline/tunbee/internal/infra/xraydispatch"
)

func main() {

	cfgData, err := os.ReadFile("./settings.json")
	if err != nil {
		panic(err)
	}

	var cfg json.RawMessage

	err = json.Unmarshal(cfgData, &cfg)
	if err != nil {
		panic(err)
	}

	access, err := log.NewAccessLogger("xray_access.log")
	if err != nil {
		panic(err)
	}

	go func() {
		access.Rotate()

		t := time.NewTimer(30 * time.Second)
		defer t.Stop()

		for range t.C {
			slog.Info("rotate acess log")
			access.Rotate()
		}
	}()

	core, err := log.NewCoreLogger("xray_core.log")
	if err != nil {
		panic(err)
	}

	xray := xraydispatch.NewXrayDispatcher(access, core)
	xray.Run(context.Background(), domain.CoreConfiguration{}, "info")
}
