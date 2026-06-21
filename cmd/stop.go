package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-azure/pkg/azure"
	"github.com/devsy-org/log"
	"github.com/spf13/cobra"
)

// StopCmd holds the cmd flags.
type StopCmd struct{}

// NewStopCmd defines a command.
func NewStopCmd() *cobra.Command {
	cmd := &StopCmd{}
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop an instance",
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
func (cmd *StopCmd) Run(ctx context.Context, providerAzure *azure.AzureProvider) error {
	return azure.Stop(ctx, providerAzure)
}
