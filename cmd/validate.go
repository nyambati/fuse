package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nyambati/fuse/internal/am"
	"github.com/nyambati/fuse/internal/dsl"
	"github.com/nyambati/fuse/internal/parse"
	"github.com/nyambati/fuse/internal/secrets"
	"github.com/nyambati/fuse/internal/validate"
)

func newValidateCmd() *cobra.Command {
	var (
		path          string
		teams         []string
		secretsProv   string
		secretsConfig string
		amtoolPath    string
		strict        bool
		jsonOut       bool
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate Fuse DSL and generated Alertmanager config",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1) Discover project root
			root, err := dsl.FindProjectRoot(path)
			if err != nil {
				return fmt.Errorf("not a Fuse project (no .fuse.yaml): %w", err)
			}

			// 2) Load DSL (global + teams)
			proj, loadDiags := dsl.LoadProject(root, teams)
			if len(loadDiags) > 0 {
				// continue; we’ll include loader diagnostics
			}

			// 3) Secrets provider
			prov, err := secrets.NewProvider(secretsProv, secretsConfig)
			if err != nil {
				return fmt.Errorf("secrets provider: %w", err)
			}

			// 4) Build AM model in-memory (translate DSL → AM)
			amc, parseDiags := parse.ToAlertmanager(proj, prov)

			// 5) Semantic validation
			valDiags := validate.Project(proj, amc, validate.Options{Strict: strict})

			// 6) (Optional) amtool check-config
			toolDiags := am.CheckWithAmtool(amc, amtoolPath) // returns empty if not configured/found

			// 7) Collate diagnostics and decide exit code
			all := validate.Merge(loadDiags, parseDiags, valDiags, toolDiags)
			exit := validate.ExitCode(all, strict)

			// 8) Output
			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(all); err != nil {
					return fmt.Errorf("json output: %w", err)
				}
			} else {
				for _, d := range all {
					fmt.Printf("%s: %s", d.Level, d.Message)
					if d.File != "" {
						fmt.Printf(" [%s", d.File)
						if d.Line > 0 {
							fmt.Printf(":%d", d.Line)
						}
						fmt.Printf("]")
					}
					fmt.Println()
				}
			}

			// 9) Exit code handling
			switch exit {
			case 0:
				return nil
			case 1:
				// diff “changes” isn’t relevant here; treat as warnings
				return nil
			case 2:
				// warnings - non-strict
				return nil
			case 3:
				return fmt.Errorf("validation errors")
			case 4:
				return fmt.Errorf("external tool/provider failure")
			default:
				return fmt.Errorf("unknown exit code %d", exit)
			}
		},
	}

	cmd.Flags().StringVar(&path, "path", ".", "Project path or subdirectory")
	cmd.Flags().StringSliceVar(&teams, "team", nil, "Validate only specific team(s)")
	cmd.Flags().StringVar(&secretsProv, "secrets", "env", "Secrets provider: env|sops|vault|ssm")
	cmd.Flags().StringVar(&secretsConfig, "secrets-config", "", "Secrets provider config file")
	cmd.Flags().StringVar(&amtoolPath, "amtool", "", "Path to amtool for check-config (optional)")
	cmd.Flags().BoolVar(&strict, "strict", false, "Treat warnings as errors")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output diagnostics as JSON")

	return cmd
}
