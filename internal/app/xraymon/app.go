// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.
package xraymon

import (
	"log/slog"
	"time"

	"github.com/eterline/xraymon/internal/config"
	"github.com/eterline/xraymon/internal/infra/log"
	xraycommon "github.com/eterline/xraymon/internal/infra/xray/common"
	"github.com/eterline/xraymon/internal/interface/grpc/commands"
	"github.com/eterline/xraymon/internal/interface/grpc/server"
	"github.com/eterline/xraymon/internal/usecase/manager"
	"github.com/eterline/xraymon/internal/usecase/statspool"
	"github.com/eterline/xraymon/pkg/toolkit"
	"google.golang.org/grpc"
)

func Execute(root *toolkit.AppStarter, flags InitFlags, conf config.Configuration) {
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

	log.Info("init base xray settings file", "file", conf.ConfigFile)
	cfgExporter, err := xraycommon.NewConfigFileProvider(conf.ConfigFile)
	if err != nil {
		log.Error("failed init config provider", "file", conf.ConfigFile, "error", err)
		root.MustStopApp(1)
	}
	defer cfgExporter.Close()

	log.Info("init access logger", "file", conf.CoreAccess)
	accessLog, err := xraycommon.NewAccessLogger(conf.CoreAccess)
	if err != nil {
		log.Error("failed init access logger", "file", conf.CoreAccess, "error", err)
		root.MustStopApp(1)
	}
	defer accessLog.Close()

	log.Info("init core logger", "file", conf.CoreLog)
	coreLog, err := xraycommon.NewCoreLogger(conf.CoreLog)
	if err != nil {
		log.Error("failed init core logger", "file", conf.CoreLog, "error", err)
		root.MustStopApp(1)
	}
	defer coreLog.Close()

	// ========================================================

	dsp := xraycommon.NewXrayDispatcher(accessLog, coreLog)
	coreMg := manager.NewCoreManager(ctx, dsp, cfgExporter, coreLog, "warning")

	root.WrapWorker(func() {
		log.Info("starting core")
		err := coreMg.Start()
		if err != nil {
			slog.Error("start core failed", "error", err)
		}
	})

	statProv, err := xraycommon.NewStatsProvider()
	if err != nil {
		log.Error("failed init stats provider", "error", err)
		root.MustStopApp(1)
	}

	statsPool := statspool.NewStatsPool(statProv, 5*time.Second, log)
	statsPool.Start(ctx)
	defer statsPool.Stop()

	// ========================================================

	var grpcSrv *grpc.Server
	if conf.CrtFileSSL != "" && conf.KeyFileSSL != "" {
		log.Info(
			"init grpc with tls",
			"cert", conf.CrtFileSSL,
			"key", conf.KeyFileSSL,
		)

		grpcSrv, err = server.NewTLSGrpcServer(conf.CrtFileSSL, conf.KeyFileSSL)
		if err != nil {
			log.Error("failed init tls grpc server", "error", err)
			root.MustStopApp(1)
		}
	} else {
		grpcSrv = grpc.NewServer()
	}

	// ==========

	coreManage := commands.NewCoreManageHandlers(cfgExporter, cfgExporter, coreMg, log)
	commands.RegisterCoreManagmentServiceServer(grpcSrv, coreManage)

	jrnl := commands.NewJournalHandlers(accessLog, coreLog, statsPool, log)
	commands.RegisterJournalProviderServer(grpcSrv, jrnl)

	// ==========

	srv, err := server.NewGrpcServerWrapper(grpcSrv, conf.Listen)
	if err != nil {
		log.Error("failed init grpc server", "error", err)
		root.MustStopApp(1)
	}

	root.WrapWorker(func() {
		log.Info("starting grpc server", "listen", conf.Listen)
		err := srv.Run(ctx)
		if err != nil {
			slog.Error("start grpc server failed", "error", err)
		}
	})
	defer srv.Close()

	// ========================================================

	root.WaitWorkers(10 * time.Second)
}
