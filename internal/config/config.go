package config

import "time"

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	RetryCount     int
	ServerAddress  string
	ServerPort     string
}

var AgentConfig = Config{
	PollInterval:   2 * time.Second,
	ReportInterval: 10 * time.Second,
	RetryCount:     3,
	ServerAddress:  "127.0.0.1",
	ServerPort:     "8080",
}

var ServerConfig = Config{
	ServerAddress: "127.0.0.1",
	ServerPort:    "8080",
}
