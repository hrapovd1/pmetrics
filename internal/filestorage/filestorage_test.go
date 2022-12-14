package filestorage

import (
	"bufio"
	"context"
	"os"
	"testing"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileStorage(t *testing.T) {
	backConf := config.Config{ServerAddress: "", StoreFile: "", StoreInterval: 0, IsRestore: false}
	test := struct {
		backConf config.Config
		want     *FileStorage
	}{
		backConf: backConf,
		want:     &FileStorage{ctx: context.Background(), config: backConf, file: nil, writer: nil, buff: nil},
	}
	assert.Equal(t, test.want, NewFileStorage(context.Background(), test.backConf, map[string]interface{}{}))
}

func TestFileStorage_Close(t *testing.T) {
	var conf config.Config
	conf2, err := config.NewServerConf(config.Flags{})
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
			fileStorage := NewFileStorage(context.Background(), test.cfg, map[string]interface{}{})
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
	result := map[string]interface{}{
		"M1": int64(4),
		"M2": float64(3.9),
	}
	fs := FileStorage{
		ctx:    context.Background(),
		file:   tmpFile,
		writer: bufio.NewWriter(tmpFile),
		buff:   result,
	}
	want := map[string]interface{}{
		"M1": int64(4),
		"M2": float64(3.9),
	}
	//data := `[{"id":"M1","type":"counter","delta":4},{"id":"M2","type":"gauge","value":3.9}]`
	err := fs.Store()
	require.NoError(t, err)
	require.NoError(t, fs.writer.Flush())
	require.Error(t, fs.Restore())
	assert.Equal(t, want, fs.buff)
}
