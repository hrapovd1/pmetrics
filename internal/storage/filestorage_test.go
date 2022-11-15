package storage

import (
	"testing"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackend(t *testing.T) {
	backConf := config.Config{ServerAddress: "", StoreFile: "", StoreInterval: 0, IsRestore: false}
	test := struct {
		backConf config.Config
		want     FileStorage
	}{
		backConf: backConf,
		want:     FileStorage{config: backConf, file: nil, writer: nil, buff: nil},
	}
	assert.Equal(t, test.want, NewBackend(test.backConf))
}

func TestFileStorage_Close(t *testing.T) {
	var conf config.Config
	var conf2 config.Config
	require.NoError(t, conf2.NewServer())
	tests := []struct {
		name string
		cfg  config.Config
		want bool
	}{
		{name: "With error", cfg: conf, want: true},
		{name: "NO erro", cfg: conf2, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileStorage := NewBackend(test.cfg)
			if test.want {
				assert.Nil(t, fileStorage.Close())
			} else {
				assert.NoError(t, fileStorage.Close())
			}

		})
	}

}
