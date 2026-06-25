package cmd

import (
	"os"
	"os/exec"

	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// NewRootCmd returns a new root command.
func NewRootCmd() *cobra.Command {
	azureCmd := &cobra.Command{
		Use:           "devsy-provider-azure",
		Short:         "azure Provider commands",
		SilenceErrors: true,
		SilenceUsage:  true,

		PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
			cfg := log.Config{Verbosity: 1}
			if os.Getenv("DEVSY_DEBUG") == "true" {
				cfg.Debug = true
			}
			log.Init(cfg)

			return nil
		},
	}

	return azureCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd := BuildRoot()

	err := rootCmd.Execute()
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			os.Exit(exitErr.ExitStatus())
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
				log.Error(string(exitErr.Stderr))
			}

			os.Exit(exitErr.ExitCode())
		}

		log.Fatal(err)
	}
}

// BuildRoot creates a new root command.
func BuildRoot() *cobra.Command {
	rootCmd := NewRootCmd()

	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewDeleteCmd())
	rootCmd.AddCommand(NewCommandCmd())
	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewStopCmd())
	rootCmd.AddCommand(NewStatusCmd())

	return rootCmd
}
