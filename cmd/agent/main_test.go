package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/mygrpc"
	pb "github.com/hrapovd1/pmetrics/internal/proto"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	t.Run(test.name, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*600)
		defer cancel()
		wg.Add(1)
		pollHwMetrics(ctx, wg, &test.args, time.Microsecond*500, log.New(os.Stdout, "AGENT\t", log.Ldate|log.Ltime))
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
	t.Run(test.name, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*600)
		defer cancel()
		wg.Add(1)
		pollMetrics(ctx, wg, &test.args, time.Microsecond*500)
		for _, val := range test.want {
			_, ok := test.args.mtrcs[val]
			assert.True(t, ok)
		}
	})
}

func Test_metricToJSON(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
		want  []byte
	}{
		{
			name:  "Check counter",
			key:   "M1",
			value: counter(345),
			want:  []byte(`{"id":"M1","type":"counter","delta":345}`),
		},
		{
			name:  "Check gauge",
			key:   "M2",
			value: gauge(34.5),
			want:  []byte(`{"id":"M2","type":"gauge","value":34.5}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := metricToJSON(tt.key, tt.value, "")
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
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

func Test_dataToEnc(t *testing.T) {
	data := []byte("Test data")

	symmKey, err := genSymmKey(24)
	require.NoError(t, err)
	encData, err := dataToEnc(symmKey, data)
	require.NoError(t, err)
	assert.NotNil(t, encData)
	assert.Equal(t, 52, len(encData))

}

func Test_symmKeyToEnc(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "*pubkey.pem")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA8Y30n0rq5x6D/lXbN/m3
Th1RWQJO8XNYjmDTAEbFxx79qpXZVqSZYQnAZVD/9R05wzCn9mQFxJsJRgbGoymz
vXlatCGlNDtPz/kgpecY5ppOdBslEC6GiF1F6EvYAneNtN19dn1P+xm5A3mlmj7Z
d/gBGbtJILPXC8sEKNdrYS6WVIg8JkRC5eqC5uRPuFBAutG1fuG6WfNwlMAobaKP
CXXeaKdC4kgMNv8MgdnMzBMTXnkVDorbErfws7DNplGzy/ejwpGY5qiejMWl1Nxg
/VHheDrBUpFJqR1Qq1V1GE2fSKwG1L8SWqf2Tg0OC+l2XHcCEDrs6d+vhC6TDH61
GGnt+psAOGESkJDud3OPmlTZ4EdpGeMaQ0E3PiH3qhYCmPo10FvkwRJWWCnPEKJc
bpdMp8vec+EjD3lEean4hkdqA6hiIGSXCt+iKvn55zFmPElhwLAUlB1ucjogaKet
PKSN8ee0mWHQPzo5M7gZqHNdUGFLibmSDVPDkMc88jRNzaRsCoTwqxbXk0NQnZvi
I2A2REewXW8+qVgD3dF1J5Yvqpf54U0ofiEys95WtyqfuIrkfP3t8Rsgj4dwXuiI
Wtf8ZcFxHEBcOlrWyXcFAy6jhrnSazlDqflqSDLnzNCP2SynDgcAe8bIcejL7AWd
Lb01QRoI9TYWwpv+KMf8S4sCAwEAAQ==
-----END PUBLIC KEY-----`)
	require.NoError(t, err)
	tmpFile2, _ := os.CreateTemp("", "*key.pem")
	defer os.Remove(tmpFile2.Name())
	_, err = tmpFile2.WriteString(`-----BEGIN PRIVATE KEY-----
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
58qM7mUsPO5Imn8pJSz+Vsj3YTsJXoaKBg==
-----END PRIVATE KEY-----`)
	require.NoError(t, err)

	symmKey, err := genSymmKey(24)
	require.NoError(t, err)

	key, err := getPubKey(tmpFile.Name(), &log.Logger{})
	require.NoError(t, err)

	// encrypt
	encKey, err := symmKeyToEnc(key, symmKey)
	require.NoError(t, err)
	assert.NotNil(t, encKey)
	// decrypt
	privKey, err := usecase.GetPrivKey(tmpFile2.Name(), log.Default())
	require.NoError(t, err)
	symmKeyDec, err := usecase.DecryptKey(encKey, privKey)
	require.NoError(t, err)

	assert.Equal(t, symmKey, symmKeyDec)

}

func Test_reportMetrics(t *testing.T) {
	tmpFile1, _ := os.CreateTemp("", "*pubkey.pem")
	defer os.Remove(tmpFile1.Name())
	_, err := tmpFile1.WriteString(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA8Y30n0rq5x6D/lXbN/m3
Th1RWQJO8XNYjmDTAEbFxx79qpXZVqSZYQnAZVD/9R05wzCn9mQFxJsJRgbGoymz
vXlatCGlNDtPz/kgpecY5ppOdBslEC6GiF1F6EvYAneNtN19dn1P+xm5A3mlmj7Z
d/gBGbtJILPXC8sEKNdrYS6WVIg8JkRC5eqC5uRPuFBAutG1fuG6WfNwlMAobaKP
CXXeaKdC4kgMNv8MgdnMzBMTXnkVDorbErfws7DNplGzy/ejwpGY5qiejMWl1Nxg
/VHheDrBUpFJqR1Qq1V1GE2fSKwG1L8SWqf2Tg0OC+l2XHcCEDrs6d+vhC6TDH61
GGnt+psAOGESkJDud3OPmlTZ4EdpGeMaQ0E3PiH3qhYCmPo10FvkwRJWWCnPEKJc
bpdMp8vec+EjD3lEean4hkdqA6hiIGSXCt+iKvn55zFmPElhwLAUlB1ucjogaKet
PKSN8ee0mWHQPzo5M7gZqHNdUGFLibmSDVPDkMc88jRNzaRsCoTwqxbXk0NQnZvi
I2A2REewXW8+qVgD3dF1J5Yvqpf54U0ofiEys95WtyqfuIrkfP3t8Rsgj4dwXuiI
Wtf8ZcFxHEBcOlrWyXcFAy6jhrnSazlDqflqSDLnzNCP2SynDgcAe8bIcejL7AWd
Lb01QRoI9TYWwpv+KMf8S4sCAwEAAQ==
-----END PUBLIC KEY-----`)
	require.NoError(t, err)
	tmpFile2, _ := os.CreateTemp("", "*key.pem")
	defer os.Remove(tmpFile2.Name())
	_, err = tmpFile2.WriteString(`-----BEGIN PRIVATE KEY-----
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
58qM7mUsPO5Imn8pJSz+Vsj3YTsJXoaKBg==
-----END PRIVATE KEY-----`)
	require.NoError(t, err)

	t.Run("simple mode", func(t *testing.T) {
		srvAddr := ":63200"
		// grpc client
		clntConn, err := grpc.Dial(srvAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		client := pb.NewMetricsClient(clntConn)

		// grpc server
		srvListen, err := net.Listen("tcp", srvAddr)
		require.NoError(t, err)

		defer func() {
			if err := clntConn.Close(); err != nil {
				fmt.Println(err)
			}
		}()
		buff1 := make(map[string]interface{})
		memStor1 := storage.NewMemStorage(storage.WithBuffer(buff1))
		simpleServer := mygrpc.NewMetricsServer(config.Config{
			IsRestore:   false,
			StoreFile:   "",
			DatabaseDSN: "",
			CryptoKey:   "",
		}, log.Default())
		simpleServer.Storage = memStor1
		grpcSrv := grpc.NewServer()
		metrics := mmetrics{
			mtrcs: map[string]interface{}{"M1": gauge(43.1), "M2": counter(2)},
		}
		want := map[string]interface{}{"M2": int64(2), "M1": float64(43.1)}

		pb.RegisterMetricsServer(grpcSrv, simpleServer)

		ctx, cancel := context.WithTimeout(context.Background(), 9*time.Millisecond)
		defer cancel()
		wg := sync.WaitGroup{}

		// run server
		wg.Add(1)
		go func(w *sync.WaitGroup, s *grpc.Server) {
			defer w.Done()
			if err := s.Serve(srvListen); err != nil {
				log.Println(err)
			}
		}(&wg, grpcSrv)
		// run client
		wg.Add(1)
		go reportMetrics(ctx, &wg, &metrics, config.Config{
			ReportInterval: 5 * time.Millisecond,
		}, client, log.Default())
		// stop all
		<-ctx.Done()
		grpcSrv.GracefulStop()
		wg.Wait()
		// check result
		assert.Equal(t, want, buff1)
	})
	t.Run("encrypted mode", func(t *testing.T) {
		srvAddr := ":63201"
		// grpc client
		clntConn, err := grpc.Dial(srvAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		client := pb.NewMetricsClient(clntConn)

		// grpc server
		srvListen, err := net.Listen("tcp", srvAddr)
		require.NoError(t, err)

		defer func() {
			if err := clntConn.Close(); err != nil {
				fmt.Println(err)
			}
		}()
		buff2 := make(map[string]interface{})
		memStor2 := storage.NewMemStorage(storage.WithBuffer(buff2))
		encryptServer := mygrpc.NewMetricsServer(config.Config{
			IsRestore:   false,
			StoreFile:   "",
			DatabaseDSN: "",
			CryptoKey:   tmpFile2.Name(),
		}, log.Default())
		encryptServer.Storage = memStor2
		grpcSrv2 := grpc.NewServer()
		metrics := mmetrics{
			mtrcs: map[string]interface{}{"M3": gauge(43.1), "M4": counter(2)},
		}
		want := map[string]interface{}{"M4": int64(2), "M3": float64(43.1)}

		pb.RegisterMetricsServer(grpcSrv2, encryptServer)

		ctx, cancel := context.WithTimeout(context.Background(), 99*time.Millisecond)
		defer cancel()
		wg := sync.WaitGroup{}

		// run server
		wg.Add(1)
		go func(w *sync.WaitGroup, s *grpc.Server) {
			defer w.Done()
			if err := s.Serve(srvListen); err != nil {
				log.Println(err)
			}
		}(&wg, grpcSrv2)
		// run client
		wg.Add(1)
		go reportMetrics(ctx, &wg, &metrics, config.Config{
			ReportInterval: 50 * time.Millisecond,
			CryptoKey:      tmpFile1.Name(),
		}, client, log.Default())
		// stop all
		<-ctx.Done()
		grpcSrv2.GracefulStop()
		wg.Wait()
		// check result
		assert.Equal(t, want, buff2)
	})
}

func Test_genSymmKey(t *testing.T) {
	length := 24
	result, err := genSymmKey(length)
	require.NoError(t, err)
	require.Equal(t, length, len(result))
}

func Test_getLocalAddr(t *testing.T) {
	simpleHandl := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(""))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(simpleHandl))
	defer ts.Close()
	srvAddr := strings.Split(ts.URL, "//")[1]
	t.Run("without addr in config", func(t *testing.T) {
		conf := config.Config{
			ServerAddress: srvAddr,
		}
		assert.Equal(t, "127.0.0.1", getLocalAddr(conf, log.Default()))

	})
	t.Run("with addr in config", func(t *testing.T) {
		conf := config.Config{
			TrustedSubnet: "192.168.1.1",
		}
		assert.Equal(t, "192.168.1.1", getLocalAddr(conf, log.Default()))

	})
	t.Run("wrong server addr in config", func(t *testing.T) {
		conf := config.Config{
			ServerAddress: "127.0.01:80",
		}
		assert.Equal(t, "", getLocalAddr(conf, log.Default()))

	})
}
