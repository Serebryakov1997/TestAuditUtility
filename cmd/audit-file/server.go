package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/serebryakov1997/utility/audit"
	"github.com/serebryakov1997/utility/config"
	"github.com/serebryakov1997/utility/restapi"
	"github.com/serebryakov1997/utility/rules"
)

func runServer(args []string) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)

	addr := flags.String(
		"addr",
		"127.0.0.1:8080",
		"HTTP listen address",
	)

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if flags.NArg() != 0 {
		return errors.New("serve does not accept positional arguments")
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	analyzer := audit.New(rules.RegisteredRules()...)

	auditFn := func(
		ctx context.Context,
		data []byte,
		format string,
	) ([]audit.Finding, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		parseConfig, err := config.Parse(bytes.NewReader(data), format)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", restapi.ErrInvalidConfig, err)
		}

		if err := ctx.Err(); err != nil {
			return nil, err
		}

		return analyzer.Analyze(parseConfig)
	}

	handler, err := restapi.NewHandler(auditFn)
	if err != nil {
		return fmt.Errorf("create HTTP handler: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Starting HTTP server on %s\n", *addr)

	return restapi.Serve(ctx, *addr, handler)
}
