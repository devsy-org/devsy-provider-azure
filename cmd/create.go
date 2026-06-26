package cmd

import (
	"context"

	"github.com/devsy-org/devsy-provider-azure/pkg/azure"
	"github.com/spf13/cobra"
)

// CreateCmd holds the cmd flags.
type CreateCmd struct{}

// NewCreateCmd defines a command.
func NewCreateCmd() *cobra.Command {
	cmd := &CreateCmd{}
	return &cobra.Command{
		Use:   "create",
		Short: "Create an instance",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			azureProvider, err := azure.NewProvider()
			if err != nil {
				return err
			}

			return cmd.Run(cobraCmd.Context(), azureProvider)
		},
	}
}

// Run runs the command logic.
func (cmd *CreateCmd) Run(ctx context.Context, providerAzure *azure.AzureProvider) error {
	return azure.Create(ctx, providerAzure)
}
