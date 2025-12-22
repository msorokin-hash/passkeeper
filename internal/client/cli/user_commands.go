package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// RegisterCmd returns a Cobra command that handles user registration.
// The command accepts a login and password as flags and invokes the
// Service.Register method. On success, it prints the received auth token.
func (c *CLI) RegisterCmd(ctx context.Context) *cobra.Command {
	var login, password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register",
		Long:  "Register a new user in GophKeeper",
		RunE: func(_ *cobra.Command, _ []string) error {
			token, err := c.service.Register(ctx, login, password)
			if err != nil {
				return err
			}
			fmt.Println("User successfully registered, token:", token)
			return nil
		},
	}

	cmd.Flags().StringVarP(&login, "login", "l", "", "user login")
	_ = cmd.MarkFlagRequired("login")
	cmd.Flags().StringVarP(&password, "password", "p", "", "user password")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

// LoginCmd returns a Cobra command that authenticates an existing user.
// The command accepts a login and password as flags and invokes the
// Service.Login method. On success, it prints the issued auth token.
func (c *CLI) LoginCmd(ctx context.Context) *cobra.Command {
	var login, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login",
		Long:  "Authenticate using login and password",
		RunE: func(_ *cobra.Command, _ []string) error {
			token, err := c.service.Login(ctx, login, password)
			if err != nil {
				return err
			}
			fmt.Println("Login successful, token:", token)
			return nil
		},
	}

	cmd.Flags().StringVarP(&login, "login", "l", "", "user login")
	_ = cmd.MarkFlagRequired("login")
	cmd.Flags().StringVarP(&password, "password", "p", "", "user password")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}
