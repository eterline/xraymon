package xraycommon

import (
	"context"
	"fmt"
	"time"

	"github.com/cespare/xxhash"
	"github.com/eterline/xraymon/internal/domain"
	xrayapi "github.com/eterline/xraymon/internal/infra/xray/api"
)

type statsProvider struct {
	api *xrayapi.XrayAPI
}

func NewStatsProvider() (*statsProvider, error) {
	api, err := xrayapi.New(apiListenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed init stats provider: %w", err)
	}

	p := &statsProvider{
		api: api,
	}

	return p, nil
}

func trafficKey(t xrayapi.Traffic) uint64 {
	var kind uint64
	switch t.Type {
	case "inbound":
		kind = 1
	case "outbound":
		kind = 2
	default:
		kind = 0
	}

	return (kind << 62) | (xxhash.Sum64String(t.Tag) & ((1 << 62) - 1))
}

func (sp *statsProvider) Stats(ctx context.Context) ([]domain.StatsSnapshot, error) {
	tr0, cl0, err := sp.api.GetTraffic(ctx, false)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, context.Canceled
	case <-time.After(1 * time.Second):
	}

	tr1, cl1, err := sp.api.GetTraffic(ctx, false)
	if err != nil {
		return nil, err
	}

	snapshots := make([]domain.StatsSnapshot, 0, len(tr1)+len(cl1))

	/* ---------- CLIENTS ---------- */

	clPrev := make(map[string]xrayapi.ClientTraffic, len(cl0))
	for _, c := range cl0 {
		clPrev[c.Email] = c
	}

	for _, c := range cl1 {
		prev, ok := clPrev[c.Email]
		if !ok {
			continue
		}

		s := domain.NewUserMetric(c.Email)

		s.IO.IncRX(c.RX - prev.RX)
		s.IO.IncTX(c.TX - prev.TX)

		snapshots = append(snapshots, s)
	}

	/* ---------- TRAFFIC (IN / OUT) ---------- */

	trPrev := make(map[uint64]xrayapi.Traffic, len(tr0))
	for _, t := range tr0 {
		trPrev[trafficKey(t)] = t
	}

	for _, t := range tr1 {
		prev, ok := trPrev[trafficKey(t)]
		if !ok {
			continue
		}

		var s domain.StatsSnapshot
		switch {
		case t.IsInbound():
			s = domain.NewInboundMetric(t.Tag)
		case t.IsOutbound():
			s = domain.NewOutboundMetric(t.Tag)
		default:
			continue
		}

		s.IO.IncRX(t.RX - prev.RX)
		s.IO.IncTX(t.TX - prev.TX)

		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}
