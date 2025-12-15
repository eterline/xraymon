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

func domain2stoConnectionMeta(s domain.ConnectionMetadata) *ConnectionMeta {
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
