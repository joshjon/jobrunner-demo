package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	servicev1 "github.com/joshjon/jobrunner-demo/gen/rpc/service/v1"
	"github.com/joshjon/jobrunner-demo/internal/job"
)

const (
	port        = 50051
	workerCount = 2
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	srv, serverErrs := job.NewServer(workerCount)

	grpcSrv := grpc.NewServer()
	servicev1.RegisterServiceServer(grpcSrv, srv)
	reflection.Register(grpcSrv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("unable to listen on port %d: %v", port, err)
	}
	defer lis.Close()

	grpcErrs := make(chan error)
	go func() {
		if serveErr := grpcSrv.Serve(lis); err != nil {
			grpcErrs <- serveErr
		}
	}()
	defer grpcSrv.GracefulStop()
	log.Printf("grpc server started on port %d\n", port)

	for {
		select {
		case <-ctx.Done():
			log.Println("server stopped: context cancelled")
			os.Exit(0)
		case err = <-grpcErrs:
			log.Printf("server stopped: unexpected grpc serve error: %v\n", err)
			os.Exit(1)
		case err = <-serverErrs:
			log.Printf("unexpected server error: %v\n", err)
		}
	}
}
