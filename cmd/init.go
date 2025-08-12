package cmd

import (
	"fmt"
	"path/filepath"

	initer "github.com/nyambati/fuse/internal/init"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		team     string
		force    bool
		quiet    bool
		noSample bool
	)

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new Fuse project or team folder",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("failed to resolve absolute path: %w", err)
			}

			options := initer.InitOptions{
				Path:     absPath,
				Team:     team,
				Force:    force,
				Quiet:    quiet,
				NoSample: noSample,
			}

			if team != "" {
				return initer.InitTeam(options)
			}

			return initer.InitProject(options)
		},
	}

	cmd.Flags().StringVarP(&team, "team", "t", "", "Initialize a team folder in an existing fuse project")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress output")
	cmd.Flags().BoolVarP(&noSample, "no-sample", "n", false, "Do not create sample files")

	return cmd
}
