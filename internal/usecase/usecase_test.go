package usecase

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSONMetric(t *testing.T) {
	M1 := int64(5)
	M2 := float64(-4.65)
	tests := []struct {
		name string
		data types.Metric
		want string
	}{
		{
			name: "M1",
			data: types.Metric{ID: "M1", MType: "counter", Delta: &M1},
			want: "5",
		},
		{
			name: "M2",
			data: types.Metric{ID: "M2", MType: "gauge", Value: &M2},
			want: "-4.65",
		},
	}
	stor := make(map[string]interface{})
	ctx := context.Background()
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteJSONMetric(ctx, tt.data, locStorage)
			require.NoError(t, err)
			switch result := locStorage.Get(ctx, tt.data.ID).(type) {
			case int64:
				assert.Equal(t, tt.want, fmt.Sprint(result))
			case float64:
				assert.Equal(t, tt.want, fmt.Sprint(result))
			}
		})
	}
}

func TestGetJSONMetric(t *testing.T) {
	tests := []struct {
		name    string
		data    types.Metric
		withErr bool
		want    string
	}{
		{
			name:    "M1",
			data:    types.Metric{ID: "M1", MType: "counter"},
			withErr: false,
			want:    "5",
		},
		{
			name:    "M2",
			data:    types.Metric{ID: "M2", MType: "gauge"},
			withErr: false,
			want:    "-4.65",
		},
		{
			name:    "M3",
			data:    types.Metric{ID: "M3", MType: "type"},
			withErr: true,
			want:    "<nil>",
		},
	}
	stor := make(map[string]interface{})
	stor["M1"] = int64(5)
	stor["M2"] = float64(-4.65)
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GetJSONMetric(ctx, locStorage, &tt.data)
			if tt.withErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, fmt.Sprint(stor[tt.data.ID]))
		})
	}
}

func TestWriteMetric(t *testing.T) {
	tests := []struct {
		name       string
		path       []string
		metricName string
		want       string
	}{
		{
			name:       "M1",
			path:       []string{"", "update", "counter", "M1", "5"},
			metricName: "M1",
			want:       "5",
		},
		{
			name:       "M2",
			path:       []string{"", "update", "gauge", "M2", "0"},
			metricName: "M2",
			want:       "0",
		},
		{
			name:       "M1_1",
			path:       []string{"", "update", "counter", "M1", "3"},
			metricName: "M1",
			want:       "8",
		},
		{
			name:       "M2_1",
			path:       []string{"", "update", "gauge", "M2", "-3.3"},
			metricName: "M2",
			want:       "-3.3",
		},
	}
	stor := make(map[string]interface{})
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteMetric(ctx, tt.path, locStorage)
			require.NoError(t, err)
			switch result := locStorage.Get(ctx, tt.metricName).(type) {
			case int64:
				assert.Equal(t, tt.want, fmt.Sprint(result))
			case float64:
				assert.Equal(t, tt.want, fmt.Sprint(result))
			}
		})
	}
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		name       string
		path       []string
		metricName string
		withErr    bool
		want       string
	}{
		{
			name:       "M1",
			path:       []string{"", "value", "counter", "M1"},
			metricName: "M1",
			withErr:    false,
			want:       "5",
		},
		{
			name:       "M2",
			path:       []string{"", "update", "gauge", "M2"},
			metricName: "M2",
			withErr:    false,
			want:       "0",
		},
		{
			name:       "M1_1",
			path:       []string{"", "update", "simple", "M1"},
			metricName: "M1",
			withErr:    true,
			want:       "",
		},
		{
			name:       "M3",
			path:       []string{"", "update", "gauge", "M3"},
			metricName: "M3",
			withErr:    false,
			want:       "",
		},
	}
	stor := make(map[string]interface{})
	stor["M1"] = int64(5)
	stor["M2"] = float64(0)
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetric(ctx, locStorage, tt.path)
			if tt.withErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetTableMetrics(t *testing.T) {
	test := struct {
		name string
		want map[string]string
	}{
		name: "Check table",
		want: map[string]string{"M1": "5", "M2": "0"},
	}
	stor := make(map[string]interface{})
	stor["M1"] = int64(5)
	stor["M2"] = float64(0)
	locStorage := storage.NewMemStorage(storage.WithBuffer(stor))
	ctx := context.Background()
	result := GetTableMetrics(ctx, locStorage)

	t.Run(test.name, func(t *testing.T) {
		assert.True(t, cmp.Equal(test.want, result))
	})
}

