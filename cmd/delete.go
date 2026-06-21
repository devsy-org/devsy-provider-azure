package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-azure/pkg/azure"
	"github.com/devsy-org/log"
	"github.com/spf13/cobra"
)

// DeleteCmd holds the cmd flags
type DeleteCmd struct{}

// NewDeleteCmd defines a command
func NewDeleteCmd() *cobra.Command {
	cmd := &DeleteCmd{}
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete an instance",
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
func (cmd *DeleteCmd) Run(ctx context.Context, providerAzure *azure.AzureProvider) error {
	return azure.Delete(ctx, providerAzure)
}
