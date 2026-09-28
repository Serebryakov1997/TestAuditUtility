package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/serebryakov1997/utility/audit"
	grpcapi "github.com/serebryakov1997/utility/grpc/api"
	"github.com/serebryakov1997/utility/rules"
)

func main() {
	address := flag.String(
		"addr",
		"127.0.0.1:50051",
		"gRPC listen address",
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	analyzer := audit.New(rules.RegisteredRules()...)

	if err := grpcapi.Run(ctx, *address, *analyzer); err != nil {
		log.Printf("gRPC server error: %v", err)
		stop()
		os.Exit(1)
	}
}
