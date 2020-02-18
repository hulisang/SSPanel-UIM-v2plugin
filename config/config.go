package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/utility"
	"io/ioutil"
	"os"
	"path/filepath"
	"v2ray.com/core/common/errors"
	"v2ray.com/core/common/platform"
	"v2ray.com/core/infra/conf"
)

//var (
//	CommandLine = flag.NewFlagSet(os.Args[0]+"-sspanel_v3_mod_Uim_plugin", flag.ContinueOnError)
//
//	ConfigFile = CommandLine.String("config", "", "Config file for V2Ray.")
//	_          = CommandLine.Bool("version", false, "Show current version of V2Ray.")
//	Test       = CommandLine.Bool("test", false, "Test config file only, without launching V2Ray server.")
//	_          = CommandLine.String("format", "json", "Format of input file.")
//	_          = CommandLine.Bool("plugin", false, "True to load plugins.")
//)

var (
	CommandLine = flag.NewFlagSet(os.Args[0]+"-sspanel_v3_mod_Uim_plugin", flag.ContinueOnError)

	ConfigFile = CommandLine.String("config", "", "Config file for V2Ray.")
	_          = CommandLine.Bool("version", false, "Show current version of V2Ray.")
	Test       = CommandLine.Bool("test", false, "Test config file only, without launching V2Ray server.")
	_          = CommandLine.String("format", "json", "Format of input file.")
	_          = CommandLine.Bool("plugin", false, "True to load plugins.")
)

type Config struct {
	NodeID             uint   `json:"nodeId"`
	CheckRate          int    `json:"checkRate"`
	PanelUrl           string `json:"panelUrl"`
	PanelKey           string `json:"panelKey"`
	SpeedTestCheckRate int    `json:"speedTestCheckrate"`
	V2rayConfig        *conf.Config
}

func GetConfig() (*Config, error) {
	type config struct {
		*conf.Config
		SSPanel *Config `json:"sspanel"`
	}

	configFile := GetConfigFilePath()
	// Open our jsonFile
	jsonFile, err := os.Open(configFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, errors.New("failed to open config: ", configFile).Base(err)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	cfg := &config{}
	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		return nil, err
	}
	if cfg.SSPanel != nil {
		cfg.SSPanel.V2rayConfig = cfg.Config
		if err = CheckCfg(cfg.SSPanel); err != nil {
			return nil, err
		}
	}
	return cfg.SSPanel, err
}

func CheckCfg(cfg *Config) error {

	if cfg.V2rayConfig.Api == nil {
		return errors.New("Api must be set")
	}

	apiTag := cfg.V2rayConfig.Api.Tag
	if len(apiTag) == 0 {
		return errors.New("Api tag can't be empty")
	}

	services := cfg.V2rayConfig.Api.Services
	if !utility.InStr("HandlerService", services) {
		return errors.New("Api service, HandlerService, must be enabled")
	}
	if !utility.InStr("StatsService", services) {
		return errors.New("Api service, StatsService, must be enabled")
	}

	if cfg.V2rayConfig.Stats == nil {
		return errors.New("Stats must be enabled")
	}

	if apiInbound := GetInboundConfigByTag(apiTag, cfg.V2rayConfig.InboundConfigs); apiInbound == nil {
		return errors.New(fmt.Sprintf("Miss an inbound tagged %s", apiTag))
	} else if apiInbound.Protocol != "dokodemo-door" {
		return errors.New(fmt.Sprintf("The protocol of inbound tagged %s must be \"dokodemo-door\"", apiTag))
	} else {
		if apiInbound.ListenOn == nil || apiInbound.PortRange == nil {
			return errors.New(fmt.Sprintf("Fields, \"listen\" and \"port\", of inbound tagged %s must be set", apiTag))
		}
	}

	return nil
}

func GetInboundConfigByTag(apiTag string, inbounds []conf.InboundDetourConfig) *conf.InboundDetourConfig {
	for _, inbound := range inbounds {
		if inbound.Tag == apiTag {
			return &inbound
		}
	}
	return nil
}

func GetConfigFilePath() string {
	if len(*ConfigFile) > 0 {
		return *ConfigFile
	}

	if workingDir, err := os.Getwd(); err == nil {
		configFile := filepath.Join(workingDir, "config.json")
		if fileExists(configFile) {
			return configFile
		}
	}

	if configFile := platform.GetConfigurationPath(); fileExists(configFile) {
		return configFile
	}

	return ""
}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}
