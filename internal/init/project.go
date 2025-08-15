package init

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
)

//go:embed templates/project/*
var projectTemplates embed.FS

func InitProject(options InitOptions) error {
	dirs := []string{options.Path, "global", "teams", "dist"}
	for _, dir := range dirs {
		if dir != options.Path {
			dir = filepath.Join(options.Path, dir)
		}

		if err := ensureDir(dir); err != nil {
			return err
		}
	}

	files := []string{
		"project/.fuse.yaml",
		"project/global/global.yaml",
		"project/global/silence_windows.yaml",
		"project/teams/README.md",
	}

	for _, file := range files {
		target := filepath.Join(options.Path, trimTemplatePrefix(file))

		if options.NoSample && !strings.Contains(file, ".fuse.yaml") {
			continue
		}

		if err := writeTemplateFile(projectTemplates, file, target, options.Force, options.Quiet); err != nil {
			return err
		}
	}

	if !options.Quiet {
		fmt.Printf("Initialized fuse project at %s\n", options.Path)
	}

	return nil
}
