package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentConf(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "*conf.json")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`{"}`)
	require.NoError(t, err)
	tests := []struct {
		name   string
		fields Config
	}{
		{
			name: "Agent config",
			fields: Config{
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				ServerAddress:  "localhost:8080",
				StoreInterval:  0,
				StoreFile:      "",
				IsRestore:      false,
				DatabaseDSN:    "",
				CryptoKey:      "",
				Key:            "",
				tagsDefault: map[string]bool{
					"ADDRESS":         true,
					"CONFIG":          true,
					"CRYPTO_KEY":      true,
					"KEY":             true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
					"TRUSTED_SUBNET":  true,
				},
			},
		},
		{
			name: "cfg from flags",
			fields: Config{
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				ServerAddress:  "localhost:8080",
				StoreInterval:  0,
				StoreFile:      "",
				IsRestore:      false,
				DatabaseDSN:    "",
				CryptoKey:      "",
				Key:            "",
				tagsDefault: map[string]bool{
					"ADDRESS":         true,
					"CONFIG":          true,
					"CRYPTO_KEY":      true,
					"KEY":             true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
					"TRUSTED_SUBNET":  true,
				},
			},
		},
		{
			name: "bad file config",
			fields: Config{
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				ServerAddress:  "localhost:8080",
				StoreInterval:  0,
				StoreFile:      "",
				IsRestore:      false,
				DatabaseDSN:    "",
				CryptoKey:      "",
				Key:            "",
				tagsDefault: map[string]bool{
					"ADDRESS":         true,
					"CONFIG":          false,
					"CRYPTO_KEY":      true,
					"KEY":             true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
					"TRUSTED_SUBNET":  true,
				},
			},
		},
		{
			name: "bad flag config",
			fields: Config{
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				ServerAddress:  "localhost:8080",
				StoreInterval:  0,
				StoreFile:      "",
				IsRestore:      false,
				DatabaseDSN:    "",
				CryptoKey:      "",
				Key:            "",
				tagsDefault: map[string]bool{
					"ADDRESS":         true,
					"CONFIG":          true,
					"CRYPTO_KEY":      true,
					"KEY":             true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
					"TRUSTED_SUBNET":  true,
				},
			},
		},
	}
	for _, tt := range tests {
		if strings.Contains(tt.name, "bad file") {
			t.Run(tt.name, func(t *testing.T) {
				defer os.Unsetenv("CONFIG")
				os.Setenv("CONFIG", tmpFile.Name())
				cfg, err := NewAgentConf(Flags{})
				require.Error(t, err)
				assert.Nil(t, cfg)
			})
		} else if strings.Contains(tt.name, "bad flag") {
			t.Run(tt.name, func(t *testing.T) {
				cfg, err := NewAgentConf(Flags{configFile: tmpFile.Name()})
				require.Error(t, err)
				assert.Nil(t, cfg)
			})
		} else if strings.Contains(tt.name, "cfg from") {
			t.Run(tt.name, func(t *testing.T) {
				cfg, err := NewAgentConf(Flags{
					pollInterval:   "2s",
					reportInterval: "10s",
					address:        "localhost:8080",
					storeInterval:  "0s",
					storeFile:      "",
					restore:        false,
					dbDSN:          "",
					cryptoKey:      "",
					key:            "",
				})
				require.NoError(t, err)
				assert.Equal(t, tt.fields, *cfg)
			})
		} else {
			t.Run(tt.name, func(t *testing.T) {
				cfg, err := NewAgentConf(Flags{})
				require.NoError(t, err)
				assert.Equal(t, tt.fields, *cfg)
			})
		}
	}
}

