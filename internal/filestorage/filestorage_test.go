package filestorage

import (
	"bufio"
	"context"
	"io"
	"os"
	"testing"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileStorage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "*storage")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	require.NoError(t, tmpFile.Close())
	storage := NewFileStorage(
		config.Config{
			StoreFile:     tmpFile.Name(),
			StoreInterval: 0,
		},
		make(map[string]interface{}),
	)
	assert.Equal(t, tmpFile.Name(), storage.file.Name())
	storage.ms.Append(context.Background(), "M1", int64(34))
	storage.ms.Append(context.Background(), "M1", int64(34))
	assert.Equal(t, int64(68), storage.ms.Get(context.Background(), "M1"))
}

func TestFileStorage_Restore(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "devops*.json")
	defer os.Remove(tmpFile.Name())
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
	// data := `[{"id":"M1","type":"counter","delta":4},{"id":"M2","type":"gauge","value":3.9}]`
	err := fs.Store(ctx)
	require.NoError(t, err)
	require.NoError(t, fs.writer.Flush())
	require.Error(t, fs.Restore(ctx))
	assert.Equal(t, want, fs.GetAll(ctx))
}

func TestFileStorage_Append(t *testing.T) {
	buff := make(map[string]interface{})
	fs := FileStorage{
		ms: storage.NewMemStorage(storage.WithBuffer(buff)),
	}
	ctx := context.Background()
	result := map[string]interface{}{
		"M1": int64(4),
		"M2": int64(8),
	}
	fs.Append(ctx, "M1", int64(4))
	fs.Append(ctx, "M2", int64(4))
	fs.Append(ctx, "M2", int64(4))
	assert.Equal(t, result, buff)
}

func TestFileStorage_Rewrite(t *testing.T) {
	buff := make(map[string]interface{})
	fs := FileStorage{
		ms: storage.NewMemStorage(storage.WithBuffer(buff)),
	}
	ctx := context.Background()
	result := map[string]interface{}{
		"M1": float64(4),
		"M2": float64(8.7),
	}
	fs.Rewrite(ctx, "M1", float64(4))
	fs.Rewrite(ctx, "M2", float64(0))
	fs.Rewrite(ctx, "M2", float64(8.7))
	assert.Equal(t, result, buff)
}

func TestFileStorage_Ping(t *testing.T) {
	fs := FileStorage{}
	assert.False(t, fs.Ping(context.Background()))
}

func TestFileStorage_Store(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "devops*.json")
	defer os.Remove(tmpFile.Name())
	data := map[string]interface{}{
		"M1": int64(4),
	}
	fs := FileStorage{
		file:   tmpFile,
		writer: bufio.NewWriter(tmpFile),
		ms:     storage.NewMemStorage(storage.WithBuffer(data)),
	}
	result := make([]byte, 300)
	want := "[{\"id\":\"M1\",\"type\":\"counter\",\"delta\":4}]\n"
	ctx := context.Background()
	// data := `[{"id":"M1","type":"counter","delta":4},{"id":"M2","type":"gauge","value":3.9}]`
	require.NoError(t, fs.Store(ctx))
	require.NoError(t, fs.writer.Flush()) // Write to file
	n, err := tmpFile.ReadAt(result, 0)   // Read from File
	if err != io.EOF {
		require.NoError(t, err)
	}
	assert.Equal(t, want, string(result[:n]))
}

func TestFileStorage_Get(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "devops*.json")
	defer os.Remove(tmpFile.Name())
	result := map[string]interface{}{
		"M1": int64(4),
		"M2": float64(3.9),
	}
	fs := FileStorage{
		file:   tmpFile,
		writer: bufio.NewWriter(tmpFile),
		ms:     storage.NewMemStorage(storage.WithBuffer(result)),
	}
	want := int64(4)
	ctx := context.Background()
	// data := `[{"id":"M1","type":"counter","delta":4},{"id":"M2","type":"gauge","value":3.9}]`
	err := fs.Store(ctx)
	require.NoError(t, err)
	require.NoError(t, fs.writer.Flush())
	assert.Equal(t, want, fs.Get(ctx, "M1"))
}

func TestFileStorage_GetAll(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "devops*.json")
	defer os.Remove(tmpFile.Name())
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
	// data := `[{"id":"M1","type":"counter","delta":4},{"id":"M2","type":"gauge","value":3.9}]`
	err := fs.Store(ctx)
	require.NoError(t, err)
	require.NoError(t, fs.writer.Flush())
	assert.Equal(t, want, fs.GetAll(ctx))
}

func TestFileStorage_StoreAll(t *testing.T) {
	buff := make(map[string]interface{})
	fs := FileStorage{
		ms: storage.NewMemStorage(storage.WithBuffer(buff)),
	}
	var m1 = float64(8.7)
	var m2 = int64(4)
	metrics := []types.Metric{
		{ID: "M1", MType: "gauge", Value: &m1},
		{ID: "M2", MType: "counter", Delta: &m2},
	}
	ctx := context.Background()
	result := map[string]interface{}{
		"M1": float64(8.7),
		"M2": int64(4),
	}
	fs.StoreAll(ctx, &metrics)
	assert.Equal(t, result, buff)
}
