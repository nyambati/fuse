package dsl

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nyambati/fuse/internal/types"
)

//
// ===== Discovery =====
//

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

// LoadProject discovers global + teams and prepares in-memory placeholders.
// Parsing of YAML files is intentionally deferred to parser layer; this
// function focuses on discovery and path wiring. It returns diagnostics
// (warnings/errors) that occur during discovery.
func LoadProject(root string, teamFilter []string) (types.Project, []types.Diagnostic) {
	var diags []types.Diagnostic

	p := types.Project{
		Root: root,
		// Global/SilenceWindows will be populated by a loader later.
	}

	// Discover teams directory
	teamsDir := filepath.Join(root, "teams")
	info, err := os.Stat(teamsDir)
	if err != nil {
		// If teams/ does not exist, warn but continue (project may only have global for now)
		diags = append(diags, types.Diagnostic{
			Level:   types.LevelWarn,
			Code:    "DISCOVER_NO_TEAMS_DIR",
			Message: "teams/ directory not found; continuing with global-only project",
			File:    filepath.Join(root, "teams"),
		})
		return p, diags
	}
	if !info.IsDir() {
		diags = append(diags, types.Diagnostic{
			Level:   types.LevelError,
			Code:    "DISCOVER_TEAMS_NOT_DIR",
			Message: "teams exists but is not a directory",
			File:    teamsDir,
		})
		return p, diags
	}

	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		diags = append(diags, types.Diagnostic{
			Level:   types.LevelError,
			Code:    "READ_TEAMS_DIR",
			Message: fmt.Sprintf("failed to read teams directory: %v", err),
			File:    teamsDir,
		})
		return p, diags
	}

	// Build a filter set if provided.
	filter := map[string]struct{}{}
	for _, t := range teamFilter {
		filter[strings.TrimSpace(t)] = struct{}{}
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if len(filter) > 0 {
			if _, ok := filter[name]; !ok {
				continue
			}
		}
		teamPath := filepath.Join(teamsDir, name)
		p.Teams = append(p.Teams, types.Team{
			Name: name,
			Path: teamPath,
			// Channels/Flows/SilenceWindows will be populated by YAML loaders later.
		})
	}

	// Sort teams for stable output
	sort.Slice(p.Teams, func(i, j int) bool { return p.Teams[i].Name < p.Teams[j].Name })

	if len(p.Teams) == 0 {
		// If a filter was provided and nothing matched, warn; otherwise, info-level would be fine,
		// but we don't have info level in diagnostics yet—use WARN.
		if len(filter) > 0 {
			diags = append(diags, types.Diagnostic{
				Level:   types.LevelWarn,
				Code:    "DISCOVER_NO_TEAMS_MATCH",
				Message: "no teams matched the provided --team filter",
				File:    teamsDir,
			})
		}
	}

	return p, diags
}
