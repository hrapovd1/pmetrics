// Модуль config определяет типы и методы для формирования
// конфигурации приложения через флаги и переменные среды.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

// environ содержит значения переменных среды
type environ struct {
	PollInterval   string `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval string `env:"REPORT_INTERVAL" envDefault:"10s"`
	StoreInterval  string `env:"STORE_INTERVAL" envDefault:"300s"`
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreFile      string `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	IsRestore      bool   `env:"RESTORE" envDefault:"true"`
	Key            string `env:"KEY" envDefault:""`
	CryptoKey      string `env:"CRYPTO_KEY" envDefault:""`
	DatabaseDSN    string `env:"DATABASE_DSN" envDefault:""`
	ConfigFile     string `env:"CONFIG" envDefault:""`
	TrustedSubnet  string `env:"TRUSTED_SUBNET" envDefault:""`
}

// Config тип итоговой конфигурации агента или сервера
type Config struct {
	PollInterval   time.Duration   `json:"poll_interval,omitempty"`
	ReportInterval time.Duration   `json:"report_interval,omitempty"`
	ServerAddress  string          `json:"address,omitempty"`
	StoreInterval  time.Duration   `json:"store_interval,omitempty"`
	StoreFile      string          `json:"store_file,omitempty"`
	IsRestore      bool            `json:"restore,omitempty"`
	Key            string          `json:"key,omitempty"`
	CryptoKey      string          `json:"crypto_key,omitempty"`
	DatabaseDSN    string          `json:"database_dsn,omitempty"`
	TrustedSubnet  string          `json:"trusted_subnet,omitempty"`
	tagsDefault    map[string]bool `json:"-"`
}

// NewAgentConf генерирует рабочую конфигурацию агента
func NewAgentConf(flags Flags) (*Config, error) {
	var cfg Config
	var fileCfg Config
	cfg.tagsDefault = make(map[string]bool)
	var err error
	var envs environ
	if err = env.Parse(&envs, env.Options{OnSet: cfg.getTags}); err != nil {
		return nil, err
	}
	// Определяю файл конфигурации и использую как
	// 3й источник конфигурации
	if !cfg.tagsDefault["CONFIG"] {
		if err := fileCfg.setConfigFromFile(envs.ConfigFile); err != nil {
			return nil, err
		}
	}
	if flags.configFile != "" {
		if err := fileCfg.setConfigFromFile(flags.configFile); err != nil {
			return nil, err
		}
	}
	// Определяю интервал опроса метрик
	var pollInterval string
	if flags.pollInterval != "" && cfg.tagsDefault["POLL_INTERVAL"] {
		pollInterval = flags.pollInterval
	} else {
		pollInterval = envs.PollInterval
	}
	if flags.pollInterval == "" && cfg.tagsDefault["POLL_INTERVAL"] && fileCfg.valueExists("PollInterval") {
		cfg.PollInterval = fileCfg.PollInterval
	} else {
		if cfg.PollInterval, err = parseInterval(pollInterval); err != nil {
			return nil, err
		}
	}
	// Определяю интервал отправки метрик
	var reportInterval string
	if flags.reportInterval != "" && cfg.tagsDefault["REPORT_INTERVAL"] {
		reportInterval = flags.reportInterval
	} else {
		reportInterval = envs.ReportInterval
	}
	if flags.reportInterval == "" && cfg.tagsDefault["REPORT_INTERVAL"] && fileCfg.valueExists("ReportInterval") {
		cfg.ReportInterval = fileCfg.ReportInterval
	} else {
		if cfg.ReportInterval, err = parseInterval(reportInterval); err != nil {
			return nil, err
		}
	}
	// Определяю адрес сервера
	if flags.address != "" && cfg.tagsDefault["ADDRESS"] {
		cfg.ServerAddress = flags.address
	} else {
		cfg.ServerAddress = envs.Address
	}
	if flags.address == "" && cfg.tagsDefault["ADDRESS"] && fileCfg.valueExists("ServerAddress") {
		cfg.ServerAddress = fileCfg.ServerAddress
	}
	// Определяю ключ для подписи метрик
	if flags.key != "" && cfg.tagsDefault["KEY"] {
		cfg.Key = flags.key
	} else {
		cfg.Key = envs.Key
	}
	if flags.key == "" && cfg.tagsDefault["KEY"] && fileCfg.valueExists("Key") {
		cfg.Key = fileCfg.Key
	}
	// Определяю файл с публичным ключом шифрования
	if flags.cryptoKey != "" && cfg.tagsDefault["CRYPTO_KEY"] {
		cfg.CryptoKey = flags.cryptoKey
	} else {
		cfg.CryptoKey = envs.CryptoKey
	}
	if flags.cryptoKey == "" && cfg.tagsDefault["CRYPTO_KEY"] && fileCfg.valueExists("CryptoKey") {
		cfg.CryptoKey = fileCfg.CryptoKey
	}
	// Определяю доверенную подсеть для заголовка X-Real-IP
	if cfg.tagsDefault["TRUSTED_SUBNET"] {
		cfg.TrustedSubnet = flags.trustedSubnet
	} else {
		cfg.TrustedSubnet = envs.TrustedSubnet
	}
	if flags.trustedSubnet == "" && cfg.tagsDefault["TRUSTED_SUBNET"] && fileCfg.valueExists("TrustedSubnet") {
		cfg.TrustedSubnet = fileCfg.TrustedSubnet
	}

	return &cfg, err
}

