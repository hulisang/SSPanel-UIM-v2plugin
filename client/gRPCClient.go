package client

import (
	"google.golang.org/grpc"
	"time"
)

func ConnectGRPC(address string, timeoutDuration time.Duration) (conn *grpc.ClientConn, err error) {
	timeout := time.After(timeoutDuration)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return
		case <-tick:
			conn, err = grpc.Dial(address, grpc.WithInsecure())
			if err == nil {
				return
			}
		}
	}

	return
}
