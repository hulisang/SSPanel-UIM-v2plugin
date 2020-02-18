package speedtest

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
)

type Client interface {
	Log(format string, a ...interface{})
	Config() (*Config, error)
	LoadConfig(ret chan ConfigRef)
	NewRequest(method string, url string, body io.Reader) (*http.Request, error)
	Get(url string) (resp *Response, err error)
	Post(url string, bodyType string, body io.Reader) (resp *Response, err error)
	AllServers() (*Servers, error)
	LoadAllServers(ret chan ServersRef)
	ClosestServers() (*Servers, error)
	LoadClosestServers(ret chan ServersRef)
}

type client struct {
	http.Client
	opts           *Opts
	mutex          sync.Mutex
	config         chan ConfigRef
	allServers     chan ServersRef
	closestServers chan ServersRef
}

type Response http.Response

func NewClient(opts *Opts) Client {
	dialer := &net.Dialer{
		Timeout:   opts.Timeout,
		KeepAlive: opts.Timeout,
	}

	if len(opts.Interface) != 0 {
		dialer.LocalAddr = &net.IPAddr{IP: net.ParseIP(opts.Interface)}
		if dialer.LocalAddr == nil {
			newErrorf("Invalid source IP: %s\n", opts.Interface).AtWarning().WriteToLog()
		}
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  dialer.Dial,
		TLSHandshakeTimeout:   opts.Timeout,
		ExpectContinueTimeout: opts.Timeout,
	}

	client := &client{
		Client: http.Client{
			Transport: transport,
			Timeout:   opts.Timeout,
		},
		opts: opts,
	}

	return client
}

func (client *client) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	if strings.HasPrefix(url, ":") {
		if client.opts.Secure {
			url = "https" + url
		} else {
			url = "http" + url
		}
	}
	req, err := http.NewRequest(method, url, body)
	if err == nil {
		req.Header.Set(
			"User-Agent",
			"Mozilla/5.0 "+
				fmt.Sprintf("(%s; U; %s; en-us)", runtime.GOOS, runtime.GOARCH)+
				fmt.Sprintf("Go/%s", runtime.Version())+
				fmt.Sprintf("(KHTML, like Gecko) speedtest-cli/%s", Version))
	}
	return req, err
}

func (client *client) Get(url string) (resp *Response, err error) {
	req, err := client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	htResp, err := client.Client.Do(req)

	return (*Response)(htResp), err
}

func (client *client) Post(url string, bodyType string, body io.Reader) (resp *Response, err error) {
	req, err := client.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", bodyType)
	htResp, err := client.Client.Do(req)

	return (*Response)(htResp), err
}

func (resp *Response) ReadContent() ([]byte, error) {
	content, err := ioutil.ReadAll(resp.Body)
	cerr := resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return content, cerr
	}
	return content, nil
}

func (resp *Response) ReadXML(out interface{}) error {
	content, err := resp.ReadContent()
	if err != nil {
		return err
	}
	return xml.Unmarshal(content, out)
}
