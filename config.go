package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	defaultListenAddress = ":9549"
	defaultMetricsPath   = "/metrics"
)

type LocalConfig struct {
	Listen    ListenConfig    `yaml:"listen"`
	Brickd    BrickdConfig    `yaml:"brickd"`
	Collector CollectorConfig `yaml:"collector"`
}

type ListenConfig struct {
	Address     string `yaml:"address"`
	MetricsPath string `yaml:"metrics_path"`
}

type BrickdConfig struct {
	Address  string `yaml:"address"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type CollectorConfig struct {
	LogLevel string `yaml:"log_level"`
}

func parseConfig() (*LocalConfig, error) {
	var configFile = flag.String("config.file", "", "Path to configuration file.")
	flag.Parse()

	if *configFile == "" {
		return defaultConfig()
	}

	file, err := os.Open(*configFile)
	if err != nil {
		return nil, fmt.Errorf("can not open config file: %s", err)
	}

	config := &LocalConfig{}
	if err := yaml.NewDecoder(file).Decode(config); err != nil {
		return nil, fmt.Errorf("error decoding config file %q: %s", *configFile, err)
	}

	return config, nil
}

func defaultConfig() (*LocalConfig, error) {
	return &LocalConfig{
		Brickd: BrickdConfig{
			Address: "localhost:4223",
		},
		Listen: ListenConfig{
			Address:     defaultListenAddress,
			MetricsPath: defaultMetricsPath,
		},
		Collector: CollectorConfig{
			LogLevel: "info",
		},
	}, nil
}
