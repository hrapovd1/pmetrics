package storage

import (
	"bufio"
	"os"
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
	conf2, err := config.NewServer(config.Flags{})
	require.NoError(t, err)
	tests := []struct {
		name string
		cfg  config.Config
		want bool
	}{
		{name: "With error", cfg: conf, want: true},
		{name: "NO erro", cfg: *conf2, want: false},
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

func TestFileStorage_Restore(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "devops*.json")
	result := make(map[string]interface{})
	fs := FileStorage{
		file:   tmpFile,
		writer: bufio.NewWriter(tmpFile),
		buff:   result,
	}
	want := map[string]interface{}{
		"M1": int64(4),
		"M2": float64(3.9),
	}
	data := `[{"id":"M1","type":"counter","delta":4},{"id":"M2","type":"gauge","value":3.9}]`
	_, err := fs.writer.Write([]byte(data))
	require.NoError(t, err)
	require.NoError(t, fs.writer.Flush())
	require.NoError(t, fs.Restore())
	assert.Equal(t, want, fs.buff)
}
