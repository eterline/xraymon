// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package interceptors

import (
	"context"
	"errors"
	"net"
	"net/netip"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type SubnetFilter interface {
	InAllowedSubnets(ip netip.Addr) bool
}

func SubnetAuthInterceptor(f SubnetFilter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, errors.New("unable to get peer info")
		}

		tcpAddr, ok := p.Addr.(*net.TCPAddr)
		if !ok {
			return nil, errors.New("peer is not TCP")
		}

		ip, _ := netip.AddrFromSlice(tcpAddr.IP)

		if !f.InAllowedSubnets(ip) {
			return nil, errors.New("unauthorized: IP not allowed")
		}

		return handler(ctx, req)
	}
}
