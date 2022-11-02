package config

import "time"

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
	ServerPort     string
}

var AgentConfig = Config{
	PollInterval:   2 * time.Second,
	ReportInterval: 10 * time.Second,
	ServerAddress:  "127.0.0.1",
	ServerPort:     "8080",
}

var ServerConfig = Config{
	ServerAddress: "127.0.0.1",
	ServerPort:    "8080",
}
