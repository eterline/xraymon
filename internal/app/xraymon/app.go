// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.
package xraymon

import (
	"log/slog"
	"time"

	"github.com/eterline/xraymon/internal/infra/log"
	xraycommon "github.com/eterline/xraymon/internal/infra/xray/common"
	"github.com/eterline/xraymon/internal/usecase/manager"
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
	defer cfgExporter.Close()

	f = "xray_access.log"
	accessLog, err := xraycommon.NewAccessLogger("xray_access.log")
	if err != nil {
		log.Error("failed init access logger", "file", f, "error", err)
		root.MustStopApp(1)
	}
	defer accessLog.Close()

	f = "xray_core.log"
	coreLog, err := xraycommon.NewCoreLogger("xray_core.log")
	if err != nil {
		log.Error("failed init core logger", "file", f, "error", err)
		root.MustStopApp(1)
	}
	defer coreLog.Close()

	// ========================================================

	dsp := xraycommon.NewXrayDispatcher(accessLog, coreLog)
	coreMg := manager.NewCoreManager(ctx, dsp, cfgExporter, coreLog, "warning")

	root.WrapWorker(func() {
		err := coreMg.Start()
		if err != nil {
			slog.Error("start core failed", "error", err)
		}
	})

	// ========================================================

	root.WaitWorkers(10 * time.Second)
}
