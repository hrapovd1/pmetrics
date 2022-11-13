package config

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

type environ struct {
	PollInterval   string `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval string `env:"REPORT_INTERVAL" envDefault:"10s"`
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
}

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	RetryCount     int
	ServerAddress  string
}

func (cfg *Config) NewAgent() error {
	var err error
	var envs environ
	if err = env.Parse(&envs); err != nil {
		return err
	}
	if cfg.PollInterval, err = parseInterval(envs.PollInterval); err != nil {
		return err
	}
	if cfg.ReportInterval, err = parseInterval(envs.ReportInterval); err != nil {
		return err
	}
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

func parseInterval(interval string) (time.Duration, error) {
	value := make([]string, 0)
	for _, ch := range interval {
		value = append(value, string(ch))
	}
	switch value[len(value)-1] {
	case "s":
		seconds, err := strconv.Atoi(strings.Join(value[:len(value)-1], ""))
		return time.Duration(seconds) * time.Second, err
	case "m":
		minutes, err := strconv.Atoi(strings.Join(value[:len(value)-1], ""))
		return time.Duration(minutes) * time.Minute, err
	default:
		return time.Second, errors.New("wrong interval format")
	}
}