// NewServerConf генерирует рабочую конфигурацию сервера
func NewServerConf(flags Flags) (*Config, error) {
	var err error
	var cfg Config
	var fileCfg Config
	cfg.tagsDefault = make(map[string]bool)
	var envs environ
	// Разбираю переменные среды и проверяю значение тегов на значение по умолчанию
	if err = env.Parse(&envs, env.Options{OnSet: cfg.getTags}); err != nil {
		return nil, err
	}
	// Определяю файл конфигурации и использую как
	// 3й источник конфигурации
	if !cfg.tagsDefault["CONFIG"] {
		if err := fileCfg.setConfigFromFile(envs.ConfigFile); err != nil {
			return nil, err
		}
	}
	if flags.configFile != "" {
		if err := fileCfg.setConfigFromFile(flags.configFile); err != nil {
			return nil, err
		}
	}
	// Определяю адрес сервера
	if flags.address != "" && cfg.tagsDefault["ADDRESS"] {
		cfg.ServerAddress = flags.address
	} else {
		cfg.ServerAddress = envs.Address
	}
	if flags.address == "" && cfg.tagsDefault["ADDRESS"] && fileCfg.valueExists("ServerAddress") {
		cfg.ServerAddress = fileCfg.ServerAddress
	}
	// Определяю интервал сохранения в файл
	var storeInterval string
	if flags.storeInterval != "" && cfg.tagsDefault["STORE_INTERVAL"] {
		storeInterval = flags.storeInterval
	} else {
		storeInterval = envs.StoreInterval
	}
	if flags.storeInterval == "" && cfg.tagsDefault["STORE_INTERVAL"] && fileCfg.valueExists("StoreInterval") {
		cfg.StoreInterval = fileCfg.StoreInterval
	} else {
		if cfg.StoreInterval, err = parseInterval(storeInterval); err != nil {
			return nil, err
		}
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
	if flags.storeFile == "" && cfg.tagsDefault["STORE_FILE"] && fileCfg.valueExists("StoreFile") {
		cfg.StoreFile = fileCfg.StoreFile
	}
	// Определяю флаг восстановления метрик при запуске
	if cfg.tagsDefault["RESTORE"] {
		cfg.IsRestore = flags.restore
	} else {
		cfg.IsRestore = envs.IsRestore
	}
	if !flags.restore && cfg.tagsDefault["RESTORE"] && fileCfg.valueExists("IsRestore") {
		cfg.IsRestore = fileCfg.IsRestore
	}
	// Определяю ключ для подписи метрик
	if cfg.tagsDefault["KEY"] {
		cfg.Key = flags.key
	} else {
		cfg.Key = envs.Key
	}
	if flags.key == "" && cfg.tagsDefault["KEY"] && fileCfg.valueExists("Key") {
		cfg.Key = fileCfg.Key
	}
	// Определяю файл с приватным ключом шифрования
	if cfg.tagsDefault["CRYPTO_KEY"] {
		cfg.CryptoKey = flags.cryptoKey
	} else {
		cfg.CryptoKey = envs.CryptoKey
	}
	if flags.cryptoKey == "" && cfg.tagsDefault["CRYPTO_KEY"] && fileCfg.valueExists("CryptoKey") {
		cfg.CryptoKey = fileCfg.CryptoKey
	}
	// Определяю подключение к БД
	if cfg.tagsDefault["DATABASE_DSN"] {
		cfg.DatabaseDSN = flags.dbDSN
	} else {
		cfg.DatabaseDSN = envs.DatabaseDSN
	}
	if flags.dbDSN == "" && cfg.tagsDefault["DATABASE_DSN"] && fileCfg.valueExists("DatabaseDSN") {
		cfg.DatabaseDSN = fileCfg.DatabaseDSN
	}
	// Определяю доверенную подсеть
	if cfg.tagsDefault["TRUSTED_SUBNET"] {
		cfg.TrustedSubnet = flags.trustedSubnet
	} else {
		cfg.TrustedSubnet = envs.TrustedSubnet
	}
	if flags.trustedSubnet == "" && cfg.tagsDefault["TRUSTED_SUBNET"] && fileCfg.valueExists("TrustedSubnet") {
		cfg.TrustedSubnet = fileCfg.TrustedSubnet
	}

	return &cfg, err
}

// getTags проверка и отметка значений переменных среды что они по умолчанию или нет
func (cfg *Config) getTags(tag string, value interface{}, isDefault bool) {
	cfg.tagsDefault[tag] = isDefault
}

// setConfigFromFile устанавливает параметры конфигурации из файла в формате JSON
func (cfg *Config) setConfigFromFile(cFile string) error {
	rawJSON, err := getRawJSONConfig(cFile)
	if err != nil {
		return err
	}
	if !json.Valid(rawJSON) {
		return errors.New("JSON from " + cFile + " NOT valid.")
	}
	conf := Config{}
	if err := json.Unmarshal(rawJSON, &conf); err != nil {
		return err
	}
	*cfg = conf
	return nil
}

func (cfg *Config) UnmarshalJSON(data []byte) error {
	type ConfigAlias Config

	aliasValue := &struct {
		*ConfigAlias

		PollInterval   string `json:"poll_interval,omitempty"`
		ReportInterval string `json:"report_interval,omitempty"`
		StoreInterval  string `json:"store_interval,omitempty"`
	}{
		ConfigAlias: (*ConfigAlias)(cfg),
	}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}
	if aliasValue.PollInterval != "" {
		pollInterval, err := parseInterval(aliasValue.PollInterval)
		if err != nil {
			return err
		}
		cfg.PollInterval = pollInterval
	}
	if aliasValue.ReportInterval != "" {
		reportInterval, err := parseInterval(aliasValue.ReportInterval)
		if err != nil {
			return err
		}
		cfg.ReportInterval = reportInterval
	}
	if aliasValue.StoreInterval != "" {
		storeInterval, err := parseInterval(aliasValue.StoreInterval)
		if err != nil {
			return err
		}
		cfg.StoreInterval = storeInterval
	}
	return nil
}

