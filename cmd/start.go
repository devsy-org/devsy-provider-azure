package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-azure/pkg/azure"
	"github.com/devsy-org/log"
	"github.com/spf13/cobra"
)

// StartCmd holds the cmd flags.
type StartCmd struct{}

// NewStartCmd defines a command.
func NewStartCmd() *cobra.Command {
	cmd := &StartCmd{}
	return &cobra.Command{
		Use:   "start",
		Short: "Start an instance",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			azureProvider, err := azure.NewProvider(log.Default)
			if err != nil {
				return err
			}

			return cmd.Run(cobraCmd.Context(), azureProvider)
		},
	}
}

// Run runs the command logic.
func (cmd *StartCmd) Run(ctx context.Context, providerAzure *azure.AzureProvider) error {
	return azure.Start(ctx, providerAzure)
}
