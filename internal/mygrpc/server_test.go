package mygrpc

import (
	"context"
	"io"
	"log"
	"os"
	"testing"

	"github.com/hrapovd1/pmetrics/internal/config"
	dbstorage "github.com/hrapovd1/pmetrics/internal/dbstrorage"
	"github.com/hrapovd1/pmetrics/internal/filestorage"
	pb "github.com/hrapovd1/pmetrics/internal/proto"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestNewMetricsServer(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "*.json")
	defer os.Remove(tmpFile.Name())
	tests := []struct {
		name string
		conf config.Config
		stor types.Repository
	}{
		{
			name: "mem only",
			conf: config.Config{StoreFile: "", DatabaseDSN: ""},
			stor: &storage.MemStorage{},
		},
		{
			name: "file storage",
			conf: config.Config{StoreFile: tmpFile.Name(), DatabaseDSN: ""},
			stor: &filestorage.FileStorage{},
		},
		{
			name: "db storage",
			conf: config.Config{StoreFile: "", DatabaseDSN: "postgres"},
			stor: &dbstorage.DBStorage{},
		},
		{
			name: "all types storage",
			conf: config.Config{StoreFile: tmpFile.Name(), DatabaseDSN: "postgres"},
			stor: &dbstorage.DBStorage{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMetricsServer(test.conf, log.Default())
			assert.IsType(t, test.stor, ms.Storage)
		})
	}
}

func TestMetricsServer_ReportMetric(t *testing.T) {
	ms := NewMetricsServer(config.Config{StoreFile: "", DatabaseDSN: ""}, log.Default())
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		resp    *pb.MetricResponse
	}{
		{
			name:    "good",
			data:    []byte(`{"id":"M1","type":"gauge","value":45.1}`),
			wantErr: false,
			resp:    &pb.MetricResponse{},
		},
		{
			name:    "bad",
			data:    []byte(`{"id":"M1","type":"guge","value":45.1}`),
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := ms.ReportMetric(
				context.Background(),
				&pb.MetricRequest{Metric: test.data},
			)
			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.resp, res)
			}
		})
	}

}