func TestNewServerConf(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "*conf.json")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`{"}`)
	require.NoError(t, err)
	tests := []struct {
		name   string
		fields Config
	}{
		{
			name: "Server config",
			fields: Config{
				ServerAddress:  "localhost:8080",
				ReportInterval: 10 * time.Second,
				StoreInterval:  300 * time.Second,
				StoreFile:      "/tmp/devops-metrics-db.json",
				IsRestore:      false,
				Key:            "",
				CryptoKey:      "",
				DatabaseDSN:    "",
				tagsDefault: map[string]bool{
					"ADDRESS":         true,
					"CONFIG":          true,
					"KEY":             true,
					"CRYPTO_KEY":      true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
					"TRUSTED_SUBNET":  true,
				},
			},
		},
		{
			name: "cfg from flags",
			fields: Config{
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				ServerAddress:  "localhost:8080",
				StoreInterval:  0,
				StoreFile:      "",
				IsRestore:      false,
				DatabaseDSN:    "",
				CryptoKey:      "",
				Key:            "",
				tagsDefault: map[string]bool{
					"ADDRESS":         true,
					"CONFIG":          true,
					"CRYPTO_KEY":      true,
					"KEY":             true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
					"TRUSTED_SUBNET":  true,
				},
			},
		},
		{
			name: "bad file config",
			fields: Config{
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				ServerAddress:  "localhost:8080",
				StoreInterval:  0,
				StoreFile:      "",
				IsRestore:      false,
				DatabaseDSN:    "",
				CryptoKey:      "",
				Key:            "",
				tagsDefault: map[string]bool{
					"ADDRESS":         true,
					"CONFIG":          false,
					"CRYPTO_KEY":      true,
					"KEY":             true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
					"TRUSTED_SUBNET":  true,
				},
			},
		},
		{
			name: "bad flag config",
			fields: Config{
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				ServerAddress:  "localhost:8080",
				StoreInterval:  0,
				StoreFile:      "",
				IsRestore:      false,
				DatabaseDSN:    "",
				CryptoKey:      "",
				Key:            "",
				tagsDefault: map[string]bool{
					"ADDRESS":         true,
					"CONFIG":          true,
					"CRYPTO_KEY":      true,
					"KEY":             true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
					"TRUSTED_SUBNET":  true,
				},
			},
		},
	}
	for _, tt := range tests {
		if strings.Contains(tt.name, "bad file") {
			t.Run(tt.name, func(t *testing.T) {
				defer os.Unsetenv("CONFIG")
				os.Setenv("CONFIG", tmpFile.Name())
				cfg, err := NewServerConf(Flags{})
				require.Error(t, err)
				assert.Nil(t, cfg)
			})
		} else if strings.Contains(tt.name, "bad flag") {
			t.Run(tt.name, func(t *testing.T) {
				cfg, err := NewServerConf(Flags{configFile: tmpFile.Name()})
				require.Error(t, err)
				assert.Nil(t, cfg)
			})
		} else if strings.Contains(tt.name, "cfg from") {
			t.Run(tt.name, func(t *testing.T) {
				cfg, err := NewAgentConf(Flags{
					pollInterval:   "2s",
					reportInterval: "10s",
					address:        "localhost:8080",
					storeInterval:  "0s",
					storeFile:      "",
					restore:        false,
					dbDSN:          "",
					cryptoKey:      "",
					key:            "",
				})
				require.NoError(t, err)
				assert.Equal(t, tt.fields, *cfg)
			})
		} else {
			t.Run(tt.name, func(t *testing.T) {
				cfg, err := NewServerConf(Flags{})
				require.NoError(t, err)
				assert.Equal(t, tt.fields, *cfg)
			})
		}
	}
}

func TestConfig_getTags(t *testing.T) {
	conf := Config{tagsDefault: make(map[string]bool)}
	conf.getTags("true", nil, true)
	conf.getTags("false", nil, false)
	assert.True(t, conf.tagsDefault["true"])
	assert.False(t, conf.tagsDefault["false"])
}

func Test_parseInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval string
		want     time.Duration
		wantErr  bool
	}{
		{
			name:     "test 2 second",
			interval: "2s",
			want:     2 * time.Second,
			wantErr:  false,
		},
		{
			name:     "test 2 minutes",
			interval: "2m",
			want:     2 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "wrong format",
			interval: "2h",
			want:     2 * time.Minute,
			wantErr:  true,
		},
		{
			name:     "wrong value",
			interval: "2hs",
			want:     2 * time.Second,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := parseInterval(tt.interval)
			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, val)
		})
	}
}

