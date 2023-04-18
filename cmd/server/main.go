package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/hrapovd1/pmetrics/internal/config"
	"github.com/hrapovd1/pmetrics/internal/mygrpc"
	pb "github.com/hrapovd1/pmetrics/internal/proto"
	"github.com/hrapovd1/pmetrics/internal/types"
	"google.golang.org/grpc"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	logger := log.New(os.Stdout, "SERVER\t", log.Ldate|log.Ltime)
	// Чтение флагов и установка конфигурации сервера
	serverConf, err := config.NewServerConf(config.GetServerFlags())
	if err != nil {
		logger.Fatalln(err)
	}

	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	logger.Printf("\tBuild version: %s\n", buildVersion)
	logger.Printf("\tBuild date: %s\n", buildDate)
	logger.Printf("\tBuild commit: %s\n", buildCommit)
	logger.Println("Server start on ", serverConf.ServerAddress)

	grpcServer := mygrpc.NewMetricsServer(*serverConf, logger)
	srvStorage := grpcServer.Storage.(types.Storager)
	defer func() {
		if err := srvStorage.Close(); err != nil {
			logger.Print(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go srvStorage.Storing(ctx, &wg, logger, serverConf.StoreInterval, serverConf.IsRestore)

	listen, err := net.Listen("tcp", serverConf.ServerAddress)
	if err != nil {
		log.Fatalf("when open port got error: %v\n", err)
	}

	srv := grpc.NewServer(grpc.StreamInterceptor(grpcServer.StreamInterceptor))
	pb.RegisterMetricsServer(srv, grpcServer)

	wg.Add(1)
	go func(c context.Context, w *sync.WaitGroup, s *grpc.Server, l *log.Logger) {
		defer wg.Done()
		<-c.Done()
		l.Println("got signal to stop")
		s.GracefulStop()

	}(ctx, &wg, srv, logger)

	if err := srv.Serve(listen); err != http.ErrServerClosed {
		logger.Fatal(err)
	}

	wg.Wait()
	logger.Println("server stoped gracefully")
}