func TestSignData(t *testing.T) {
	const key = "wersdjfl23.w3"
	counter := int64(34567)
	gauge := float64(8723.098)
	tests := []struct {
		name string
		data types.Metric
		want string
	}{
		{
			name: "counter value",
			data: types.Metric{ID: "test1", MType: "counter", Delta: &counter},
			want: "4003975ccfa11fdd45fc8ad03202a1f1dab466c3d118b2c7200e45de6f90da37",
		},
		{
			name: "gauge value",
			data: types.Metric{ID: "test2", MType: "gauge", Value: &gauge},
			want: "de0a02dd05ed708397ab730155bdf233ef415eebe8778c921b1aeff8333570e6",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, SignData(&test.data, key))
			assert.Equal(t, test.want, test.data.Hash)
		})
	}
}

func TestIsSignEqual(t *testing.T) {
	const key = "wersdjfl23.w3"
	value := int64(34567)
	tests := []struct {
		isPositive bool
		data       types.Metric
		want       bool
	}{
		{
			isPositive: true,
			data: types.Metric{
				ID:    "test",
				MType: "counter",
				Delta: &value,
				Hash:  "bf9f25fcb26f7df011969f39a83d81c7746a07ea400bf93e4758fd565fce32f8",
			},
			want: true,
		},
		{
			isPositive: false,
			data: types.Metric{
				ID:    "test",
				MType: "counter",
				Delta: &value,
				Hash:  "af9f25fcb26f7df011969f39a83d81c7746a07ea400bf93e4758fd565fce32f8",
			},
			want: true,
		},
	}
	for _, test := range tests {
		if test.isPositive {
			t.Run("positive", func(t *testing.T) {
				assert.True(t, IsSignEqual(test.data, key))
			})
		} else {
			t.Run("negative", func(t *testing.T) {
				assert.False(t, IsSignEqual(test.data, key))
			})

		}

	}

}

func TestDecryptData(t *testing.T) {
	encyptData := types.EncData{
		Data0: `6A+BQQ1R/fiFxmBOuLC9bVdP18CFaX5eVUI3vKX7xqU0xtg+VqT44lMecchX8uIshWLkQv9hHVtqdrEI38ET3QPEqxqQ7S8uDNRhAw90Om1EcbXI2mQQdOTk1wJY+od1cTj21BG48DyGpXoqTkLSgyoVhViSBGoi4Vx7desC4QIGJOkcXeYy3y50mbeyj/96Z5KuCWZiPA3KYpEmlpuxZGo1RwK2ykSIZl8zMlYxEcHgg/Wn8WN5mzyql+VH03U7Jgo+lkL503bD1y+YjyxmfkS4tk1ASP1k29yuEMjmpgJrN7cAGjYSP6PttsS+BG9TrSxKbUs6S/UJo13i5VptdTuCnlV35A/1C7ZG6DbX0c2Pu4nNXcusRaMkrx+NhMTadYgvDKNi5o4fQ8pKeJu2zyNxI+gjn1Tkk6wjeB0xowtrIC+mn7cyQcvf5f0J65pSgyc2DGiXbWSrW2QumTODE4NcPJlltVbKne7ytJaQtmZBzTdgKi+M45xwLlbR+jvtxRWRCq5QnbV1lEafJ0AQjgeJwut1jfFcm+FhtKgd338MQKRPJnZxsTzoxRMoHZxXbD2wuN/j5FD0Ygs+ABj+4W1POntzWymv8hn+fa+9J4cSyxqUReERh1Om++50cuaVpdNW4bPxbygVmNXn2a7TzXDUx2LvTXkw74zj0vdthIM=`,
		Data:  `NoRmzDXCbCeyTLn4yt1FJrymudnW6+RGi5ByzdzKKOCB8wzKlEo=`,
	}

	tmpFile, _ := os.CreateTemp("", "*privkey.pem")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`-----BEGIN PRIVATE KEY-----
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
58qM7mUsPO5Imn8pJSz+Vsj3YTsJXoaKBg==
-----END PRIVATE KEY-----`)
	require.NoError(t, err)

	privKey, err := GetPrivKey(tmpFile.Name(), log.Default())
	require.NoError(t, err)

	symmKey, err := DecryptKey(encyptData.Data0, privKey)
	require.NoError(t, err)

	result, err := DecryptData(encyptData.Data, symmKey)
	require.NoError(t, err)

	assert.Equal(t, "Test data.", string(result))

}

func TestCheckAddr(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		subNet   string
		positive bool
	}{
		{"correct", "192.168.0.1", "192.168.0.0/16", true},
		{"wrong subNet", "192.168.0.1", "192.168.1.0/24", false},
		{"wrong subNet address", "192.168.0.1", "192.168.10/24", false},
		{"wrong subNet mask", "192.168.0.1", "192.168.1.0/33", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addr := net.ParseIP(test.address)
			if test.positive {
				res, err := CheckAddr(addr, test.subNet)
				require.NoError(t, err)
				assert.True(t, res)
			} else {
				res, err := CheckAddr(addr, test.subNet)
				if err == nil {
					assert.False(t, res)
				} else {
					assert.Error(t, err)
				}
			}
		})
	}
}
