package speedtest

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"time"
)

type ServerID uint64

type Server struct {
	Coordinates
	URL      string        `xml:"url,attr"`
	Name     string        `xml:"name,attr"`
	Country  string        `xml:"country,attr"`
	CC       string        `xml:"cc,attr"`
	Sponsor  string        `xml:"sponsor,attr"`
	ID       ServerID      `xml:"id,attr"`
	URL2     string        `xml:"url2,attr"`
	Host     string        `xml:"host,attr"`
	client   Client        `xml:"-"`
	Distance float64       `xml:"-"`
	Latency  time.Duration `xml:"-"`
}

func (s *Server) String() string {
	return fmt.Sprintf("%8d: %s (%s, %s) [%.2f km] %s", s.ID, s.Sponsor, s.Name, s.Country, s.Distance, s.URL)
}

func (s *Server) RelativeURL(local string) string {
	u, err := url.Parse(s.URL)
	if err != nil {
		log.Fatalf("[%s] Failed to parse server URL: %v\n", s.URL, err)
		return ""
	}
	localURL, err := url.Parse(local)
	if err != nil {
		log.Fatalf("Failed to parse local URL `%s`: %v\n", local, err)
	}
	return u.ResolveReference(localURL).String()
}

type Servers struct {
	List []*Server `xml:"servers>server"`
}

type ServersRef struct {
	Servers *Servers
	Error   error
}

func (servers *Servers) First() *Server {
	if len(servers.List) == 0 {
		return nil
	}
	return servers.List[0]
}

func (servers *Servers) Find(id ServerID) *Server {
	for _, server := range servers.List {
		if server.ID == id {
			return server
		}
	}
	return nil
}

func (servers *Servers) Len() int {
	return len(servers.List)
}

func (servers *Servers) Less(i, j int) bool {
	server1 := servers.List[i]
	server2 := servers.List[j]
	if server1.ID == server2.ID {
		return false
	}
	if server1.Distance < server2.Distance {
		return true
	}
	if server1.Distance > server2.Distance {
		return false
	}
	return server1.ID < server2.ID
}

func (servers *Servers) Swap(i, j int) {
	temp := servers.List[i]
	servers.List[i] = servers.List[j]
	servers.List[j] = temp
}

func (servers *Servers) truncate(max int) *Servers {
	size := servers.Len()
	if size <= max {
		return servers
	}
	return &Servers{servers.List[:max]}
}

func (servers *Servers) String() string {
	out := ""
	for _, server := range servers.List {
		out += server.String() + "\n"
	}
	return out
}

func (servers *Servers) append(other *Servers) *Servers {
	if servers == nil {
		return other
	}
	servers.List = append(servers.List, other.List...)
	return servers
}

func (servers *Servers) sort(client Client, config *Config) {
	for _, server := range servers.List {
		server.client = client
		server.Distance = server.DistanceTo(config.Client.Coordinates)
	}
	sort.Sort(servers)
}

func (servers *Servers) deduplicate() {
	dedup := make([]*Server, 0, len(servers.List))
	var prevId ServerID = 0
	for _, server := range servers.List {
		if prevId != server.ID {
			prevId = server.ID
			dedup = append(dedup, server)
		}
	}
	servers.List = dedup
}

var serverURLs = [...]string{
	"://www.speedtest.net/speedtest-servers-static.php",
	"://c.speedtest.net/speedtest-servers-static.php",
	"://www.speedtest.net/speedtest-servers.php",
	"://c.speedtest.net/speedtest-servers.php",
}

var NoServersError error = errors.New("No servers available")

func (client *client) AllServers() (*Servers, error) {
	serversChan := make(chan ServersRef)
	client.LoadAllServers(serversChan)
	serversRef := <-serversChan
	return serversRef.Servers, serversRef.Error
}

func (client *client) LoadAllServers(ret chan ServersRef) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.allServers == nil {
		client.allServers = make(chan ServersRef)
		go client.loadServers()
	}

	go func() {
		result := <-client.allServers
		ret <- result
		client.allServers <- result // Make it available again
	}()
}

func (client *client) loadServers() {
	configChan := make(chan ConfigRef)
	client.LoadConfig(configChan)

	client.Log("Retrieving speedtest.net server list...")

	serversChan := make(chan *Servers, len(serverURLs))
	for _, url := range serverURLs {
		go client.loadServersFrom(url, serversChan)
	}

	var servers *Servers

	for range serverURLs {
		servers = servers.append(<-serversChan)
	}

	result := ServersRef{}

	if servers.Len() == 0 {
		result.Error = NoServersError
	} else {
		configRef := <-configChan
		if configRef.Error != nil {
			result.Error = configRef.Error
		} else {
			servers.sort(client, configRef.Config)
			servers.deduplicate()
			result.Servers = servers
		}
	}

	client.allServers <- result
}

func (client *client) loadServersFrom(url string, ret chan *Servers) {
	resp, err := client.Get(url)
	if resp != nil {
		url = resp.Request.URL.String()
	}
	if err != nil {
		client.Log("[%s] Failed to retrieve server list: %v", url, err)
	}

	servers := &Servers{}
	if err = resp.ReadXML(servers); err != nil {
		client.Log("[%s] Failed to read server list: %v", url, err)
	}
	ret <- servers
}

func (client *client) ClosestServers() (*Servers, error) {
	serversChan := make(chan ServersRef)
	client.LoadClosestServers(serversChan)
	serversRef := <-serversChan
	return serversRef.Servers, serversRef.Error
}

func (client *client) LoadClosestServers(ret chan ServersRef) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.closestServers == nil {
		client.closestServers = make(chan ServersRef)
		go client.loadClosestServers()
	}

	go func() {
		result := <-client.closestServers
		ret <- result
		client.closestServers <- result // Make it available again
	}()
}

func (client *client) loadClosestServers() {
	serversChan := make(chan ServersRef)
	client.LoadAllServers(serversChan)
	serversRef := <-serversChan
	if serversRef.Error != nil {
		client.closestServers <- serversRef
	} else {
		client.closestServers <- ServersRef{serversRef.Servers.truncate(5), nil}
	}
}
