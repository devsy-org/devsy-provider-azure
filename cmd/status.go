package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/devsy-org/devsy-provider-azure/pkg/azure"
	"github.com/spf13/cobra"
)

// StatusCmd holds the cmd flags.
type StatusCmd struct{}

// NewStatusCmd defines a command.
func NewStatusCmd() *cobra.Command {
	cmd := &StatusCmd{}
	return &cobra.Command{
		Use:   "status",
		Short: "Status an instance",
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
func (cmd *StatusCmd) Run(ctx context.Context, providerAzure *azure.AzureProvider) error {
	status, err := azure.Status(ctx, providerAzure)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, status)
	return err
}
