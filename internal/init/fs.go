package init

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func WriteTemplateFile(efs embed.FS, templatePath, targetPath string, force, quiet bool) error {
	templatePath = filepath.Join("templates", templatePath)
	// Skip existing unless force
	if _, err := os.Stat(targetPath); err == nil && !force {
		if !quiet {
			fmt.Printf("SKIPPED: %s (already exists)\n", targetPath)
		}
		return nil
	}

	content, err := fs.ReadFile(efs, templatePath)
	if err != nil {
		return fmt.Errorf("failed to read embedded template %s: %w", templatePath, err)
	}

	if err := ensureDir(filepath.Dir(targetPath)); err != nil {
		return err
	}

	if err := os.WriteFile(targetPath, content, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", targetPath, err)
	}

	if !quiet {
		if force {
			fmt.Printf("OVERWRITTEN: %s\n", targetPath)
		} else {
			fmt.Printf("CREATED: %s\n", targetPath)
		}
	}
	return nil
}

func trimTemplatePrefix(path string) string {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return path
}

func findFuseProjectRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".fuse.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no .fuse.yaml found")
}
