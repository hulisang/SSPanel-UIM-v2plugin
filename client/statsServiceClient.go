package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"strings"
	statsservice "v2ray.com/core/app/stats/command"
)

type StatsServiceClient struct {
	statsservice.StatsServiceClient
}

func NewStatsServiceClient(client *grpc.ClientConn) *StatsServiceClient {
	return &StatsServiceClient{
		StatsServiceClient: statsservice.NewStatsServiceClient(client),
	}
}

// traffic
func (s *StatsServiceClient) GetUserUplink(email string) (uint64, error) {
	return s.GetUserTraffic(fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email), true)
}

func (s *StatsServiceClient) GetUserDownlink(email string) (uint64, error) {
	return s.GetUserTraffic(fmt.Sprintf("user>>>%s>>>traffic>>>downlink", email), true)
}

func (s *StatsServiceClient) GetUserTraffic(name string, reset bool) (uint64, error) {
	req := &statsservice.GetStatsRequest{
		Name:   name,
		Reset_: reset,
	}

	res, err := s.GetStats(context.Background(), req)
	if err != nil {
		if status, ok := status.FromError(err); ok && strings.HasSuffix(status.Message(), fmt.Sprintf("%s not found.", name)) {
			return 0, nil
		}

		return 0, err
	}

	return uint64(res.Stat.Value), nil
}

// ips

func (s *StatsServiceClient) GetUserIPs(email string) ([]string, error) {
	name := fmt.Sprintf("user>>>%s>>>traffic>>>ips", email)
	req := &statsservice.GetStatsRequest{
		Name:   name,
		Reset_: true,
	}

	res, err := s.GetStats(context.Background(), req)

	if err != nil {
		if status, ok := status.FromError(err); ok && strings.HasSuffix(status.Message(), fmt.Sprintf("%s not found.", name)) {
			return []string{}, nil
		}
		return []string{}, err
	}
	ips := strings.Split(res.Stat.Name, ";")
	if len(ips) > 1 {
		ips = ips[1:]
	} else {
		ips = []string{}
	}
	return ips, nil
}
