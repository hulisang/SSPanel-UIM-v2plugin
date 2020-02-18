package speedtest

import (
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test_measureLatency(t *testing.T) {
	tests := []struct {
		name   string
		client Client
		input  time.Duration
		want   time.Duration
	}{
		{
			name:   "Client.Get() error with DefaultErrorLatency input",
			client: &latencyErrorClient{},
			input:  DefaultErrorLatency,
			want:   DefaultErrorLatency,
		},
		{
			name:   "Client.Get() error with 10 Second input",
			client: &latencyErrorClient{},
			input:  10 * time.Second,
			want:   10 * time.Second,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{client: tc.client}
			if got, want := s.measureLatency(tc.input), tc.want; got != want {
				t.Fatalf("unexpected result:\n- want: %v\n-  got: %v",
					want, got)
			}
		})
	}
}

// latencyErrorClient is a client returns error from most of the methods.
type latencyErrorClient struct{}

func (c *latencyErrorClient) Log(_ string, _ ...interface{}) {}
func (c *latencyErrorClient) Config() (*Config, error) {
	return nil, errors.New("Config()")
}
func (c *latencyErrorClient) LoadConfig(_ chan ConfigRef) {}
func (c *latencyErrorClient) NewRequest(_ string, _ string, _ io.Reader) (*http.Request, error) {
	return nil, errors.New("NewRequest()")
}
func (c *latencyErrorClient) Get(_ string) (resp *Response, err error) {
	return nil, errors.New("Get()")
}
func (c *latencyErrorClient) Post(_ string, _ string, _ io.Reader) (*Response, error) {
	return nil, errors.New("Post()")
}
func (c *latencyErrorClient) AllServers() (*Servers, error) {
	return nil, errors.New("AllServers()")
}
func (c *latencyErrorClient) LoadAllServers(_ chan ServersRef) {}
func (c *latencyErrorClient) ClosestServers() (*Servers, error) {
	return nil, errors.New("ClosestServers()")
}
func (c *latencyErrorClient) LoadClosestServers(_ chan ServersRef) {}
