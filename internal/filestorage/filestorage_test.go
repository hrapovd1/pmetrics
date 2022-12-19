package filestorage

import (
	"bufio"
	"context"
	"os"
	"testing"

	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_Restore(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "devops*.json")
	result := map[string]interface{}{
		"M1": int64(4),
		"M2": float64(3.9),
	}
	fs := FileStorage{
		file:   tmpFile,
		writer: bufio.NewWriter(tmpFile),
		ms:     storage.NewMemStorage(storage.WithBuffer(result)),
	}
	want := map[string]interface{}{
		"M1": int64(4),
		"M2": float64(3.9),
	}
	ctx := context.Background()
	//data := `[{"id":"M1","type":"counter","delta":4},{"id":"M2","type":"gauge","value":3.9}]`
	err := fs.Store(ctx)
	require.NoError(t, err)
	require.NoError(t, fs.writer.Flush())
	require.Error(t, fs.Restore(ctx))
	assert.Equal(t, want, fs.GetAll(ctx))
}
