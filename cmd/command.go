package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/devsy-org/devsy-provider-azure/pkg/azure"
	"github.com/devsy-org/devsy/pkg/ssh"
	"github.com/spf13/cobra"
)

// CommandCmd holds the cmd flags.
type CommandCmd struct{}

// NewCommandCmd defines a command.
func NewCommandCmd() *cobra.Command {
	cmd := &CommandCmd{}
	return &cobra.Command{
		Use:   "command",
		Short: "Command an instance",
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
func (cmd *CommandCmd) Run(ctx context.Context, providerAzure *azure.AzureProvider) error {
	command := os.Getenv("COMMAND")
	if command == "" {
		return fmt.Errorf("command environment variable is missing")
	}

	privateKey, err := ssh.GetPrivateKeyRawBase(providerAzure.Config.MachineFolder)
	if err != nil {
		return fmt.Errorf("load private key: %w", err)
	}

	externalIP, err := azure.GetInstanceIP(ctx, providerAzure)
	if err != nil {
		return err
	}

	sshClient, err := ssh.NewSSHClient("devsy", externalIP+":22", privateKey)
	if err != nil {
		return fmt.Errorf("create ssh client: %w", err)
	}
	defer func() { _ = sshClient.Close() }()

	return ssh.Run(ctx, ssh.RunOptions{
		Client:  sshClient,
		Command: command,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
}
