package client

import (
	"context"

	"github.com/msorokin-hash/passkeeper/internal/client/cli"
	"github.com/msorokin-hash/passkeeper/internal/client/config"
	"github.com/msorokin-hash/passkeeper/internal/client/grpc"
	"github.com/msorokin-hash/passkeeper/internal/client/service"
)

// Client represents the top-level GophKeeper client wrapper.
// It initializes the CLI layer and binds it to the underlying service
// and gRPC communication components.
type Client struct {
	cli *cli.CLI
}

// NewClient creates and returns a new Client instance. It establishes a gRPC
// connection using the provided configuration, initializes the service layer,
// sets up the CLI, and binds all components together.
// An error is returned if the gRPC connection cannot be created.
func NewClient(cfg *config.ClientConfig) (*Client, error) {
	ctx := context.Background()

	grpcClient, err := grpc.NewGRPCConnection(cfg)
	if err != nil {
		return nil, err
	}

	service := service.NewService(grpcClient, cfg.WorkDir)
	cliClient := cli.NewCLI(ctx, service)

	return &Client{
		cli: cliClient,
	}, nil
}

// Execute starts processing CLI commands. It delegates execution to the
// underlying Cobra CLI instance.
func (c *Client) Execute() error {
	return c.cli.Execute()
}
