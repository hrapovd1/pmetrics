package config

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type environ struct {
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
}

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	RetryCount     int
	ServerAddress  string
}

func (cfg *Config) NewAgent() error {
	var envs environ
	err := env.Parse(&envs)
	cfg.PollInterval = time.Duration(envs.PollInterval) * time.Second
	cfg.ReportInterval = time.Duration(envs.ReportInterval) * time.Second
	cfg.ServerAddress = envs.Address
	cfg.RetryCount = 3
	return err
}

func (cfg *Config) NewServer() error {
	var envs environ
	err := env.Parse(&envs)
	cfg.ServerAddress = envs.Address
	return err
}

var AgentConfig = Config{
	PollInterval:   2 * time.Second,
	ReportInterval: 10 * time.Second,
	RetryCount:     3,
	ServerAddress:  "127.0.0.1",
}

var ServerConfig = Config{
	ServerAddress: "127.0.0.1",
}
