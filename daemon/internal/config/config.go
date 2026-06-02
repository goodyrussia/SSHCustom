package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

type Config struct {
	SSHHost        string
	SSHPort        int
	SSHUser        string
	SSHPassword    string
	SSHMode        string
	PayloadEnabled bool
	Payload        string
	SocksPort      int
	TProxyPort     int
	DNSForward     bool
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := &Config{
		SSHPort:    22,
		SocksPort:  1080,
		TProxyPort: 1081,
		DNSForward: true,
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		switch key {
		case "ssh_host":
			cfg.SSHHost = val
		case "ssh_port":
			cfg.SSHPort, _ = strconv.Atoi(val)
		case "ssh_user":
			cfg.SSHUser = val
		case "ssh_password":
			cfg.SSHPassword = val
		case "ssh_mode":
			cfg.SSHMode = val
		case "payload_enabled":
			cfg.PayloadEnabled = val == "1" || val == "true"
		case "payload":
			cfg.Payload = val
		case "socks_port":
			cfg.SocksPort, _ = strconv.Atoi(val)
		case "tproxy_port":
			cfg.TProxyPort, _ = strconv.Atoi(val)
		}
	}
	return cfg, nil
}

type AtomicConfig struct {
	v atomic.Value
}

func NewAtomicConfig(c *Config) *AtomicConfig {
	a := &AtomicConfig{}
	a.Store(c)
	return a
}

func (a *AtomicConfig) Get() *Config {
	return a.v.Load().(*Config)
}

func (a *AtomicConfig) Store(c *Config) {
	a.v.Store(c)
}