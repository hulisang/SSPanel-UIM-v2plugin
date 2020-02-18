package speedtest

import (
	"testing"
	"time"
)

func TestUpload(t *testing.T) {
	tests := []struct {
		name string
		opts Opts
	}{
		{
			name: "default options",
			opts: Opts{},
		},
		{
			name: "quiet option",
			opts: Opts{Quiet: true},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// set timeout to avoid the longer tests.
			tc.opts.Timeout = 10 * time.Second
			c := NewClient(&tc.opts)
			if _, err := c.Config(); err != nil {
				t.Fatalf("unexpected config error: %v", err)
			}
			s, err := c.ClosestServers()
			if err != nil {
				t.Fatalf("unexpected server selection error: %v", err)
			}
			// pick the firstest server to test.
			upload := s.MeasureLatencies(
				DefaultLatencyMeasureTimes,
				DefaultErrorLatency,
			).First().UploadSpeed()
			t.Logf("upload %d bps", upload)
		})
	}
}
