package commands

import (
	"github.com/eterline/xraymon/internal/domain"
	"google.golang.org/protobuf/types/known/durationpb"
)

func domain2dtoCoreStatusResponse(s domain.CoreStatus) *CoreStatusResponse {
	return &CoreStatusResponse{
		Working:     s.Working,
		LastLog:     s.LastLog,
		WorkingTime: durationpb.New(s.WorkingTime),
	}
}

func domain2dtoConnectionMeta(s domain.ConnectionMetadata) *ConnectionMeta {
	return &ConnectionMeta{
		Client:   s.Client,
		Server:   s.Server,
		Proto:    selectType(s.Proto),
		Inbound:  s.Inbound,
		Outbound: s.Outbound,
		User:     s.User,
	}
}

func selectType(v string) NetType {
	switch v {
	case "udp":
		return NetType_UDP
	case "tcp":
		return NetType_TCP
	default:
		return NetType_HTTP
	}
}

func domain2dtoNetworkStatsResponse(s []domain.StatsSnapshot) *NetworkStatsResponse {
	r := &NetworkStatsResponse{}
	r.Stats = make([]*StatsMeta, 0, len(s))

	for _, snap := range s {
		meta := &StatsMeta{
			Alias: snap.Name,
			Type:  determConnType(snap.Type),
			Io:    domain2connIO(snap.IO),
		}

		r.Stats = append(r.Stats, meta)
	}

	return r
}

func domain2connIO(c domain.DataIO) *ConnectionIO {
	return &ConnectionIO{
		BytesRx:       c.RX,
		BytesTx:       c.TX,
		BytesPerSecRx: c.PerSecRX,
		BytesPerSecTx: c.PerSecTX,
	}
}

func determConnType(t domain.StatsType) ConnectionType {
	switch t {
	case domain.TypeInbound:
		return ConnectionType_INBOUND
	case domain.TypeOubnound:
		return ConnectionType_OUTBOUND
	default:
		return ConnectionType_USER
	}
}
