package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

type environ struct {
	PollInterval   string `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval string `env:"REPORT_INTERVAL" envDefault:"10s"`
	StoreInterval  string `env:"STORE_INTERVAL" envDefault:"300s"`
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreFile      string `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	IsRestore      bool   `env:"RESTORE" envDefault:"true"`
}

type Flags struct {
	ADDRESS         string
	POLL_INTERVAL   string
	REPORT_INTERVAL string
	RESTORE         bool
	STORE_FILE      string
	STORE_INTERVAL  string
}

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	RetryCount     int
	ServerAddress  string
	StoreInterval  time.Duration
	StoreFile      string
	IsRestore      bool
	tagsDefault    map[string]bool
}

func (cfg *Config) getTags(tag string, value interface{}, isDefault bool) {
	cfg.tagsDefault[tag] = isDefault
}

func GetServerFlags() Flags {
	flags := Flags{}
	flag.StringVar(&flags.ADDRESS, "a", "", "Address of server, for example: 0.0.0.0:8000")
	flag.BoolVar(&flags.RESTORE, "r", true, "Restore last data from file, true/false")
	flag.StringVar(&flags.STORE_INTERVAL, "i", "", "Interval of write to file in seconds, for example: 30s")
	flag.StringVar(&flags.STORE_FILE, "f", "", "File where server keep data, for example: /tmp/server.json")
	flag.Parse()
	return flags
}

func GetAgentFlags() Flags {
	flags := Flags{}
	flag.StringVar(&flags.ADDRESS, "a", "", "Address of server, for example: 0.0.0.0:8000")
	flag.StringVar(&flags.REPORT_INTERVAL, "r", "", "Interval of sent data to server in seconds, for example: 30s")
	flag.StringVar(&flags.POLL_INTERVAL, "p", "", "Interval of query metrics in seconds, for example: 30s")
	flag.Parse()
	return flags
}

func NewAgent(flags Flags) (*Config, error) {
	var cfg Config
	cfg.tagsDefault = make(map[string]bool)
	var err error
	var envs environ
	if err = env.Parse(&envs, env.Options{OnSet: cfg.getTags}); err != nil {
		return nil, err
	}
	// Определяю интервал опроса метрик
	var pollInterval string
	if flags.POLL_INTERVAL != "" && cfg.tagsDefault["POLL_INTERVAL"] {
		pollInterval = flags.POLL_INTERVAL
	} else {
		pollInterval = envs.PollInterval
	}
	if cfg.PollInterval, err = parseInterval(pollInterval); err != nil {
		return nil, err
	}
	// Определяю интервал отправки метрик
	var reportInterval string
	if flags.REPORT_INTERVAL != "" && cfg.tagsDefault["REPORT_INTERVAL"] {
		reportInterval = flags.REPORT_INTERVAL
	} else {
		reportInterval = envs.ReportInterval
	}
	if cfg.ReportInterval, err = parseInterval(reportInterval); err != nil {
		return nil, err
	}
	// Определяю адрес сервера
	if flags.ADDRESS != "" && cfg.tagsDefault["ADDRESS"] {
		cfg.ServerAddress = flags.ADDRESS
	} else {
		cfg.ServerAddress = envs.Address
	}
	cfg.RetryCount = 3
	return &cfg, err
}

func NewServer(flags Flags) (*Config, error) {
	var err error
	var cfg Config
	cfg.tagsDefault = make(map[string]bool)
	var envs environ
	// Разбираю переменные среды и проверяю значение тегов на значение по умолчанию
	if err = env.Parse(&envs, env.Options{OnSet: cfg.getTags}); err != nil {
		return nil, err
	}
	// Определяю адрес сервера
	if flags.ADDRESS != "" && cfg.tagsDefault["ADDRESS"] {
		cfg.ServerAddress = flags.ADDRESS
	} else {
		cfg.ServerAddress = envs.Address
	}
	// Определяю интервал сохранения в файл
	var storeInterval string
	if flags.STORE_INTERVAL != "" && cfg.tagsDefault["STORE_INTERVAL"] {
		storeInterval = flags.STORE_INTERVAL
	} else {
		storeInterval = envs.StoreInterval
	}
	if cfg.StoreInterval, err = parseInterval(storeInterval); err != nil {
		return nil, err
	}
	// Определяю интервал отправки отчета, как значение при storeInterval == 0
	if cfg.ReportInterval, err = parseInterval(envs.ReportInterval); err != nil {
		return nil, err
	}
	// Определяю файл для хранения метрик
	if flags.STORE_FILE != "" && cfg.tagsDefault["STORE_FILE"] {
		cfg.StoreFile = flags.STORE_FILE
	} else {
		cfg.StoreFile = envs.StoreFile
	}
	// Определяю флаг восстановления метрик при запуске
	if cfg.tagsDefault["RESTORE"] {
		cfg.IsRestore = flags.RESTORE
	} else {
		cfg.IsRestore = envs.IsRestore
	}

	return &cfg, err
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
