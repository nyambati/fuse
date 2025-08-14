package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nyambati/fuse/internal/am"
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

	if err := loadGlobal(root, &p); err != nil {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "LOAD_GLOBAL",
			Message: fmt.Sprintf("failed to load global configuration: %v", err),
		})
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
		team := types.Team{
			Name: name,
			Path: teamPath,
		}

		if err := loadTeam(teamPath, &team); err != nil {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "READ_TEAM",
				Message: fmt.Sprintf("failed to read team %s: %v", name, err),
				File:    teamPath,
			})
		} else {
			p.Teams = append(p.Teams, team)
		}
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

func loadGlobal(root string, p *types.Project) error {
	// global/global.yaml
	b, err := os.ReadFile(filepath.Join(root, "global", "global.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read global/global.yaml: %w", err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("failed to parse global/global.yaml: %w", err)
	}

	if g, ok := raw["global"]; ok {
		if m, ok2 := g.(map[string]any); ok2 {
			p.Global = m
		} else {
			p.Global = raw
		}
	}

	// global/silence_windows.yaml
	b, err = os.ReadFile(filepath.Join(root, "global", "silence_windows.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read global/silence_windows.yaml: %w", err)
	}

	var swWrapped struct {
		SilenceWindows []types.SilenceWindow `yaml:"silence_windows"`
	}

	if err := yaml.Unmarshal(b, &swWrapped); err != nil {
		return fmt.Errorf("failed to parse global/silence_windows.yaml: %w", err)
	}

	p.SilenceWindows = append(p.SilenceWindows, swWrapped.SilenceWindows...)

	// global/inhibitors.yaml (optional)
	b, err = os.ReadFile(filepath.Join(root, "global", "inhibitors.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read global/inhibitors.yaml: %w", err)
	}

	var ihWrapped struct {
		Inhibitors []types.Inhibitor `yaml:"inhibitors"`
	}
	if err := yaml.Unmarshal(b, &ihWrapped); err != nil {
		return fmt.Errorf("failed to parse global/inhibitors.yaml: %w", err)
	}
	p.Inhibitors = append(p.Inhibitors, ihWrapped.Inhibitors...)

	// global/root_route.yaml
	b, err = os.ReadFile(filepath.Join(root, "global", "root_route.yaml"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read global/root_route.yaml: %w", err)
	}

	var routeWrapped struct {
		Route am.Route `yaml:"route"`
	}
	if err := yaml.Unmarshal(b, &routeWrapped); err != nil {
		return fmt.Errorf("failed to parse global/root_route.yaml: %w", err)
	}

	p.RootRoute = &routeWrapped.Route

	return nil
}

func loadTeam(teamPath string, t *types.Team) error {
	// channels.yaml
	b, err := os.ReadFile(filepath.Join(teamPath, "channels.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filepath.Join(teamPath, "channels.yaml"), err)
	}
	var chWrapped struct {
		Channels []types.Channel `yaml:"channels"`
	}
	if err := yaml.Unmarshal(b, &chWrapped); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filepath.Join(teamPath, "channels.yaml"), err)
	}

	t.Channels = append(t.Channels, chWrapped.Channels...)

	// flows.yaml
	b, err = os.ReadFile(filepath.Join(teamPath, "flows.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filepath.Join(teamPath, "flows.yaml"), err)
	}
	var fWrapped struct {
		Flows []types.Flow `yaml:"flows"`
	}

	if err := yaml.Unmarshal(b, &fWrapped); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filepath.Join(teamPath, "flows.yaml"), err)
	}

	for _, fy := range fWrapped.Flows {
		t.Flows = append(t.Flows, fy)
	}

	// silence_windows.yaml
	b, err = os.ReadFile(filepath.Join(teamPath, "silence_windows.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filepath.Join(teamPath, "silence_windows.yaml"), err)
	}

	var swWrapped struct {
		SilenceWindows []types.SilenceWindow `yaml:"silence_windows"`
	}
	if err := yaml.Unmarshal(b, &swWrapped); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filepath.Join(teamPath, "silence_windows.yaml"), err)
	}

	t.SilenceWindows = append(t.SilenceWindows, swWrapped.SilenceWindows...)

	// // inhibitors.yaml (optional)
	// b, err = os.ReadFile(filepath.Join(teamPath, "inhibitors.yaml"))
	// if err != nil {
	// 	if os.IsNotExist(err) {
	// 		return nil
	// 	}
	// 	return fmt.Errorf("failed to read %s: %w", filepath.Join(teamPath, "inhibitors.yaml"), err)
	// }

	// var ihwrapped struct {
	// 	Inhibitors []types.Inhibitor `yaml:"inhibitors"`
	// }

	// if err := yaml.Unmarshal(b, &ihwrapped); err != nil {
	// 	return fmt.Errorf("failed to parse %s: %w", filepath.Join(teamPath, "inhibitors.yaml"), err)
	// }

	// t.Inhibitors = append(t.Inhibitors, ihwrapped.Inhibitors...)

	return nil
}
