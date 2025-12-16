// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.
package commands

import (
	context "context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/eterline/xraymon/internal/domain"
	"github.com/eterline/xraymon/internal/utils/usecase"
	grpc "google.golang.org/grpc"
)

// Limiter - interface for call rate limiting.
type Limiter interface {
	// InLimits returns true if the operation is allowed according to the limits.
	InLimits() bool
}

// LastConnProvider - provider of the last connections.
type LastConnProvider interface {
	// LastConnections returns the last N connection metadata records.
	LastConnections(context.Context, int) ([]domain.ConnectionMetadata, error)
}

// coreManageHandlers - gRPC handler for core management operations.
type coreManageHandlers struct {
	confSave  domain.ConfigSaver
	confLoad  domain.ConfigLoader
	coreState domain.CoreState

	confSaveLim    Limiter
	coreRestartLim Limiter

	log *slog.Logger

	UnimplementedCoreManagmentServiceServer
}

// NewCoreManageHandlers - creates a new coreManageHandlers instance with interval limiters.
func NewCoreManageHandlers(
	s domain.ConfigSaver,
	l domain.ConfigLoader,
	r domain.CoreState,
	lc LastConnProvider,
	log *slog.Logger,
) *coreManageHandlers {
	return &coreManageHandlers{
		confSave:  s,
		confLoad:  l,
		coreState: r,

		confSaveLim:    usecase.NewIntervalLimiter(5 * time.Second),
		coreRestartLim: usecase.NewIntervalLimiter(5 * time.Second),

		log: log,
	}
}

// CoreStatus - returns the current core status.
func (cmh *coreManageHandlers) CoreStatus(ctx context.Context, r *CoreStatusRequest) (*CoreStatusResponse, error) {

	status := cmh.coreState.Status()
	cmh.log.Debug("core status requested", "status", status)

	return domain2dtoCoreStatusResponse(status), nil
}

// CoreRestart - triggers a core restart with rate-limiting protection.
func (cmh *coreManageHandlers) CoreRestart(ctx context.Context, r *CoreRestartRequest) (*CoreRestartResponse, error) {

	if !cmh.coreRestartLim.InLimits() {
		cmh.log.Warn("core restart request ignored due to rate limit")
		return &CoreRestartResponse{}, nil
	}

	cmh.log.Info("core restart requested")

	if err := cmh.coreState.Restart(); err != nil {
		cmh.log.Error("core restart failed", "error", err)
		return nil, err
	}

	cmh.log.Info("core successfully restarted")
	return &CoreRestartResponse{}, nil
}

// GetConfig - returns the current core configuration in JSON format.
func (cmh *coreManageHandlers) GetConfig(ctx context.Context, r *GetConfigRequest) (*GetConfigResponse, error) {

	cfg, err := cmh.confLoad.LoadConfig()
	if err != nil {
		cmh.log.Error("failed to load config", "error", err)
		return nil, err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		cmh.log.Error("failed to marshal config", "error", err)
		return nil, err
	}

	cmh.log.Debug("config requested")
	return &GetConfigResponse{Data: string(data)}, nil
}

// UploadConfig - uploads a new core configuration with rate-limiting.
func (cmh *coreManageHandlers) UploadConfig(ctx context.Context, r *UploadConfigRequest) (*UploadConfigResponse, error) {

	if !cmh.confSaveLim.InLimits() {
		cmh.log.Warn("config upload rejected due to rate limit")
		return nil, errors.New("too many upload requests")
	}

	if !json.Valid([]byte(r.Data)) {
		cmh.log.Warn("invalid JSON config format")
		return nil, errors.New("invalid JSON config format")
	}

	var cfg domain.CoreConfiguration
	if err := json.Unmarshal([]byte(r.Data), &cfg); err != nil {
		cmh.log.Warn("invalid config payload", "error", err)
		return nil, err
	}

	cmh.log.Info("config upload requested")

	if err := cmh.confSave.SaveConfig(cfg); err != nil {
		cmh.log.Error("failed to save config", "error", err)
		return nil, err
	}

	if r.RestartCore {
		cmh.log.Info("core restart requested")

		if err := cmh.coreState.Restart(); err != nil {
			cmh.log.Error("core restart failed", "error", err)
			return nil, err
		}

		cmh.log.Info("core successfully restarted")
	}

	cmh.log.Info("config successfully saved")
	return &UploadConfigResponse{}, nil
}

// ===================================================

type StatsActual interface {
	StatsNow(ctx context.Context) ([]domain.StatsSnapshot, error)
}

type journalHandlers struct {
	lastConns LastConnProvider
	statsNet  StatsActual

	log *slog.Logger

	UnimplementedJournalProviderServer
}

func NewJournalHandlers(lc LastConnProvider, sa StatsActual, log *slog.Logger) *journalHandlers {
	return &journalHandlers{
		lastConns: lc,
		statsNet:  sa,
		log:       log,
	}
}

// ConnectionJournal - streams the last connection records to the client.
func (jh *journalHandlers) ConnectionJournal(r *ConnectionJournalRequest, stream grpc.ServerStreamingServer[ConnectionMeta]) error {

	metaList, err := jh.lastConns.LastConnections(stream.Context(), int(r.Last))
	if err != nil {
		jh.log.Error("failed to load connection journal", "error", err)
		return err
	}

	ctx := stream.Context()

	for _, meta := range metaList {
		if err := ctx.Err(); err != nil {
			jh.log.Debug("connection journal stream canceled by client")
			return nil
		}

		dto := domain2dtoConnectionMeta(meta)
		if err := stream.Send(dto); err != nil {
			jh.log.Warn("failed to send connection journal item", "error", err)
			return err
		}
	}

	return nil
}

func (jh *journalHandlers) NetworkStatsNetworkStats(ctx context.Context, r *NetworkStatsRequest) (*NetworkStatsResponse, error) {
	stats, err := jh.statsNet.StatsNow(ctx)
	if err != nil {
		return nil, err
	}

	return domain2dtoNetworkStatsResponse(stats), nil
}
