package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_gzipWriter_Write(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		Writer         io.Writer
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Positive",
			fields:  fields{Writer: &bytes.Buffer{}},
			args:    args{b: []byte(`Test string`)},
			want:    11,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := gzipWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				Writer:         tt.fields.Writer,
			}
			got, err := w.Write(tt.args.b)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetricsHandler_GzipMiddle(t *testing.T) {
	mh := MetricsHandler{}
	simpleHandl := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Test is OK."))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("without compress", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mh.GzipMiddle(http.HandlerFunc(simpleHandl)).ServeHTTP(rec, request)
		result := rec.Result()
		defer assert.Nil(t, result.Body.Close())
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, 11, len(body))
	})
	t.Run("with compress", func(t *testing.T) {
		request.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		mh.GzipMiddle(http.HandlerFunc(simpleHandl)).ServeHTTP(rec, request)
		result := rec.Result()
		defer assert.Nil(t, result.Body.Close())
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, 39, len(body))
	})

}

func TestMetricsHandler_DecryptMiddle(t *testing.T) {
	encyptBody := `{"data0":"EzzCAxNQHdI0ZEvSee8Og3ODT1tdMu9THUpHpZtnnFjklrkKMZ+858YlsJJr6mw59BtW9sD6XuPICpCNsK92zaYVE2GLrNHGSKxLJLgi+HnkLlcjA0FOIExCU/RPOQu+fgFMWIrSk6+yodawNtb6t9jYy7L7bm6AMwixUZvVKq13Oq3Qn3I3WRzoBWxZ2XxfTPGX1OMWtkCEaSWRnMKidHlqmyE469YxWbVE8fuCZEvfGfRqTBJ/Hn+fwE6IaNR16BZNsvymQYC6H+/ZudFxFi0AP7DttYOGkjF1hK5vJq9mEfXk6BdTfs9+CTwmTrg2fr9YbYgBNFsCoknvuxvViZTejF3ka/J4B0BBMAyjBUx1U+3aOiEQkHtTkO5PzCIiuCshA+du0XMcSvOOIuZvC56LRLCOF1DyLs9mR0V0vmHRykT/KDF32+N5EllS324aK4rssoR8AwVPWKIaNQonM6sPK3PxOAJVjY4vVl6xPXG2GOw2oMjMdH84yurw3IA06pllC46U3Z6okjxC/3dEK29Otji9xj4SD7b4Q1So5qRGsYKpgwKhZcdgwLbE8K7o+Wc2DnzEi+NvppsyJuV7D81jlND9Vb2m7vV3/jvkRcjXrHC7QQLGcAPd8KyGrGU4LDZVls7ngXy/RCnY2mTjF7iyWP5/BcjiNZlTPZVrhnI=","data1":"DpNSK2X+M0E0qPpX/w3iSOAFRVpM+H3SmvXwKE7+uOkkPSUh0EZc4iCjF7fZHj8LoPFFmROPvAn1jW9eAfapIyniCGbf7CVgNbFlq88E+hDLz80LZvxZFhwEU3NJxOyPZ5sjt/cvXcFujdwTxkMO1RXKyjE="}`

	tmpFile, _ := os.CreateTemp("", "*privkey.pem")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`-----BEGIN PRIVATE KEY-----
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
58qM7mUsPO5Imn8pJSz+Vsj3YTsJXoaKBg==
-----END PRIVATE KEY-----`)
	require.NoError(t, err)

	mh := MetricsHandler{
		Storage: storage.NewMemStorage(),
		logger:  log.New(os.Stderr, "test", log.Default().Flags()),
		Config: config.Config{
			CryptoKey: tmpFile.Name(),
		},
	}

	alloc1 := float64(-4.5)
	count1 := int64(5)

	want := []types.Metric{
		{ID: "Alloc1", MType: "gauge", Value: &alloc1},
		{ID: "Count1", MType: "counter", Delta: &count1},
	}

	var simpleData []types.Metric
	simpleHandl := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &simpleData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(""))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	t.Run("decrypt success", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(encyptBody))
		request.Header.Set("Encrypt-Type", "1")

		rec := httptest.NewRecorder()
		mh.DecryptMiddle(http.HandlerFunc(simpleHandl)).ServeHTTP(rec, request)
		result := rec.Result()
		defer assert.Nil(t, result.Body.Close())
		_, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		assert.Equal(t, want, simpleData)
	})

}

func Test_decryptData(t *testing.T) {
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

	privKey, err := getPrivKey(tmpFile.Name(), log.Default())
	require.NoError(t, err)

	result, err := decryptData(encyptData, privKey)
	require.NoError(t, err)

	assert.Equal(t, "Test data.", string(result))

}

func Test_checkAddr(t *testing.T) {
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
				res, err := checkAddr(addr, test.subNet)
				require.NoError(t, err)
				assert.True(t, res)
			} else {
				res, err := checkAddr(addr, test.subNet)
				if err == nil {
					assert.False(t, res)
				} else {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestMetricsHandler_CheckAgentNetMiddle(t *testing.T) {
	var req []byte
	simpleHandl := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var err error
		req, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(""))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	t.Run("with header", func(t *testing.T) {
		clientReq := []byte("test")
		mh := MetricsHandler{
			Config: config.Config{TrustedSubnet: "192.168.1.0/24"},
			logger: log.Default(),
		}
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(clientReq))
		request.Header.Set("X-Real-IP", "192.168.1.1")

		rec := httptest.NewRecorder()
		mh.CheckAgentNetMiddle(http.HandlerFunc(simpleHandl)).ServeHTTP(rec, request)
		assert.Equal(t, clientReq, req)
	})
	t.Run("without header", func(t *testing.T) {
		clientReq := []byte("test1")
		mh := MetricsHandler{
			Config: config.Config{TrustedSubnet: "192.168.1.0/24"},
			logger: log.Default(),
		}
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(clientReq))

		rec := httptest.NewRecorder()
		mh.CheckAgentNetMiddle(http.HandlerFunc(simpleHandl)).ServeHTTP(rec, request)
		assert.Equal(t, http.StatusForbidden, rec.Result().StatusCode)
		assert.NotEqual(t, clientReq, req)
	})
	t.Run("wrong header", func(t *testing.T) {
		clientReq := []byte("test2")
		mh := MetricsHandler{
			Config: config.Config{TrustedSubnet: "192.168.1.0/24"},
			logger: log.Default(),
		}
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(clientReq))
		request.Header.Set("X-Real-IP", "192.168.0.1")

		rec := httptest.NewRecorder()
		mh.CheckAgentNetMiddle(http.HandlerFunc(simpleHandl)).ServeHTTP(rec, request)
		assert.Equal(t, http.StatusForbidden, rec.Result().StatusCode)
		assert.NotEqual(t, clientReq, req)
	})
	t.Run("without trusted subnet", func(t *testing.T) {
		clientReq := []byte("test3")
		mh := MetricsHandler{
			Config: config.Config{TrustedSubnet: ""},
			logger: log.Default(),
		}
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(clientReq))
		request.Header.Set("X-Real-IP", "192.168.1.1")

		rec := httptest.NewRecorder()
		mh.CheckAgentNetMiddle(http.HandlerFunc(simpleHandl)).ServeHTTP(rec, request)
		assert.Equal(t, clientReq, req)
	})
}