func TestMetricsServer_ReportEncMetric(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "*key.pem")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`-----BEGIN PRIVATE KEY-----
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
58qM7mUsPO5Imn8pJSz+Vsj3YTsJXoaKBg==
-----END PRIVATE KEY-----`)
	require.NoError(t, err)
	tests := []struct {
		name    string
		data    *pb.EncMetric
		conf    config.Config
		wantErr bool
		resp    *pb.MetricResponse
	}{
		{
			name: "good",
			data: &pb.EncMetric{
				Data0: `7o8PZ1KZwYOkGm5mpNNRGXUDtLEaIl9jIYgwcgf5dHJkFB+GBOSfpmhdIoSfLqGUlHdyVuXhrT8FibzIpiUTgu9FqHDQPIX+5cW8q/rjTracSySzs/QyeqCRj9Fktlx9pV3GMtTwVmUIpmRwuafXfBCTo6mRw+PzwA9xNCaHjQOIFb1qls2mBJ6srGLnek9E4+KnSZVEoIwOuCnRao3dTZwLPvbny49+3FGRPXqAH6M6kqR6MrpO1veA8NmHfjxT9XcfbNS/JDEpJEzvY3VH1f7BQr0oN6XE6+o+0lrJkc2uodGtnXUbnUP96coaQDweQWWJFZRgKjpD8W6WOLjeAXgkvFXbkhpO1N5R9087JljGZEiA4dnP+QLl7f7D8ovzxtNy0ynLWCANisDA3aAjsCJcM6wYcY33C4YvLlK7tNj5y5nlh6OMljEINDbbiIYRpNXKek/UWIUpy28F+LOo3pv8zCde6DyYySOleSIkmAHT5YmyNMYzb4kHLOl2hjB+T46db2zRM22u5gFVzA/EEeR9QxknYb4EFwuRP7FhMIMfed7jDB3p+H/uPqOSWsZAQBpAYwV1+NmvIkj3Uym/LklRoTK5mjx7yC5M1DxhbT2GMK1ocxR2OOqSXpUskMcmfLq/6tqIuQrhtHIZfRN2v5Ob/TxEQ9PD8CtacXr2r88=`,
				Data:  `9DrvPNa9FjbJ1niTkkQJIFq7GAoRtq3vFww+cZD2K/dj5OWsFlMjNHPvRDCcSna3cvNTaryEi6ikJeEyVXfvCykOkqyOGVsbGs4PZA==`,
			},
			wantErr: false,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: tmpFile.Name()},
			resp:    &pb.MetricResponse{},
		},
		{
			name:    "bad config",
			wantErr: true,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: ""},
		},
		{
			name:    "bad private key file",
			wantErr: true,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: "/tmp/key"},
		},
		{
			name: "bad encrypted key",
			data: &pb.EncMetric{
				Data0: `7o8PZ1KZwYOkGmNNRGXUDtLEaIl9jIYgwcgf5dHJkFB+GBOSfpmhdIoSfLqGUlHdyVuXhrT8FibzIpiUTgu9FqHDQPIX+5cW8q/rjTracSySzs/QyeqCRj9Fktlx9pV3GMtTwVmUIpmRwuafXfBCTo6mRw+PzwA9xNCaHjQOIFb1qls2mBJ6srGLnek9E4+KnSZVEoIwOuCnRao3dTZwLPvbny49+3FGRPXqAH6M6kqR6MrpO1veA8NmHfjxT9XcfbNS/JDEpJEzvY3VH1f7BQr0oN6XE6+o+0lrJkc2uodGtnXUbnUP96coaQDweQWWJFZRgKjpD8W6WOLjeAXgkvFXbkhpO1N5R9087JljGZEiA4dnP+QLl7f7D8ovzxtNy0ynLWCANisDA3aAjsCJcM6wYcY33C4YvLlK7tNj5y5nlh6OMljEINDbbiIYRpNXKek/UWIUpy28F+LOo3pv8zCde6DyYySOleSIkmAHT5YmyNMYzb4kHLOl2hjB+T46db2zRM22u5gFVzA/EEeR9QxknYb4EFwuRP7FhMIMfed7jDB3p+H/uPqOSWsZAQBpAYwV1+NmvIkj3Uym/LklRoTK5mjx7yC5M1DxhbT2GMK1ocxR2OOqSXpUskMcmfLq/6tqIuQrhtHIZfRN2v5Ob/TxEQ9PD8CtacXr2r88=`,
				Data:  `9DrvPNa9FjbJ1niTkkQJIFq7GAoRtq3vFww+cZD2K/dj5OWsFlMjNHPvRDCcSna3cvNTaryEi6ikJeEyVXfvCykOkqyOGVsbGs4PZA==`,
			},
			wantErr: true,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: tmpFile.Name()},
			resp:    &pb.MetricResponse{},
		},
		{
			name: "bad encrypted data",
			data: &pb.EncMetric{
				Data0: `7o8PZ1KZwYOkGm5mpNNRGXUDtLEaIl9jIYgwcgf5dHJkFB+GBOSfpmhdIoSfLqGUlHdyVuXhrT8FibzIpiUTgu9FqHDQPIX+5cW8q/rjTracSySzs/QyeqCRj9Fktlx9pV3GMtTwVmUIpmRwuafXfBCTo6mRw+PzwA9xNCaHjQOIFb1qls2mBJ6srGLnek9E4+KnSZVEoIwOuCnRao3dTZwLPvbny49+3FGRPXqAH6M6kqR6MrpO1veA8NmHfjxT9XcfbNS/JDEpJEzvY3VH1f7BQr0oN6XE6+o+0lrJkc2uodGtnXUbnUP96coaQDweQWWJFZRgKjpD8W6WOLjeAXgkvFXbkhpO1N5R9087JljGZEiA4dnP+QLl7f7D8ovzxtNy0ynLWCANisDA3aAjsCJcM6wYcY33C4YvLlK7tNj5y5nlh6OMljEINDbbiIYRpNXKek/UWIUpy28F+LOo3pv8zCde6DyYySOleSIkmAHT5YmyNMYzb4kHLOl2hjB+T46db2zRM22u5gFVzA/EEeR9QxknYb4EFwuRP7FhMIMfed7jDB3p+H/uPqOSWsZAQBpAYwV1+NmvIkj3Uym/LklRoTK5mjx7yC5M1DxhbT2GMK1ocxR2OOqSXpUskMcmfLq/6tqIuQrhtHIZfRN2v5Ob/TxEQ9PD8CtacXr2r88=`,
				Data:  `9DrvPNa9FjbJ1nisdjlfjiow+cZD2K/dj5OWsFlMjNHPvRDCcSna3cvNTaryEi6ikJeEyVXfvCykOkqyOGVsbGs4PZA==`,
			},
			wantErr: true,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: tmpFile.Name()},
			resp:    &pb.MetricResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMetricsServer(test.conf, log.Default())
			res, err := ms.ReportEncMetric(
				context.Background(),
				&pb.EncMetricRequest{Data: test.data},
			)
			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.resp, res)
			}
		})
	}

}

