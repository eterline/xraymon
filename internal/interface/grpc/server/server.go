package server

import (
	"context"
	"errors"
	"net"
	"time"

	"google.golang.org/grpc"
)

/*
GrpcServer – gRPC server wrapper with graceful shutdown support.
Holds underlying grpc.Server and listener.
*/
type GrpcServer struct {
	server   *grpc.Server
	listener net.Listener
	timeout  time.Duration
}

// Option – functional option for GrpcServer.
type Option func(*GrpcServer)

// WithTimeout – sets shutdown timeout.
func WithTimeout(d time.Duration) Option {
	return func(s *GrpcServer) {
		s.timeout = d
	}
}

// NewGrpcServerWrapper – creates wrapper around existing grpc.Server.
func NewGrpcServerWrapper(srv *grpc.Server, addr string, opts ...Option) (*GrpcServer, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &GrpcServer{
		server:   srv,
		listener: lis,
		timeout:  5 * time.Second,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Run – starts the gRPC server and listens for context cancellation.
func (s *GrpcServer) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.server.Serve(s.listener)
	}()

	select {
	case <-ctx.Done():
		stopped := make(chan struct{})
		go func() {
			s.server.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
			return nil
		case <-time.After(s.timeout):
			s.server.Stop()
			return errors.New("graceful shutdown timed out, forced stop")
		}

	case err := <-errCh:
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	}
}

// Close – immediately stops the server.
func (s *GrpcServer) Close() {
	s.server.Stop()
}
