package config

import (
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
				RetryCount:     3,
				ServerAddress:  "localhost:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			err := cfg.NewAgent()
			require.NoError(t, err)
			assert.Equal(t, tt.fields, cfg)
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
				ServerAddress: "localhost:8080",
				StoreInterval: 300 * time.Second,
				StoreFile:     "/tmp/devops-metrics-db.json",
				IsRestore:     true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			err := cfg.NewServer()
			require.NoError(t, err)
			assert.Equal(t, tt.fields, cfg)
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
