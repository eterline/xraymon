// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package domain

import (
	"encoding/json"
	"time"
)

type CoreConfiguration map[string]json.RawMessage

type ConfigLoader interface {
	LoadConfig() (CoreConfiguration, error)
}

type ConfigSaver interface {
	SaveConfig(CoreConfiguration) error
}

type CoreStatus struct {
	Working     bool
	LastLog     string
	WorkingTime time.Duration
}
