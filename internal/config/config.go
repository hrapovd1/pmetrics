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
	Key            string `env:"KEY" envDefault:""`
	DatabaseDSN    string `env:"DATABASE_DSN" envDefault:""`
}

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	RetryCount     int
	ServerAddress  string
	StoreInterval  time.Duration
	StoreFile      string
	IsRestore      bool
	Key            string
	DatabaseDSN    string
	tagsDefault    map[string]bool
}

func NewAgentConf(flags Flags) (*Config, error) {
	var cfg Config
	cfg.tagsDefault = make(map[string]bool)
	var err error
	var envs environ
	if err = env.Parse(&envs, env.Options{OnSet: cfg.getTags}); err != nil {
		return nil, err
	}
	// Определяю интервал опроса метрик
	var pollInterval string
	if flags.pollInterval != "" && cfg.tagsDefault["POLL_INTERVAL"] {
		pollInterval = flags.pollInterval
	} else {
		pollInterval = envs.PollInterval
	}
	if cfg.PollInterval, err = parseInterval(pollInterval); err != nil {
		return nil, err
	}
	// Определяю интервал отправки метрик
	var reportInterval string
	if flags.reportInterval != "" && cfg.tagsDefault["REPORT_INTERVAL"] {
		reportInterval = flags.reportInterval
	} else {
		reportInterval = envs.ReportInterval
	}
	if cfg.ReportInterval, err = parseInterval(reportInterval); err != nil {
		return nil, err
	}
	// Определяю адрес сервера
	if flags.address != "" && cfg.tagsDefault["ADDRESS"] {
		cfg.ServerAddress = flags.address
	} else {
		cfg.ServerAddress = envs.Address
	}
	// Определяю ключ для подписи метрик
	if cfg.tagsDefault["KEY"] {
		cfg.Key = flags.key
	} else {
		cfg.Key = envs.Key
	}
	return &cfg, err
}

func NewServerConf(flags Flags) (*Config, error) {
	var err error
	var cfg Config
	cfg.tagsDefault = make(map[string]bool)
	var envs environ
	// Разбираю переменные среды и проверяю значение тегов на значение по умолчанию
	if err = env.Parse(&envs, env.Options{OnSet: cfg.getTags}); err != nil {
		return nil, err
	}
	// Определяю адрес сервера
	if flags.address != "" && cfg.tagsDefault["ADDRESS"] {
		cfg.ServerAddress = flags.address
	} else {
		cfg.ServerAddress = envs.Address
	}
	// Определяю интервал сохранения в файл
	var storeInterval string
	if flags.storeInterval != "" && cfg.tagsDefault["STORE_INTERVAL"] {
		storeInterval = flags.storeInterval
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
	if flags.storeFile != "" && cfg.tagsDefault["STORE_FILE"] {
		cfg.StoreFile = flags.storeFile
	} else {
		cfg.StoreFile = envs.StoreFile
	}
	// Определяю флаг восстановления метрик при запуске
	if cfg.tagsDefault["RESTORE"] {
		cfg.IsRestore = flags.restore
	} else {
		cfg.IsRestore = envs.IsRestore
	}
	// Определяю ключ для подписи метрик
	if cfg.tagsDefault["KEY"] {
		cfg.Key = flags.key
	} else {
		cfg.Key = envs.Key
	}
	// Определяю подключение к БД
	if cfg.tagsDefault["DATABASE_DSN"] {
		cfg.DatabaseDSN = flags.dbDSN
	} else {
		cfg.DatabaseDSN = envs.DatabaseDSN
	}

	return &cfg, err
}

func (cfg *Config) getTags(tag string, value interface{}, isDefault bool) {
	cfg.tagsDefault[tag] = isDefault
}

type Flags struct {
	address        string
	pollInterval   string
	reportInterval string
	restore        bool
	storeFile      string
	storeInterval  string
	key            string
	dbDSN          string
}

func GetServerFlags() Flags {
	flags := Flags{}
	flag.StringVar(&flags.address, "a", "", "Address of server, for example: 0.0.0.0:8000")
	flag.BoolVar(&flags.restore, "r", true, "Restore last data from file, true/false")
	flag.StringVar(&flags.storeInterval, "i", "", "Interval of write to file in seconds, for example: 30s")
	flag.StringVar(&flags.storeFile, "f", "", "File where server keep data, for example: /tmp/server.json")
	flag.StringVar(&flags.key, "k", "", "Key for sign hash sum, if ommited data will sent without sign")
	flag.StringVar(&flags.dbDSN, "d", "", "Database connect source, for example: postgres://username:password@localhost:5432/database_name")
	flag.Parse()
	return flags
}

func GetAgentFlags() Flags {
	flags := Flags{}
	flag.StringVar(&flags.address, "a", "", "Address of server, for example: 0.0.0.0:8000")
	flag.StringVar(&flags.reportInterval, "r", "", "Interval of sent data to server in seconds, for example: 30s")
	flag.StringVar(&flags.pollInterval, "p", "", "Interval of query metrics in seconds, for example: 30s")
	flag.StringVar(&flags.key, "k", "", "Key for sign hash sum, if ommited data will sent without sign")
	flag.Parse()
	return flags
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
