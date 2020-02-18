package speedtest

import (
	"encoding/xml"
	"strconv"
	"strings"
)

type ClientConfig struct {
	Coordinates
	IP                 string  `xml:"ip,attr"`
	ISP                string  `xml:"isp,attr"`
	ISPRating          float32 `xml:"isprating,attr"`
	ISPDownloadAverage uint32  `xml:"ispdlavg,attr"`
	ISPUploadAverage   uint32  `xml:"ispulavg,attr"`
	Rating             float32 `xml:"rating,attr"`
	LoggedIn           uint8   `xml:"loggedin,attr"`
}

type ConfigTime struct {
	Upload   uint32
	Download uint32
}

type ConfigTimes []ConfigTime

type Config struct {
	Client ClientConfig `xml:"client"`
	Times  ConfigTimes  `xml:"times"`
}

func (client *client) Log(format string, a ...interface{}) {
	if !client.opts.Quiet {
		newErrorf(format, a...).AtInfo().WriteToLog()
	}
}

type ConfigRef struct {
	Config *Config
	Error  error
}

func (client *client) Config() (*Config, error) {
	configChan := make(chan ConfigRef)
	client.LoadConfig(configChan)
	configRef := <-configChan
	return configRef.Config, configRef.Error
}

func (client *client) LoadConfig(ret chan ConfigRef) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.config == nil {
		client.config = make(chan ConfigRef)
		go client.loadConfig()
	}

	go func() {
		result := <-client.config
		ret <- result
		client.config <- result
	}()
}

func (client *client) loadConfig() {
	client.Log("Retrieving speedtest.net configuration...")

	result := ConfigRef{}

	resp, err := client.Get("://www.speedtest.net/speedtest-config.php")
	if err != nil {
		result.Error = err
	} else {
		config := &Config{}
		err = resp.ReadXML(config)
		if err != nil {
			result.Error = err
		} else {
			result.Config = config
		}
	}

	client.config <- result
}

func (times ConfigTimes) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		name := attr.Name.Local
		if dl := strings.HasPrefix(name, "dl"); dl || strings.HasPrefix(name, "ul") {
			num, err := strconv.Atoi(name[2:])
			if err != nil {
				return err
			}
			if num > cap(times) {
				newTimes := make([]ConfigTime, num)
				copy(newTimes, times)
				times = newTimes[0:num]
			}

			speed, err := strconv.ParseUint(attr.Value, 10, 32)

			if err != nil {
				return err
			}
			if dl {
				times[num-1].Download = uint32(speed)
			} else {
				times[num-1].Upload = uint32(speed)
			}
		}
	}

	return d.Skip()
}
