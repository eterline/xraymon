// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package xraymon

import (
	"fmt"
	"time"

	"github.com/eterline/xraymon/internal/infra/log"
	xrayapi "github.com/eterline/xraymon/internal/infra/xray/api"
	xraycommon "github.com/eterline/xraymon/internal/infra/xray/common"
	"github.com/eterline/xraymon/pkg/toolkit"
)

func Execute(root *toolkit.AppStarter, flags InitFlags) {
	ctx := root.Context
	log := log.MustLoggerFromContext(ctx)

	// ========================================================

	log.Info(
		"starting xraymon",
		"commit", flags.GetCommitHash(),
		"version", flags.GetVersion(),
		"repository", flags.GetRepository(),
	)

	defer func() {
		log.Info(
			"exit from app",
			"running_time", root.WorkTime(),
		)
	}()

	// ========================================================

	f := "./settings.json"
	cfgExporter, err := xraycommon.NewConfigFileProvider(f)
	if err != nil {
		log.Error("failed init config provider", "file", f, "error", err)
		root.MustStopApp(1)
	}

	f = "xray_access.log"
	access, err := xraycommon.NewAccessLogger("xray_access.log")
	if err != nil {
		log.Error("failed init access logger", "file", f, "error", err)
		root.MustStopApp(1)
	}

	f = "xray_core.log"
	core, err := xraycommon.NewCoreLogger("xray_core.log")
	if err != nil {
		log.Error("failed init core logger", "file", f, "error", err)
		root.MustStopApp(1)
	}

	cfg, err := cfgExporter.LoadConfig()
	if err != nil {
		log.Error("failed to load config", "file", f, "error", err)
		root.MustStopApp(1)
	}

	dispatcher := xraycommon.NewXrayDispatcher(access, core)

	root.WrapWorker(func() {
		err := dispatcher.Run(ctx, cfg, "warning")
		if err != nil {
			log.Error("failed to run xray core", "error", err)
		}
	})

	api, err := xrayapi.New("127.0.0.1:3000")
	if err != nil {
		panic(err)
	}

	root.WrapWorker(func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				fmt.Println(api.GetTraffic(ctx, false))
			}
		}
	})

	root.WaitWorkers(10 * time.Second)
}
