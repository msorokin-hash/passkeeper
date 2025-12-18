package cli

import (
	"context"

	"github.com/spf13/cobra"
)

// CLI describes the structure of the command-line application.
// It wraps all Cobra commands and provides integration with the underlying
// Service layer.
type CLI struct {
	service Service

	rootCommand  *cobra.Command
	userCommand  *cobra.Command
	vaultCommand *cobra.Command
}

// NewCLI constructs and returns a new CLI application instance.
// It initializes the root command and all subcommands related to user
// authentication and vault operations.
// The provided context is passed to all command handlers, and the service
// defines the business logic behind each operation.
func NewCLI(ctx context.Context, service Service) *CLI {
	cli := &CLI{
		service: service,
		rootCommand: &cobra.Command{
			Use: "gophkeeper",
		},
		userCommand: &cobra.Command{
			Use:   "user",
			Short: "Authentication commands",
			Long:  "Commands for creating a new user and authenticating via login and password",
		},
		vaultCommand: &cobra.Command{
			Use:   "vault",
			Short: "Vault operations",
			Long:  "Commands for working with the vault — adding, deleting, and retrieving data",
		},
	}

	cli.userCommand.AddCommand(
		cli.RegisterCmd(ctx),
		cli.LoginCmd(ctx),
	)

	cli.vaultCommand.AddCommand(
		cli.GetDataCmd(ctx),
		cli.GetAllByTypeCmd(ctx),
		cli.AddDataCmd(ctx),
		cli.DeleteDataCmd(ctx),
	)

	cli.rootCommand.AddCommand(cli.userCommand)
	cli.rootCommand.AddCommand(cli.vaultCommand)

	return cli
}

// Execute runs the root Cobra command and starts processing CLI input.
// It should be invoked from main() to launch the CLI.
func (c *CLI) Execute() error {
	return c.rootCommand.Execute()
}
