package dsl

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nyambati/fuse/internal/am"
	"github.com/nyambati/fuse/internal/diag"
	"github.com/stretchr/testify/assert/yaml"
)

//
// ===== Discovery =====
//

// LoadProject discovers global + teams and prepares in-memory placeholders.
// Parsing of YAML files is intentionally deferred to parser layer; this
// function focuses on discovery and path wiring. It returns diagnostics
// (warnings/errors) that occur during discovery.
func LoadProject(root string, teamFilter []string) (Project, []diag.Diagnostic) {
	var diags []diag.Diagnostic

	p := Project{
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
		team := Team{
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

// unmarshalYamlFile is a helper to read and unmarshal a YAML file.
// If optional is true, os.IsNotExist errors are ignored.
func unmarshalYamlFile(filePath string, out interface{}, optional bool) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) && optional {
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	if err := yaml.Unmarshal(b, out); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return nil
}

func loadGlobal(root string, p *Project) error {
	// global/global.yaml
	var raw map[string]any
	if err := unmarshalYamlFile(filepath.Join(root, "global", "global.yaml"), &raw, false); err != nil {
		return err
	}

	if g, ok := raw["global"]; ok {
		if m, ok2 := g.(map[string]any); ok2 {
			p.Global = m
		} else {
			p.Global = raw
		}
	}

	// global/silence_windows.yaml
	var swWrapped struct {
		SilenceWindows []SilenceWindow `yaml:"silence_windows"`
	}
	if err := unmarshalYamlFile(filepath.Join(root, "global", "silence_windows.yaml"), &swWrapped, true); err != nil {
		return err
	}
	p.SilenceWindows = append(p.SilenceWindows, swWrapped.SilenceWindows...)

	// global/inhibitors.yaml (optional)
	var ihWrapped struct {
		Inhibitors []Inhibitor `yaml:"inhibitors"`
	}
	if err := unmarshalYamlFile(filepath.Join(root, "global", "inhibitors.yaml"), &ihWrapped, true); err != nil {
		return err
	}
	p.Inhibitors = append(p.Inhibitors, ihWrapped.Inhibitors...)

	// global/root_route.yaml
	var routeWrapped struct {
		Route am.Route `yaml:"route"`
	}
	if err := unmarshalYamlFile(filepath.Join(root, "global", "root_route.yaml"), &routeWrapped, true); err != nil {
		return err
	}
	p.RootRoute = routeWrapped.Route

	return nil
}

func loadTeam(teamPath string, t *Team) error {
	// channels.yaml
	var chWrapped struct {
		Channels []Channel `yaml:"channels"`
	}
	if err := unmarshalYamlFile(filepath.Join(teamPath, "channels.yaml"), &chWrapped, false); err != nil {
		return err
	}
	t.Channels = append(t.Channels, chWrapped.Channels...)

	// flows.yaml
	var fWrapped struct {
		Flows []Flow `yaml:"flows"`
	}
	if err := unmarshalYamlFile(filepath.Join(teamPath, "flows.yaml"), &fWrapped, false); err != nil {
		return err
	}
	t.Flows = append(t.Flows, fWrapped.Flows...)

	// silence_windows.yaml
	var swWrapped struct {
		SilenceWindows []SilenceWindow `yaml:"silence_windows"`
	}
	if err := unmarshalYamlFile(filepath.Join(teamPath, "silence_windows.yaml"), &swWrapped, false); err != nil {
		return err
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
	// 	Inhibitors []Inhibitor `yaml:"inhibitors"`
	// }

	// if err := yaml.Unmarshal(b, &ihwrapped); err != nil {
	// 	return fmt.Errorf("failed to parse %s: %w", filepath.Join(teamPath, "inhibitors.yaml"), err)
	// }

	// t.Inhibitors = append(t.Inhibitors, ihwrapped.Inhibitors...)

	return nil
}