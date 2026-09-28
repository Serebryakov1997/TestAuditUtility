package grpcapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/serebryakov1997/utility/audit"
	auditv1 "github.com/serebryakov1997/utility/grpc/proto/audit/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Run(
	ctx context.Context,
	address string,
	analyzer audit.Analyzer,
) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen gRPC: %w", err)
	}
	defer listener.Close()

	log.Printf("gRPC server listening on %s", listener.Addr())

	server := grpc.NewServer()

	auditv1.RegisterAuditServiceServer(
		server,
		NewHandler(analyzer),
	)

	reflection.Register(server)

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.Serve(listener)
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("serve gRPC: %w", err)
		}
		return nil

	case <-ctx.Done():
		stopped := make(chan struct{})

		go func() {
			server.GracefulStop()
			close(stopped)
		}()

		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()

		select {
		case <-stopped:
		case <-timer.C:
			server.Stop()
		}

		return nil
	}
}
