package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindProjectRoot walks upward from start (or CWD if empty) to locate .fuse.yaml.
func FindProjectRoot(start string) (string, error) {
	dir := start
	if strings.TrimSpace(dir) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getcwd: %w", err)
		}
		dir = cwd
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("abs: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(abs, ".fuse.yaml")); err == nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return "", fmt.Errorf("no .fuse.yaml found")
}

func PrettyPrint(data any) error {
	var buffer bytes.Buffer
	enc := json.NewEncoder(&buffer)
	enc.SetIndent("", "    ")
	if err := enc.Encode(data); err != nil {
		return err
	}
	fmt.Println(buffer.String())
	return nil
}
