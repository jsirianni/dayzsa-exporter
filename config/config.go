// Package config provides the configuration for the exporter
package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// CollectionInterval is the default interval
	CollectionInterval = 60 * time.Second
)

// NewFromFile reads a configuration from a file
func NewFromFile(path string) (*Config, error) {
	b, err := os.ReadFile(path) // #nosec G304 - Path is configured by the user
	if err != nil {
		return nil, fmt.Errorf("read file: %s: %w", path, err)
	}

	c := &Config{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return c, nil
}

// Config is the configuration for the exporter
type Config struct {
	Interval time.Duration `yaml:"interval"`
	Servers  []Server      `yaml:"servers"`
	Host     string        `yaml:"host"`
}

// Server is a DayZ server endpoint
type Server struct {
	IP   string `yaml:"ip"`
	Port int    `yaml:"port"`
}

// String returns the server as a string
func (s Server) String() string {
	return net.JoinHostPort(s.IP, strconv.Itoa(s.Port))
}
