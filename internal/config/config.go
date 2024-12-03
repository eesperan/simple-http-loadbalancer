package config

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type Frontend struct {
	Port int `yaml:"port"`
}

type Backend struct {
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight"`
}

type HealthCheck struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	Path     string        `yaml:"path"`
}

// Custom unmarshaler for HealthCheck to parse duration strings
func (h *HealthCheck) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawHealthCheck struct {
		Interval string `yaml:"interval"`
		Timeout  string `yaml:"timeout"`
		Path     string `yaml:"path"`
	}
	raw := &rawHealthCheck{}
	if err := unmarshal(raw); err != nil {
		return err
	}

	var err error
	if raw.Interval == "" {
		h.Interval = 10 * time.Second
	} else {
		h.Interval, err = time.ParseDuration(raw.Interval)
		if err != nil {
			return fmt.Errorf("invalid interval duration: %v", err)
		}
	}

	if raw.Timeout == "" {
		h.Timeout = 2 * time.Second
	} else {
		h.Timeout, err = time.ParseDuration(raw.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout duration: %v", err)
		}
	}

	if raw.Path == "" {
		h.Path = "/health"
	} else {
		h.Path = raw.Path
	}

	return nil
}

type Logging struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type Metrics struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type SSL struct {
	CertFile   string            `yaml:"certFile"`
	KeyFile    string            `yaml:"keyFile"`
	CAFile     string            `yaml:"caFile"`
	ClientAuth tls.ClientAuthType `yaml:"clientAuth"`
}

type Config struct {
	Frontends   []Frontend  `yaml:"frontends"`
	Backends    []string    `yaml:"backends"`
	HealthCheck HealthCheck `yaml:"healthcheck"`
	Logging     Logging     `yaml:"logging"`
	Metrics     Metrics     `yaml:"metrics"`
	SSL         *SSL        `yaml:"ssl"`
}

func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Set default values
	if config.HealthCheck.Path == "" {
		config.HealthCheck.Path = "/health"
	}
	if config.HealthCheck.Interval == 0 {
		config.HealthCheck.Interval = 10 * time.Second
	}
	if config.HealthCheck.Timeout == 0 {
		config.HealthCheck.Timeout = 2 * time.Second
	}
	if config.Metrics.Port == 0 {
		config.Metrics.Port = 9090
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}

	return config, nil
}