func (cfg *Config) valueExists(val string) bool {
	rCfg := reflect.ValueOf(*cfg)
	return !rCfg.FieldByName(val).IsZero()

}

func getRawJSONConfig(fName string) ([]byte, error) {
	fileStat, err := os.Stat(fName)
	if err != nil {
		return nil, err
	}
	if fileStat.Size() > 2000 {
		return nil, errors.New(fName + " too big.")
	}
	rawJSON := make([]byte, 2000)
	cf, err := os.Open(fName)
	if err != nil {
		return nil, err
	}
	n, err := cf.Read(rawJSON)
	if err != nil {
		return nil, err
	}
	return rawJSON[:n], nil
}

// Flags содержит значения флагов переданные при запуске
type Flags struct {
	address        string
	pollInterval   string
	reportInterval string
	restore        bool
	storeFile      string
	storeInterval  string
	key            string
	cryptoKey      string
	dbDSN          string
	configFile     string
	trustedSubnet  string
}

// GetServerFlags - считывае флаги сервера
func GetServerFlags() Flags {
	flags := Flags{}
	flag.StringVar(&flags.address, "a", "", "Address of server, for example: 0.0.0.0:8000")
	flag.BoolVar(&flags.restore, "r", true, "Restore last data from file, true/false")
	flag.StringVar(&flags.storeInterval, "i", "", "Interval of write to file in seconds, for example: 30s")
	flag.StringVar(&flags.storeFile, "f", "", "File where server keep data, for example: /tmp/server.json")
	flag.StringVar(&flags.key, "k", "", "Key for sign hash sum, if ommited data will sent without sign")
	flag.StringVar(&flags.cryptoKey, "crypto-key", "", "Path to file with private rsa key for decrypt agent's messages")
	flag.StringVar(&flags.dbDSN, "d", "", "Database connect source, for example: postgres://username:password@localhost:5432/database_name")
	flag.StringVar(&flags.configFile, "c", "", "(or -config) Path to config file in JSON format")
	flag.StringVar(&flags.configFile, "config", "", "(or -c) Path to config file in JSON format")
	flag.StringVar(&flags.trustedSubnet, "t", "", "Trusted subnet from agent is sending data, for example: 192.168.0.0/24")
	flag.Parse()
	return flags
}

// GetAgentFlags - считывает флаги агента
func GetAgentFlags() Flags {
	flags := Flags{}
	flag.StringVar(&flags.address, "a", "", "Address of server, for example: 0.0.0.0:8000")
	flag.StringVar(&flags.reportInterval, "r", "", "Interval of sent data to server in seconds, for example: 30s")
	flag.StringVar(&flags.pollInterval, "p", "", "Interval of query metrics in seconds, for example: 30s")
	flag.StringVar(&flags.key, "k", "", "Key for sign hash sum, if ommited data will sent without sign")
	flag.StringVar(&flags.cryptoKey, "crypto-key", "", "Path to file with public rsa key for encrypt agent's messages")
	flag.StringVar(&flags.configFile, "c", "", "(or -config) Path to config file in JSON format")
	flag.StringVar(&flags.configFile, "config", "", "(or -c) Path to config file in JSON format")
	flag.StringVar(&flags.trustedSubnet, "t", "", "Local agent address for X-Real-IP header, for example: 192.168.0.2")
	flag.Parse()
	return flags
}

// parseInterval преобразует строку интервала в time.Durations
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
