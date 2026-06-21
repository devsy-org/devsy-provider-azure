package cmd

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/devsy-org/devsy-provider-azure/pkg/options"
	"github.com/spf13/cobra"
)

// InitCmd holds the cmd flags.
type InitCmd struct{}

// NewInitCmd defines an init command.
func NewInitCmd() *cobra.Command {
	cmd := &InitCmd{}
	return &cobra.Command{
		Use:   "init",
		Short: "Init account",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.Run(cobraCmd.Context())
		},
	}
}

// Run runs the init logic.
func (cmd *InitCmd) Run(ctx context.Context) error {
	if _, err := options.FromEnv(true); err != nil {
		return err
	}

	if _, err := azidentity.NewDefaultAzureCredential(nil); err != nil {
		return fmt.Errorf("authentication failure: %w", err)
	}

	return nil
}
