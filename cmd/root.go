package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "fuse",
		Short: "Fuse - team-friendly Alertmanager configuration builder",
	}

	// Add subcommands
	root.AddCommand(newInitCmd())
	root.AddCommand(newValidateCmd())

	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
