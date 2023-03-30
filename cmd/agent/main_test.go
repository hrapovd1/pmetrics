package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pollHwMetrics(t *testing.T) {
	testMetrics := make(map[string]interface{})
	test := struct {
		name string
		args mmetrics
		want []string
	}{
		name: "Check metric names",
		args: mmetrics{pollCounter: counter(0), mtrcs: testMetrics},
		want: []string{},
	}
	wg := &sync.WaitGroup{}
	vctx := context.WithValue(context.Background(), waitgrp("WG"), wg)
	t.Run(test.name, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(vctx, time.Microsecond*600)
		defer cancel()
		wg.Add(1)
		pollHwMetrics(ctx, &test.args, time.Microsecond*500, log.New(os.Stdout, "AGENT\t", log.Ldate|log.Ltime))
		for _, val := range test.want {
			_, ok := test.args.mtrcs[val]
			assert.True(t, ok)
		}
	})
}

func Test_pollMetrics(t *testing.T) {
	testMetrics := make(map[string]interface{})
	test := struct {
		name string
		args mmetrics
		want []string
	}{
		name: "Check metric names",
		args: mmetrics{pollCounter: counter(0), mtrcs: testMetrics},
		want: []string{
			"Alloc",
			"TotalAlloc",
			"Sys",
			"Lookups",
			"Mallocs",
			"Frees",
			"HeapAlloc",
			"HeapSys",
			"HeapIdle",
			"HeapInuse",
			"HeapReleased",
			"HeapObjects",
			"StackInuse",
			"StackSys",
			"MSpanInuse",
			"MSpanSys",
			"MCacheInuse",
			"MCacheSys",
			"BuckHashSys",
			"GCSys",
			"OtherSys",
			"NextGC",
			"LastGC",
			"PauseTotalNs",
			"NumGC",
			"NumForcedGC",
			"GCCPUFraction",
			"RandomValue",
		},
	}
	wg := &sync.WaitGroup{}
	vctx := context.WithValue(context.Background(), waitgrp("WG"), wg)
	t.Run(test.name, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(vctx, time.Microsecond*600)
		defer cancel()
		wg.Add(1)
		pollMetrics(ctx, &test.args, time.Microsecond*500)
		for _, val := range test.want {
			_, ok := test.args.mtrcs[val]
			assert.True(t, ok)
		}
	})
}

