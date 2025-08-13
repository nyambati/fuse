package init

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/nyambati/fuse/internal/utils"
)

//go:embed templates/team/*
var teamTemplates embed.FS

func InitTeam(options InitOptions) error {
	root, err := utils.FindProjectRoot(options.Path)
	if err != nil {
		return fmt.Errorf("cannot create team: %w", err)
	}

	if options.Team == "" {
		return fmt.Errorf("team name is required")
	}

	teamDir := filepath.Join(root, "teams", options.Team)
	if err := ensureDir(teamDir); err != nil {
		return err
	}

	files := []string{
		"team/channels.yaml",
		"team/flows.yaml",
		"team/silence_windows.yaml",
		"team/alerts/example.yaml",
		"team/templates/README.md",
	}

	for _, file := range files {
		target := filepath.Join(teamDir, trimTemplatePrefix(file))
		if options.NoSample {
			if file == "team/channels.yaml" || file == "team/flows.yaml" {
				continue
			}
		}
		if err := writeTemplateFile(teamTemplates, file, target, options.Force, options.Quiet); err != nil {
			return err
		}
	}

	if !options.Quiet {
		fmt.Printf("Added team folder: teams/%s\n", options.Team)
	}

	return nil
}
