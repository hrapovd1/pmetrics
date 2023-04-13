package mygrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/hrapovd1/pmetrics/internal/config"
	dbstorage "github.com/hrapovd1/pmetrics/internal/dbstrorage"
	"github.com/hrapovd1/pmetrics/internal/filestorage"
	pb "github.com/hrapovd1/pmetrics/internal/proto"
	"github.com/hrapovd1/pmetrics/internal/storage"
	"github.com/hrapovd1/pmetrics/internal/types"
	"github.com/hrapovd1/pmetrics/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	Storage types.Repository
	conf    config.Config
	logger  *log.Logger
}

// NewMetricsServer - grpc MetricsServer constructor
func NewMetricsServer(conf config.Config, logger *log.Logger) *MetricsServer {
	ms := MetricsServer{conf: conf, logger: logger}
	var fs *filestorage.FileStorage
	// Have mem, fs and db storage
	if ms.conf.StoreFile != "" && ms.conf.DatabaseDSN != "" {
		db, err := dbstorage.NewDBStorage(
			conf.DatabaseDSN,
			logger,
			filestorage.NewFileStorage(conf, make(map[string]interface{})),
		)
		if err != nil {
			logger.Fatal(err)
		}
		ms.Storage = db
	}
	// Have mem and db storage
	if ms.conf.DatabaseDSN != "" && ms.conf.StoreFile == "" {
		db, err := dbstorage.NewDBStorage(
			conf.DatabaseDSN,
			logger,
			storage.NewMemStorage(),
		)
		if err != nil {
			logger.Fatal(err)
		}
		ms.Storage = db
	}
	// Have mem and fs storage
	if ms.conf.StoreFile != "" && ms.conf.DatabaseDSN == "" {
		fs = filestorage.NewFileStorage(conf, make(map[string]interface{}))
		ms.Storage = fs
	}
	// Have mem storage
	if ms.conf.DatabaseDSN == "" && ms.conf.StoreFile == "" {
		mms := storage.NewMemStorage()
		ms.Storage = mms
	}
	return &ms
}

// ReportMetric - unary server metric for unencrypted data
func (ms *MetricsServer) ReportMetric(c context.Context, r *pb.MetricRequest) (*pb.MetricResponse, error) {
	if err := ms.writeMetric(c, r.Metric); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.MetricResponse{}, nil
}

// ReportEncMetric - unary server metric for encrypted data
func (ms *MetricsServer) ReportEncMetric(c context.Context, r *pb.EncMetricRequest) (*pb.MetricResponse, error) {
	if ms.conf.CryptoKey == "" {
		ms.logger.Print("got encrypted request, but CryptoKey wasn't provided")
		return nil, status.Errorf(codes.Internal, "encrypt not support")
	}
	key, err := usecase.GetPrivKey(ms.conf.CryptoKey, ms.logger)
	if err != nil {
		ms.logger.Printf("when open key file %s, got error: %v", ms.conf.CryptoKey, err)
		return nil, status.Errorf(codes.Internal, "error when decrypt")
	}

	// Decrypt data and write
	symmKey, err := usecase.DecryptKey(r.Data.Data0, key)
	if err != nil {
		ms.logger.Printf("when DecryptData got error: %v", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	dataJSON, err := usecase.DecryptData(r.Data.Data, symmKey)
	if err != nil {
		ms.logger.Printf("when DecryptData got error: %v", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if err := ms.writeMetric(c, dataJSON); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.MetricResponse{}, nil
}

// ReportMetrics - stream server method for unencrypted data
func (ms *MetricsServer) ReportMetrics(strm pb.Metrics_ReportMetricsServer) error {
	for {
		select {
		case <-strm.Context().Done():
			return nil
		default:
			grpcMetric, err := strm.Recv()
			if err == io.EOF {
				return strm.SendAndClose(&pb.MetricResponse{})
			}
			if err != nil {
				return err
			}

			if err := ms.writeMetric(strm.Context(), grpcMetric.Metric); err != nil {
				ms.logger.Printf("when writeMetric got error: %v\n", err)
				return err
			}
		}
	}
}

// ReportEncMetrics - stream server method for encrypted data
func (ms *MetricsServer) ReportEncMetrics(strm pb.Metrics_ReportEncMetricsServer) error {
	if ms.conf.CryptoKey == "" {
		ms.logger.Print("got encrypted request, but CryptoKey wasn't provided")
		return fmt.Errorf("encrypt not support")
	}
	key, err := usecase.GetPrivKey(ms.conf.CryptoKey, ms.logger)
	if err != nil {
		ms.logger.Printf("when open key file %s, got error: %v", ms.conf.CryptoKey, err)
		return fmt.Errorf("error when decrypt: %v", err)
	}

	for {
		select {
		case <-strm.Context().Done():
			return nil
		default:
			grpcMetric, err := strm.Recv()
			if err == io.EOF {
				return strm.SendAndClose(&pb.MetricResponse{})
			}
			if err != nil {
				return err
			}

			symmKey, err := usecase.DecryptKey(grpcMetric.Data.Data0, key)
			if err != nil {
				ms.logger.Printf("when DecryptKey got error: %v", err)
				return err
			}

			dataJSON, err := usecase.DecryptData(grpcMetric.Data.Data, symmKey)
			if err != nil {
				ms.logger.Printf("when DecryptData got error: %v", err)
				return err
			}
			if err := ms.writeMetric(strm.Context(), dataJSON); err != nil {
				return err
			}
		}
	}
}

// StreamInterceptor - check metadata value X-Real-IP from agent
func (ms *MetricsServer) StreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		values := md.Get("X-Real-IP")
		if len(values) > 0 {
			if ms.isTrustedAddr(values[0]) {
				return handler(srv, stream)
			}
			ms.logger.Printf("got untrusted request from: %v\n", values)
			return status.Error(codes.PermissionDenied, "untrusted agent")
		}
	}
	return status.Error(codes.PermissionDenied, "untrusted agent")
}

// isTrustedAddr - check agent address value
func (ms *MetricsServer) isTrustedAddr(addr string) bool {
	if ms.conf.TrustedSubnet == "" {
		return true
	}
	agentAddr := net.ParseIP(addr)
	if agentAddr == nil {
		return false
	}
	allow, err := usecase.CheckAddr(agentAddr, ms.conf.TrustedSubnet)
	if err != nil {
		ms.logger.Printf("when CheckAddr got err: %v\n", err)
		return false
	}
	return allow
}

// writeMetric - write metric in storage and check hash sign
func (ms *MetricsServer) writeMetric(ctx context.Context, data []byte) error {
	var metric types.Metric
	if err := json.Unmarshal(data, &metric); err != nil {
		ms.logger.Printf("when Unmarshal metric got error: %v", err)
		return fmt.Errorf("when Unmarshal metric got error: %v", err)
	}

	// check metric hash in data.
	if ms.conf.Key != "" {
		if !usecase.IsSignEqual(metric, ms.conf.Key) {
			return errors.New("sign metric is bad")
		}
	}

	// Write new metrics value
	err := usecase.WriteJSONMetric(
		ctx,
		metric,
		ms.Storage,
	)
	if err != nil {
		ms.logger.Printf("when WriteJSONMetric got error: %v", err)
		return fmt.Errorf("error when WriteJSONMetric: %v", err)
	}
	return nil
}
