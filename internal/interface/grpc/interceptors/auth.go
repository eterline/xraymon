// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package interceptors

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type TesterBearer interface {
	TestBearer(bearer string) bool
}

func AuthInterceptor(b TesterBearer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		tokens := md.Get("authorization")
		if !b.TestBearer(tokens[0]) {
			return nil, errors.New("invalid token")
		}

		return handler(ctx, req)
	}
}
