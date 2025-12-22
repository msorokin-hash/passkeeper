package main

import (
	"fmt"
	"os"

	"github.com/msorokin-hash/passkeeper/internal/client"
	"github.com/msorokin-hash/passkeeper/internal/client/config"
)

var (
	Version   = "N/A"
	BuildTime = "N/A"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(Version, BuildTime)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cli, err := client.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize CLI client: %w", err)
	}

	if err := cli.Execute(); err != nil {
		return err
	}

	return nil
}
