package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	phase1sql "github.com/Emiloart/HDIP/services/internal/phase1sql"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return usageError("")
	}

	switch args[0] {
	case "migrate":
		return runMigrate(ctx, args[1:], stdout)
	case "bootstrap":
		return runBootstrap(ctx, args[1:], stdout)
	default:
		return usageError(fmt.Sprintf("unknown command %q", args[0]))
	}
}

func runMigrate(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "up" {
		return usageError("expected `migrate up`")
	}

	flags := flag.NewFlagSet("phase1sql migrate up", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	driver := flags.String("driver", "pgx", "database driver")
	dsn := flags.String("dsn", "", "database connection string")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if strings.TrimSpace(*dsn) == "" {
		return fmt.Errorf("phase1sql migrate up requires --dsn")
	}

	if err := phase1sql.MigrateUp(ctx, *driver, *dsn); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(stdout, "phase1sql migrate up applied")
	return nil
}

func runBootstrap(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "trust" {
		return usageError("expected `bootstrap trust`")
	}

	flags := flag.NewFlagSet("phase1sql bootstrap trust", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	driver := flags.String("driver", "pgx", "database driver")
	dsn := flags.String("dsn", "", "database connection string")
	filePath := flags.String("file", "", "trust bootstrap file")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if strings.TrimSpace(*dsn) == "" {
		return fmt.Errorf("phase1sql bootstrap trust requires --dsn")
	}
	if strings.TrimSpace(*filePath) == "" {
		return fmt.Errorf("phase1sql bootstrap trust requires --file")
	}

	store, err := phase1sql.Open(*driver, *dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	result, err := phase1sql.ApplyTrustBootstrapFile(ctx, store, *filePath, time.Now().UTC())
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "phase1sql bootstrap trust applied %d issuer records\n", result.Applied)
	return nil
}

func usageError(message string) error {
	usage := "usage: phase1sql migrate up --dsn <dsn> [--driver <driver>] | phase1sql bootstrap trust --dsn <dsn> --file <path> [--driver <driver>]"
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("%s", usage)
	}

	return fmt.Errorf("%s: %s", message, usage)
}