func TestMetricsServer_writeMetric(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		conf    config.Config
	}{
		{
			name:    "good",
			data:    []byte(`{"id":"HeapSys","type":"gauge","value":3637248,"hash":"36102f618dd4262bf4621a8782c433ca8abc008037d33c09ccbd074f612cceaa"}`),
			wantErr: false,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", Key: "1234rewq"},
		},
		{
			name:    "bad key",
			data:    []byte(`{"id":"HeapSys","type":"gauge","value":3637248,"hash":"36102f618dd4262bf4621a8782c433ca8abc008037d33c09ccbd074f612cceaa"}`),
			wantErr: true,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", Key: "1234rewq."},
		},
		{
			name:    "bad data",
			data:    []byte(`{"id":"HeapSys""type":"gaue","value":3637248,"hash":"36102f618dd4262bf4621a8782c433ca8abc008037d33c09ccbd074f612cceaa"}`),
			wantErr: true,
			conf:    config.Config{StoreFile: "", DatabaseDSN: ""},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMetricsServer(test.conf, log.Default())
			err := ms.writeMetric(context.Background(), test.data)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestMetricsServer_isTrustedAddr(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
		conf    config.Config
	}{
		{
			name:    "good",
			addr:    "192.168.0.1",
			wantErr: false,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", TrustedSubnet: "192.168.0.0/24"},
		},
		{
			name:    "good empty",
			addr:    "192.168.0.1",
			wantErr: false,
			conf:    config.Config{StoreFile: "", DatabaseDSN: ""},
		},
		{
			name:    "bad addr",
			addr:    "192.168.1.1",
			wantErr: true,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", TrustedSubnet: "192.168.0.0/24"},
		},
		{
			name:    "bad addr format",
			addr:    "192.168.11",
			wantErr: true,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", TrustedSubnet: "192.168.0.0/24"},
		},
		{
			name:    "bad mask",
			addr:    "192.168.1.1",
			wantErr: true,
			conf:    config.Config{StoreFile: "", DatabaseDSN: "", TrustedSubnet: "192.168.0.0/2i"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := NewMetricsServer(test.conf, log.Default())
			result := ms.isTrustedAddr(test.addr)
			if test.wantErr {
				assert.False(t, result)
			} else {
				assert.True(t, result)
			}
		})
	}
}

type testStream struct {
	grpc.ServerStream
	buff  []*pb.MetricRequest
	count *int
	ctx   context.Context
}

func (ts testStream) Recv() (*pb.MetricRequest, error) {
	if *ts.count < len(ts.buff) {
		out := ts.buff[*ts.count]
		*ts.count++
		return out, nil
	}
	return nil, io.EOF
}
func (ts testStream) SendAndClose(resp *pb.MetricResponse) error {
	return nil
}
func (ts testStream) Context() context.Context {
	return ts.ctx
}

func TestMetricsServer_ReportMetrics(t *testing.T) {
	gCount := 0
	goodStrm := testStream{
		buff: []*pb.MetricRequest{
			{Metric: []byte(`{"id":"M1","type":"gauge","value":1.2}`)},
			{Metric: []byte(`{"id":"M2","type":"gauge","value":2.5}`)},
		},
		count: &gCount,
		ctx:   context.Background(),
	}
	bCount := 0
	badStrm := testStream{
		buff: []*pb.MetricRequest{
			{Metric: []byte(`{"id":"M3","type":"gauge","value":1.2}`)},
			{Metric: []byte(`{"id":"M2","type":"gaue","value":2.5}`)},
		},
		count: &bCount,
		ctx:   context.Background(),
	}
	ms := NewMetricsServer(config.Config{StoreFile: "", DatabaseDSN: ""}, log.Default())
	t.Run("good", func(t *testing.T) {
		err := ms.ReportMetrics(goodStrm)
		assert.NoError(t, err)
	})
	t.Run("bad", func(t *testing.T) {
		err := ms.ReportMetrics(badStrm)
		assert.Error(t, err)
	})
}

type testEncStream struct {
	grpc.ServerStream
	buff  []*pb.EncMetricRequest
	count *int
	ctx   context.Context
}

func (ts testEncStream) Recv() (*pb.EncMetricRequest, error) {
	if *ts.count < len(ts.buff) {
		out := ts.buff[*ts.count]
		*ts.count++
		return out, nil
	}
	return nil, io.EOF
}
func (ts testEncStream) SendAndClose(resp *pb.MetricResponse) error {
	return nil
}
func (ts testEncStream) Context() context.Context {
	return ts.ctx
}

func TestMetricsServer_ReportEncMetrics(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "*key.pem")
	defer os.Remove(tmpFile.Name())
	_, err := tmpFile.WriteString(`-----BEGIN PRIVATE KEY-----
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
58qM7mUsPO5Imn8pJSz+Vsj3YTsJXoaKBg==
-----END PRIVATE KEY-----`)
	require.NoError(t, err)
	gCount := 0
	goodStrm := testEncStream{
		buff: []*pb.EncMetricRequest{
			{Data: &pb.EncMetric{
				Data0: `7o8PZ1KZwYOkGm5mpNNRGXUDtLEaIl9jIYgwcgf5dHJkFB+GBOSfpmhdIoSfLqGUlHdyVuXhrT8FibzIpiUTgu9FqHDQPIX+5cW8q/rjTracSySzs/QyeqCRj9Fktlx9pV3GMtTwVmUIpmRwuafXfBCTo6mRw+PzwA9xNCaHjQOIFb1qls2mBJ6srGLnek9E4+KnSZVEoIwOuCnRao3dTZwLPvbny49+3FGRPXqAH6M6kqR6MrpO1veA8NmHfjxT9XcfbNS/JDEpJEzvY3VH1f7BQr0oN6XE6+o+0lrJkc2uodGtnXUbnUP96coaQDweQWWJFZRgKjpD8W6WOLjeAXgkvFXbkhpO1N5R9087JljGZEiA4dnP+QLl7f7D8ovzxtNy0ynLWCANisDA3aAjsCJcM6wYcY33C4YvLlK7tNj5y5nlh6OMljEINDbbiIYRpNXKek/UWIUpy28F+LOo3pv8zCde6DyYySOleSIkmAHT5YmyNMYzb4kHLOl2hjB+T46db2zRM22u5gFVzA/EEeR9QxknYb4EFwuRP7FhMIMfed7jDB3p+H/uPqOSWsZAQBpAYwV1+NmvIkj3Uym/LklRoTK5mjx7yC5M1DxhbT2GMK1ocxR2OOqSXpUskMcmfLq/6tqIuQrhtHIZfRN2v5Ob/TxEQ9PD8CtacXr2r88=`,
				Data:  `9DrvPNa9FjbJ1niTkkQJIFq7GAoRtq3vFww+cZD2K/dj5OWsFlMjNHPvRDCcSna3cvNTaryEi6ikJeEyVXfvCykOkqyOGVsbGs4PZA==`,
			}},
		},
		count: &gCount,
		ctx:   context.Background(),
	}
	bCount := 0
	badStrm1 := testEncStream{
		buff: []*pb.EncMetricRequest{
			{Data: &pb.EncMetric{
				Data0: `7o8Z1KZwYOkGm5mpNNRGXUDtLEaIl9jIYgwcgf5dHJkFB+GBOSfpmhdIoSfLqGUlHdyVuXhrT8FibzIpiUTgu9FqHDQPIX+5cW8q/rjTracSySzs/QyeqCRj9Fktlx9pV3GMtTwVmUIpmRwuafXfBCTo6mRw+PzwA9xNCaHjQOIFb1qls2mBJ6srGLnek9E4+KnSZVEoIwOuCnRao3dTZwLPvbny49+3FGRPXqAH6M6kqR6MrpO1veA8NmHfjxT9XcfbNS/JDEpJEzvY3VH1f7BQr0oN6XE6+o+0lrJkc2uodGtnXUbnUP96coaQDweQWWJFZRgKjpD8W6WOLjeAXgkvFXbkhpO1N5R9087JljGZEiA4dnP+QLl7f7D8ovzxtNy0ynLWCANisDA3aAjsCJcM6wYcY33C4YvLlK7tNj5y5nlh6OMljEINDbbiIYRpNXKek/UWIUpy28F+LOo3pv8zCde6DyYySOleSIkmAHT5YmyNMYzb4kHLOl2hjB+T46db2zRM22u5gFVzA/EEeR9QxknYb4EFwuRP7FhMIMfed7jDB3p+H/uPqOSWsZAQBpAYwV1+NmvIkj3Uym/LklRoTK5mjx7yC5M1DxhbT2GMK1ocxR2OOqSXpUskMcmfLq/6tqIuQrhtHIZfRN2v5Ob/TxEQ9PD8CtacXr2r88=`,
				Data:  `9DrvPNa9FjbJ1niTkkQJIFq7GAoRtq3vFww+cZD2K/dj5OWsFlMjNHPvRDCcSna3cvNTaryEi6ikJeEyVXfvCykOkqyOGVsbGs4PZA==`,
			}},
		},
		count: &bCount,
		ctx:   context.Background(),
	}
	badStrm2 := testEncStream{
		buff: []*pb.EncMetricRequest{
			{Data: &pb.EncMetric{
				Data0: `7o8PZ1KZwYOkGm5mpNNRGXUDtLEaIl9jIYgwcgf5dHJkFB+GBOSfpmhdIoSfLqGUlHdyVuXhrT8FibzIpiUTgu9FqHDQPIX+5cW8q/rjTracSySzs/QyeqCRj9Fktlx9pV3GMtTwVmUIpmRwuafXfBCTo6mRw+PzwA9xNCaHjQOIFb1qls2mBJ6srGLnek9E4+KnSZVEoIwOuCnRao3dTZwLPvbny49+3FGRPXqAH6M6kqR6MrpO1veA8NmHfjxT9XcfbNS/JDEpJEzvY3VH1f7BQr0oN6XE6+o+0lrJkc2uodGtnXUbnUP96coaQDweQWWJFZRgKjpD8W6WOLjeAXgkvFXbkhpO1N5R9087JljGZEiA4dnP+QLl7f7D8ovzxtNy0ynLWCANisDA3aAjsCJcM6wYcY33C4YvLlK7tNj5y5nlh6OMljEINDbbiIYRpNXKek/UWIUpy28F+LOo3pv8zCde6DyYySOleSIkmAHT5YmyNMYzb4kHLOl2hjB+T46db2zRM22u5gFVzA/EEeR9QxknYb4EFwuRP7FhMIMfed7jDB3p+H/uPqOSWsZAQBpAYwV1+NmvIkj3Uym/LklRoTK5mjx7yC5M1DxhbT2GMK1ocxR2OOqSXpUskMcmfLq/6tqIuQrhtHIZfRN2v5Ob/TxEQ9PD8CtacXr2r88=`,
				Data:  "",
			}},
		},
		count: &bCount,
		ctx:   context.Background(),
	}
	t.Run("good", func(t *testing.T) {
		ms := NewMetricsServer(config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: tmpFile.Name()}, log.Default())
		err := ms.ReportEncMetrics(goodStrm)
		assert.NoError(t, err)
	})
	t.Run("bad config", func(t *testing.T) {
		ms := NewMetricsServer(config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: ""}, log.Default())
		err := ms.ReportEncMetrics(goodStrm)
		assert.Error(t, err)
	})
	t.Run("bad private key file", func(t *testing.T) {
		ms := NewMetricsServer(config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: "/tmp/key"}, log.Default())
		err := ms.ReportEncMetrics(goodStrm)
		assert.Error(t, err)
	})
	t.Run("bad encrypted key", func(t *testing.T) {
		ms := NewMetricsServer(config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: tmpFile.Name()}, log.Default())
		err := ms.ReportEncMetrics(badStrm1)
		assert.Error(t, err)
	})
	t.Run("bad encrypted data", func(t *testing.T) {
		ms := NewMetricsServer(config.Config{StoreFile: "", DatabaseDSN: "", CryptoKey: tmpFile.Name()}, log.Default())
		err1 := ms.ReportEncMetrics(badStrm2)
		assert.Nil(t, err1)
	})
}