func Test_metricsJSON(t *testing.T) {
	tests := []struct {
		name    string
		metrics map[string]interface{}
		want    []byte
		wantn   []byte
	}{
		{
			name:    "Check gauge",
			metrics: map[string]interface{}{"M1": counter(345), "M2": gauge(63.689)},
			want:    []byte(`[{"id":"M1","type":"counter","delta":345},{"id":"M2","type":"gauge","value":63.689}]`),
			wantn:   []byte(`[{"id":"M2","type":"gauge","value":63.689},{"id":"M1","type":"counter","delta":345}]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metricsToJSON(tt.metrics, "")
			require.NoError(t, err)
			assert.True(
				t,
				reflect.DeepEqual(tt.want, got) || reflect.DeepEqual(tt.wantn, got),
			)
		})
	}
}

func Test_getPubKey(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "*key.pem")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA2ecMkWHx08rDlY3a6QOe
qGSRahGYCHS8REbWS9+DV/h6idHy2Fhgq4yqQomBo9teYb10tbz73kHdishz/1mo
9Tej/IUlD5jmqrhSLkF7av/ANKYe4x8Dir5aw8VN6eOWAbA6to9qUa3ZwtcK4Hi2
GBw9KhhWHgQuSB/gZb3DdXaGddZTQm1zLw/LopcB3/B2ZXUfxLR1KRUeanwltcfQ
IAArnob0Ls+WK4HVjRE95FRvj8RhAfo7QlR7C+og7gOG0rzjpSYpfu7jZB3BOc2v
pG8XlqWTa69jef8tXVBT1RigyutsQ0ejoQTqANHpK0Wq8N6NzrR5ov2ukq2wSGf6
JXnoVFfnFkhPyqpgtd9XgkkCGY0ohRNMCsAkuwgUv0UFyDTbbX2kbMv/LL6WLJ+2
THdikaQhbaof67NFblatDpcAHl3wGj468/MoatHyVjdKEpiZ6JxCtXCOwdIEe6+F
F4hvaIWGZfFYKzT4Swk1HtrROYuc3NOzHu/mSJaOGr+WI74ZndZt0D7nWwetw1GH
2lce85MCcCXNdhN1VbUCiNN+q9lgjxzMMQYR2zXmNeXfwEev9Ledq2V3zd3ZBvf9
eS4bI4nmheWxgw0t2J74Tc+juSo7vpXyqU/PUUKjPmIAIPlJWaETSTihl6P6v6ob
1foZVm9HaA3+uwVgGq2nh60CAwEAAQ==
-----END PUBLIC KEY-----`)
	require.NoError(t, err)

	key, err := getPubKey(tmpFile.Name(), &log.Logger{})
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, 512, key.Size())
}

func Test_dataToEncJSON(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "*key.pem")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA2ecMkWHx08rDlY3a6QOe
qGSRahGYCHS8REbWS9+DV/h6idHy2Fhgq4yqQomBo9teYb10tbz73kHdishz/1mo
9Tej/IUlD5jmqrhSLkF7av/ANKYe4x8Dir5aw8VN6eOWAbA6to9qUa3ZwtcK4Hi2
GBw9KhhWHgQuSB/gZb3DdXaGddZTQm1zLw/LopcB3/B2ZXUfxLR1KRUeanwltcfQ
IAArnob0Ls+WK4HVjRE95FRvj8RhAfo7QlR7C+og7gOG0rzjpSYpfu7jZB3BOc2v
pG8XlqWTa69jef8tXVBT1RigyutsQ0ejoQTqANHpK0Wq8N6NzrR5ov2ukq2wSGf6
JXnoVFfnFkhPyqpgtd9XgkkCGY0ohRNMCsAkuwgUv0UFyDTbbX2kbMv/LL6WLJ+2
THdikaQhbaof67NFblatDpcAHl3wGj468/MoatHyVjdKEpiZ6JxCtXCOwdIEe6+F
F4hvaIWGZfFYKzT4Swk1HtrROYuc3NOzHu/mSJaOGr+WI74ZndZt0D7nWwetw1GH
2lce85MCcCXNdhN1VbUCiNN+q9lgjxzMMQYR2zXmNeXfwEev9Ledq2V3zd3ZBvf9
eS4bI4nmheWxgw0t2J74Tc+juSo7vpXyqU/PUUKjPmIAIPlJWaETSTihl6P6v6ob
1foZVm9HaA3+uwVgGq2nh60CAwEAAQ==
-----END PUBLIC KEY-----`)
	require.NoError(t, err)

	data := []byte("Test data")

	key, err := getPubKey(tmpFile.Name(), &log.Logger{})
	require.NoError(t, err)

	encData, err := dataToEncJSON(key, data)
	require.NoError(t, err)
	assert.NotNil(t, encData)
	assert.Equal(t, 759, len(encData))

}

func Test_reportMetrics(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "*key.pem")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA2ecMkWHx08rDlY3a6QOe
qGSRahGYCHS8REbWS9+DV/h6idHy2Fhgq4yqQomBo9teYb10tbz73kHdishz/1mo
9Tej/IUlD5jmqrhSLkF7av/ANKYe4x8Dir5aw8VN6eOWAbA6to9qUa3ZwtcK4Hi2
GBw9KhhWHgQuSB/gZb3DdXaGddZTQm1zLw/LopcB3/B2ZXUfxLR1KRUeanwltcfQ
IAArnob0Ls+WK4HVjRE95FRvj8RhAfo7QlR7C+og7gOG0rzjpSYpfu7jZB3BOc2v
pG8XlqWTa69jef8tXVBT1RigyutsQ0ejoQTqANHpK0Wq8N6NzrR5ov2ukq2wSGf6
JXnoVFfnFkhPyqpgtd9XgkkCGY0ohRNMCsAkuwgUv0UFyDTbbX2kbMv/LL6WLJ+2
THdikaQhbaof67NFblatDpcAHl3wGj468/MoatHyVjdKEpiZ6JxCtXCOwdIEe6+F
F4hvaIWGZfFYKzT4Swk1HtrROYuc3NOzHu/mSJaOGr+WI74ZndZt0D7nWwetw1GH
2lce85MCcCXNdhN1VbUCiNN+q9lgjxzMMQYR2zXmNeXfwEev9Ledq2V3zd3ZBvf9
eS4bI4nmheWxgw0t2J74Tc+juSo7vpXyqU/PUUKjPmIAIPlJWaETSTihl6P6v6ob
1foZVm9HaA3+uwVgGq2nh60CAwEAAQ==
-----END PUBLIC KEY-----`)
	require.NoError(t, err)

	var simpleData []types.Metric
	var encData types.EncData
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
	encryptHandl := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &encData); err != nil {
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
	rClient := resty.New()
	wg := &sync.WaitGroup{}
	vctx := context.WithValue(context.Background(), waitgrp("WG"), wg)

	t.Run("simple data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(simpleHandl))
		defer ts.Close()
		srvAddr := strings.Split(ts.URL, "//")[1]

		metrcs := mmetrics{
			mtrcs: map[string]interface{}{
				"pollCounter": counter(345),
				"M1":          gauge(23.09),
			},
		}

		config := config.Config{
			CryptoKey:      "",
			Key:            "",
			ReportInterval: time.Millisecond * 2,
			ServerAddress:  srvAddr,
		}
		counter := int64(345)
		value := float64(23.09)
		wantData := []types.Metric{
			{ID: "pollCounter", MType: "counter", Delta: &counter},
			{ID: "M1", MType: "gauge", Value: &value},
		}

		wg.Add(1)
		ctx, cancel := context.WithTimeout(vctx, time.Millisecond*3)
		defer cancel()
		go reportMetrics(ctx, &metrcs, config, rClient, log.Default())

		<-ctx.Done()
		wg.Wait()
		assert.Contains(t, simpleData, wantData[0])
		assert.Contains(t, simpleData, wantData[1])
	})

	t.Run("encrypted data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(encryptHandl))
		defer ts.Close()
		srvAddr := strings.Split(ts.URL, "//")[1]
		metrcs := mmetrics{
			mtrcs: map[string]interface{}{
				"pollCounter": counter(345),
				"M1":          gauge(23.09),
			},
		}
		config := config.Config{
			CryptoKey:      tmpFile.Name(),
			Key:            "",
			ReportInterval: time.Millisecond * 2,
			ServerAddress:  srvAddr,
		}

		wg.Add(1)
		ctx, cancel := context.WithTimeout(vctx, time.Millisecond*3)
		defer cancel()
		go reportMetrics(ctx, &metrcs, config, rClient, log.Default())

		<-ctx.Done()
		wg.Wait()
		assert.Equal(t, 160, len(encData.Data))
		assert.Equal(t, 684, len(encData.Data0))

	})
}

func Test_genSymmKey(t *testing.T) {
	length := 24
	result, err := genSymmKey(length)
	require.NoError(t, err)
	require.Equal(t, length, len(result))
}
