// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package xraycommon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/eterline/xraymon/internal/domain"
	"github.com/eterline/xraymon/internal/utils/usecase"
)

var allowedFields = usecase.NewWhitelist(
	"log",
	"api",
	"dns",
	"routing",
	"policy",
	"inbounds",
	"outbounds",
	"transport",
	"stats",
	"reverse",
	"fakedns",
	"metrics",
	"observatory",
	"burstObservatory",
)

func clearConfig(ccf *domain.CoreConfiguration) {
	for key := range *ccf {
		if !allowedFields.Allowed(key) {
			delete(*ccf, key)
		}
	}
}

// ==============

type logObject struct {
	Access      string `json:"access"`
	Error       string `json:"error"`
	Loglevel    string `json:"loglevel"`
	DNSLog      bool   `json:"dnsLog"`
	MaskAddress string `json:"maskAddress"`
}

func defineLevel(l string) string {
	levels := []string{"debug", "info", "warning", "error"}
	for _, lv := range levels {
		if lv == l {
			return l
		}
	}
	return "info"
}

func initLogging(level string) *logObject {
	return &logObject{
		Access:      "",
		Error:       "",
		Loglevel:    defineLevel(level),
		DNSLog:      false,
		MaskAddress: "",
	}
}

type statsObject struct{}

func initStats() *statsObject {
	return &statsObject{}
}

type apiObject struct {
	Tag      string   `json:"tag"`
	Listen   string   `json:"listen"`
	Services []string `json:"services"`
}

func initApiObject() *apiObject {
	return &apiObject{
		Tag:    "api",
		Listen: "127.0.0.1:3000",
		Services: []string{
			"HandlerService",
			"LoggerService",
			"StatsService",
			"RoutingService",
		},
	}
}

// ==============

type configFileProvider struct {
	path     string
	confFile *os.File

	mu sync.RWMutex
}

func NewConfigFileProvider(path string) (*configFileProvider, error) {
	if filepath.Base(path) == "config.json" {
		return nil, errors.New("core settings can't have name 'config.json'")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed open config: %w", err)
	}

	cfp := &configFileProvider{
		path:     path,
		confFile: f,
	}

	cfg, err := cfp.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed test config: %w", err)
	}

	err = cfp.SaveConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed test config: %w", err)
	}

	return cfp, nil
}

func (cfp *configFileProvider) LoadConfig() (domain.CoreConfiguration, error) {
	cfp.mu.RLock()
	defer cfp.mu.RUnlock()

	if _, err := cfp.confFile.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek: %w", err)
	}

	cfg := domain.CoreConfiguration{}

	dec := json.NewDecoder(cfp.confFile)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	clearConfig(&cfg)

	return cfg, nil
}

func (cfp *configFileProvider) SaveConfig(cfg domain.CoreConfiguration) error {
	cfp.mu.Lock()
	defer cfp.mu.Unlock()

	tmpPath := cfp.path + ".tmp"

	tmp, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open temp file: %w", err)
	}

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "    ")

	clearConfig(&cfg)

	if err := enc.Encode(cfg); err != nil {
		tmp.Close()
		return fmt.Errorf("encode: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	if err := os.Rename(tmpPath, cfp.path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	if cfp.confFile != nil {
		cfp.confFile.Close()
	}

	f, err := os.Open(cfp.path)
	if err != nil {
		return fmt.Errorf("reopen config: %w", err)
	}
	cfp.confFile = f

	return nil
}

func (cfp *configFileProvider) Close() error {
	cfp.mu.Lock()
	defer cfp.mu.Unlock()

	if cfp.confFile != nil {
		err := cfp.confFile.Close()
		cfp.confFile = nil
		return err
	}

	return nil
}