func Test_getRawJSONConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "*config.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tests := []struct {
		name     string
		dataSize int
		positive bool
	}{
		{name: "small file", dataSize: 2000, positive: true},
		{name: "big file", dataSize: 1, positive: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := tmpFile.WriteString(strings.Repeat(" ", test.dataSize))
			require.NoError(t, err)

			result, err := getRawJSONConfig(tmpFile.Name())
			if test.positive {

				require.NoError(t, err)
				assert.Equal(t, test.dataSize, len(result))

			} else {

				require.Error(t, err)
				assert.Equal(t, tmpFile.Name()+" too big.", err.Error())

			}

		})
	}
}

func TestConfig_setConfigFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "*config.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString(`{
    "address": "localhost:8080",
    "restore": true,
    "store_interval": "1s",
    "store_file": "/path/to/file.db"
	}`)
	require.NoError(t, err)

	want := Config{
		ServerAddress: "localhost:8080",
		IsRestore:     true,
		StoreInterval: time.Second * 1,
		StoreFile:     "/path/to/file.db",
	}

	t.Run("good", func(t *testing.T) {
		conf := Config{}
		require.NoError(t, conf.setConfigFromFile(tmpFile.Name()))
		assert.Equal(t, want, conf)
	})
	t.Run("bad", func(t *testing.T) {
		conf := Config{}
		require.Error(t, conf.setConfigFromFile("/tmp/ke79685"))
	})
}

func TestConfig_UnmarshalJSON(t *testing.T) {
	var conf Config
	tests := []struct {
		name     string
		data     []byte
		positive bool
	}{
		{
			name:     "right data",
			positive: true,
			data:     []byte(`{"poll_interval": "2s", "report_interval": "5s", "store_interval": "7s"}`),
		},
		{
			name:     "wrong data",
			positive: false,
			data:     []byte(`{"poll_interval: "2s", "report_interval": "5s", "store_interval": "7s"}`),
		},
		{
			name:     "wrong poll_interval",
			positive: false,
			data:     []byte(`{"poll_interval": "2", "report_interval": "5s", "store_interval": "7s"}`),
		},
		{
			name:     "wrong report_interval",
			positive: false,
			data:     []byte(`{"poll_interval": "2s", "report_interval": "5", "store_interval": "7s"}`),
		},
		{
			name:     "wrong store_interval",
			positive: false,
			data:     []byte(`{"poll_interval": "2s", "report_interval": "5s", "store_interval": "7"}`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := conf.UnmarshalJSON(test.data)
			if test.positive {
				require.NoError(t, err)
				assert.Equal(t, time.Second*2, conf.PollInterval)
				assert.Equal(t, time.Second*5, conf.ReportInterval)
				assert.Equal(t, time.Second*7, conf.StoreInterval)
			} else {
				require.Error(t, err)
			}

		})
	}
}

func TestConfig_valueExists(t *testing.T) {
	conf := Config{Key: "12345"}
	assert.True(t, conf.valueExists("Key"))
	assert.False(t, conf.valueExists("PollInterval"))
}

func ExampleGetAgentFlags() {
	agentFlags := GetAgentFlags()

	agentConfig, _ := NewAgentConf(agentFlags)

	fmt.Println(agentConfig)
}

func TestGetServerFlags(t *testing.T) {
	fmt.Printf("before: %v\n", os.Args)
	flags := GetServerFlags()
	fmt.Printf("after: %v\n", os.Args)
	assert.Equal(t,
		Flags{
			address: "", pollInterval: "", reportInterval: "",
			restore: true, storeFile: "", storeInterval: "",
			key: "", cryptoKey: "", dbDSN: "", configFile: "",
			trustedSubnet: ""},
		flags,
	)
}

// func TestGetAgentFlags(t *testing.T) {
// 	flags := GetAgentFlags()
// 	assert.Equal(t,
// 		Flags{
// 			address: "", pollInterval: "", reportInterval: "",
// 			restore: false, storeFile: "", storeInterval: "",
// 			key: "", cryptoKey: "", dbDSN: "", configFile: "",
// 			trustedSubnet: ""},
// 		flags,
// 	)
// }

func ExampleGetServerFlags() {
	serverFlags := GetServerFlags()

	serverConfig, _ := NewServerConf(serverFlags)

	fmt.Println(serverConfig)
}
