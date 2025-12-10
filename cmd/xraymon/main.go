// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package main

import (
	"github.com/eterline/xraymon/internal/app/xraymon"
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
)

func main() {
	root := toolkit.InitAppStart(
		func() error {
			return nil
		},
	)

	logger := log.NewLogger("info", false)
	root.Context = log.WrapLoggerToContext(root.Context, logger)

	xraymon.Execute(root, Flags)
}
