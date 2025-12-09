package xrayapi

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/xtls/xray-core/app/proxyman/command"
	handlerService "github.com/xtls/xray-core/app/proxyman/command"
	statsService "github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	trafficReg       = regexp.MustCompile(`(inbound|outbound)>>>([^>]+)>>>traffic>>>(downlink|uplink)`)
	clientTrafficReg = regexp.MustCompile(`user>>>([^>]+)>>>traffic>>>(downlink|uplink)`)
)

func validateSocket(s string) error {
	if addrPort, err := netip.ParseAddrPort(s); err == nil {
		_ = addrPort
		return nil
	}

	if st, err := os.Stat(s); err == nil {
		if (st.Mode() & os.ModeSocket) != 0 {
			return nil
		}
		return errors.New("path exists but is not a unix socket")
	}

	return errors.New("invalid socket: neither network nor unix socket")
}

type XrayAPI struct {
	HandlerServiceClient *handlerService.HandlerServiceClient
	StatsServiceClient   *statsService.StatsServiceClient
	grpcClient           *grpc.ClientConn
	isConnected          atomic.Bool
}

func New(addr string) (*XrayAPI, error) {
	if err := validateSocket(addr); err != nil {
		return nil, fmt.Errorf("failed to connect to Xray API: %w", err)
	}

	withTransportCred := grpc.WithTransportCredentials(insecure.NewCredentials())

	conn, err := grpc.NewClient(addr, withTransportCred, grpc.WithIdleTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Xray API: %w", err)
	}

	x := &XrayAPI{}

	x.grpcClient = conn
	x.isConnected.Store(true)

	hsClient := command.NewHandlerServiceClient(conn)
	ssClient := statsService.NewStatsServiceClient(conn)

	x.HandlerServiceClient = &hsClient
	x.StatsServiceClient = &ssClient

	return x, nil
}

func (x *XrayAPI) Close() error {
	var err error

	if x.grpcClient != nil {
		err = x.grpcClient.Close()
	}

	x.isConnected.Store(false)
	x.HandlerServiceClient = nil
	x.StatsServiceClient = nil

	return err
}

func (x *XrayAPI) grpcNotNil() error {
	if x.grpcClient == nil {
		return errors.New("xray api is not initialized")
	}
	return nil
}

func (x *XrayAPI) GetTraffic(reset bool) ([]Traffic, []ClientTraffic, error) {
	if err := x.grpcNotNil(); err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if x.StatsServiceClient == nil {
		return nil, nil, errors.New("xray StatusServiceClient is not initialized")
	}

	resp, err := (*x.StatsServiceClient).QueryStats(ctx, &statsService.QueryStatsRequest{Reset_: reset})
	if err != nil {
		return nil, nil, err
	}

	t, ct := parseStats(resp.Stat)

	return t, ct, nil
}
