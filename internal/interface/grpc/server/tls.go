package server

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewTLSGrpcServer(certFile, keyFile string) (*grpc.Server, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	creds := credentials.NewTLS(tlsCfg)

	return grpc.NewServer(
		grpc.Creds(creds),
	), nil
}
