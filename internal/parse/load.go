package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/types"
	"github.com/stretchr/testify/assert/yaml"
)

//
// ===== Discovery =====
//

// LoadProject discovers global + teams and prepares in-memory placeholders.
// Parsing of YAML files is intentionally deferred to parser layer; this
// function focuses on discovery and path wiring. It returns diagnostics
// (warnings/errors) that occur during discovery.
func LoadProject(root string, teamFilter []string) (types.Project, []diag.Diagnostic) {
	var diags []diag.Diagnostic

	p := types.Project{
		Root: root,
		// Global/SilenceWindows will be populated by a loader later.
	}

	// Discover teams directory
	teamsDir := filepath.Join(root, "teams")
	info, err := os.Stat(teamsDir)
	if err != nil {
		// If teams/ does not exist, warn but continue (project may only have global for now)
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelWarn,
			Code:    "DISCOVER_NO_TEAMS_DIR",
			Message: "teams/ directory not found; continuing with global-only project",
			File:    filepath.Join(root, "teams"),
		})
		return p, diags
	}
	if !info.IsDir() {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "DISCOVER_TEAMS_NOT_DIR",
			Message: "teams exists but is not a directory",
			File:    teamsDir,
		})
		return p, diags
	}

	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
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
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelWarn,
				Code:    "DISCOVER_NO_TEAMS_MATCH",
				Message: "no teams matched the provided --team filter",
				File:    teamsDir,
			})
		}
	}

	return p, diags
}

func loadGlobal(root string, p *types.Project, diags *[]diag.Diagnostic) {
	// global/global.yaml
	if b, err := os.ReadFile(filepath.Join(root, "global", "global.yaml")); err == nil {
		// keep permissive: accept either:
		//   global: { ... }
		// or plain map we treat as global
		var raw map[string]any
		if err := yaml.Unmarshal(b, &raw); err != nil {
			*diags = append(*diags, diag.Error("YAML_GLOBAL_PARSE",
				fmt.Sprintf("failed to parse global/global.yaml: %v", err),
				filepath.Join(root, "global", "global.yaml"), 0))
		} else {
			if g, ok := raw["global"]; ok {
				if m, ok2 := g.(map[string]any); ok2 {
					p.Global.Raw = m
				} else {
					p.Global.Raw = raw
				}
			} else {
				p.Global.Raw = raw
			}
		}
	}

	// global/silence_windows.yaml
	if b, err := os.ReadFile(filepath.Join(root, "global", "silence_windows.yaml")); err == nil {
		var wrapped struct {
			SilenceWindows []types.SilenceWindow `yaml:"silence_windows"`
		}
		if err := yaml.Unmarshal(b, &wrapped); err != nil {
			*diags = append(*diags, diag.Error("YAML_GLOBAL_SILENCE_PARSE",
				fmt.Sprintf("failed to parse global/silence_windows.yaml: %v", err),
				filepath.Join(root, "global", "silence_windows.yaml"), 0))
		} else {
			p.SilenceWindows = append(p.SilenceWindows, wrapped.SilenceWindows...)
		}
	}

	// global/inhibitors.yaml (optional)
	if b, err := os.ReadFile(filepath.Join(root, "global", "inhibitors.yaml")); err == nil {
		var wrapped struct {
			Inhibitors []types.Inhibitor `yaml:"inhibitors"`
		}
		if err := yaml.Unmarshal(b, &wrapped); err != nil {
			*diags = append(*diags, diag.Error("YAML_GLOBAL_INHIBITORS_PARSE",
				fmt.Sprintf("failed to parse global/inhibitors.yaml: %v", err),
				filepath.Join(root, "global", "inhibitors.yaml"), 0))
		} else {
			// Project-level inhibitors? If you want global + team inhibitors combined later,
			// you can add a field to Project. For now we’ll keep them in Global.Raw as passthrough.
			// If you already have p.Inhibitors on Project, append there instead:
			// p.Inhibitors = append(p.Inhibitors, wrapped.Inhibitors...)
			// For now, stash in Raw for later parse stage if needed.
			if p.Global.Raw == nil {
				p.Global.Raw = map[string]any{}
			}
			p.Global.Raw["_fuse_global_inhibitors"] = wrapped.Inhibitors
		}
	}
}
