package speedtest

import (
	"sort"
	"strings"
	"time"
)

const DefaultLatencyMeasureTimes = 4
const DefaultErrorLatency = time.Hour

// Measures latencies for each server.
// Returns server list sorted by latencies.
// This is synchronous operation, because multiple simultaneous requests may affect results.
func (servers *Servers) MeasureLatencies(times uint, errorLatency time.Duration) *Servers {
	first := true
	for _, server := range servers.List {
		if first {
			first = false
			server.client.Log("Measuring server latencies...")
		}
		server.doMeasureLatency(times, errorLatency)
	}

	latencies := &serverLatencies{List: make([]*Server, servers.Len())}
	copy(latencies.List, servers.List)
	sort.Sort(latencies)

	return (*Servers)(latencies)
}

type serverLatencies Servers

func (servers *serverLatencies) Len() int {
	return len(servers.List)
}

func (servers *serverLatencies) Less(i, j int) bool {
	return servers.List[i].Latency < servers.List[j].Latency
}

func (servers *serverLatencies) Swap(i, j int) {
	temp := servers.List[i]
	servers.List[i] = servers.List[j]
	servers.List[j] = temp
}

func (server *Server) MeasureLatency(times uint, errorLatency time.Duration) time.Duration {
	server.client.Log("Measuring server latency...\n")
	return server.doMeasureLatency(times, errorLatency)
}

func (server *Server) doMeasureLatency(times uint, errorLatency time.Duration) time.Duration {

	var results time.Duration = 0
	var i uint

	for i = 0; i < times; i++ {
		results += server.measureLatency(errorLatency)
	}

	server.Latency = time.Duration(results / time.Duration(times))

	return server.Latency
}

func (server *Server) measureLatency(errorLatency time.Duration) time.Duration {
	url := server.RelativeURL("latency.txt")
	start := time.Now()
	resp, err := server.client.Get(url)
	duration := time.Since(start)
	if resp != nil {
		url = resp.Request.URL.String()
	}
	if err != nil {
		server.client.Log("[%s] Failed to detect latency: %v\n", url, err)
		return errorLatency
	}
	if resp.StatusCode != 200 {
		server.client.Log("[%s] Invalid latency detection HTTP status: %d\n", url, resp.StatusCode)
		duration = errorLatency
	}
	content, err := resp.ReadContent()
	if err != nil {
		server.client.Log("[%s] Failed to read latency response: %v\n", url, err)
		duration = errorLatency
	}
	if !strings.HasPrefix(string(content), "test=test") {
		server.client.Log("[%s] Invalid latency response: %s\n", url, content)
		duration = errorLatency
	}
	return duration
}
