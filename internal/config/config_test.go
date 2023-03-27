package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_NewAgent(t *testing.T) {
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
					"CRYPTO_KEY":      true,
					"KEY":             true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewAgentConf(Flags{})
			require.NoError(t, err)
			assert.Equal(t, tt.fields, *cfg)
		})
	}
}

func TestConfig_NewServer(t *testing.T) {
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
					"KEY":             true,
					"CRYPTO_KEY":      true,
					"POLL_INTERVAL":   true,
					"REPORT_INTERVAL": true,
					"RESTORE":         true,
					"STORE_FILE":      true,
					"STORE_INTERVAL":  true,
					"DATABASE_DSN":    true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewServerConf(Flags{})
			require.NoError(t, err)
			assert.Equal(t, tt.fields, *cfg)
		})
	}
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

func ExampleGetAgentFlags() {
	agentFlags := GetAgentFlags()

	agentConfig, _ := NewAgentConf(agentFlags)

	fmt.Println(agentConfig)
}

func ExampleGetServerFlags() {
	serverFlags := GetServerFlags()

	serverConfig, _ := NewServerConf(serverFlags)

	fmt.Println(serverConfig)
}
