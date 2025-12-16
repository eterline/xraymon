// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.
package main

import (
	"github.com/eterline/xraymon/internal/app/xraymon"
	"github.com/eterline/xraymon/internal/config"
	"github.com/eterline/xraymon/internal/infra/log"
	"github.com/eterline/xraymon/pkg/toolkit"
)

// -ladflags variables
var (
	CommitHash = "dev"
	Version    = "dev"
)

var (
	Flags = xraymon.InitFlags{
		CommitHash: CommitHash,
		Version:    Version,
		Repository: "github.com/eterline/xraymon",
	}

	Config = config.Configuration{
		Log: config.Log{
			LogLevel: "info",
			JSONlog:  false,
		},
		Server: config.Server{
			Listen:     "0.0.0.0:3000",
			CrtFileSSL: "",
			KeyFileSSL: "",
		},
		Core: config.Core{
			CoreAccess: "core_access.log",
			CoreLog:    "core_logging.log",
			ConfigFile: "settings.json",
		},
	}
)

func main() {
	root := toolkit.InitAppStart(
		func() error {
			return config.ParseArgs(&Config)
		},
	)

	logger := log.NewLogger(Config.LogLevel, Config.JSONlog)
	root.Context = log.WrapLoggerToContext(root.Context, logger)

	xraymon.Execute(root, Flags, Config)
}
