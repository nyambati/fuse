package validate

import (
	"fmt"
	"sort"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

// Options for validation behavior
type Options struct {
	Strict bool
}

// Project runs semantic validation on a loaded DSL project and the derived AM config.
func Project(proj dsl.Project, amc any, opts Options) []diag.Diagnostic {
	var diags []diag.Diagnostic

	// ---- Basic project-level checks ----
	if proj.Root == "" {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "PROJ_ROOT_EMPTY",
			Message: "project root not set",
		})
	}

	// Teams must have unique names (safety check).
	seen := map[string]struct{}{}
	for _, t := range proj.Teams {
		if t.Name == "" {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelWarn,
				Code:    "TEAM_NAME_EMPTY",
				Message: "a team folder has an empty name",
				File:    t.Path,
			})
			continue
		}
		if _, ok := seen[t.Name]; ok {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "TEAM_NAME_DUP",
				Message: fmt.Sprintf("duplicate team name %q", t.Name),
				File:    t.Path,
			})
		}

		diags = append(diags, validateFlows(t)...)
		seen[t.Name] = struct{}{}
	}

	// TODO (next steps):
	// - Validate silence_windows names uniqueness (global vs team shadowing -> warn)
	// - Validate channels: unique names within a team, required params per type
	// - Validate inhibitors: fields present, matcher syntax sanity
	// - Time/duration parsing checks for wait/group/repeat

	return diags
}

// validateFlows checks notify presence, channel existence, when block validity, and duplicates.
func validateFlows(team dsl.Team) []diag.Diagnostic {
	var diags []diag.Diagnostic

	// Build a set of channels in this team
	channelSet := make(map[string]struct{})
	for _, ch := range team.Channels {
		channelSet[ch.Name] = struct{}{}
	}

	// Track seen (notify, when) combinations for duplicate detection
	seenCombos := make(map[string]string)

	for _, flow := range team.Flows {
		// 1. Missing notify
		if len(flow.Notify) == 0 {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "FLOW_NOTIFY_EMPTY",
				Message: fmt.Sprintf("flow in team %q has no notify target", team.Name),
			})
		}

		// 2. Non-existent notify channel

		if _, ok := channelSet[flow.Notify]; !ok {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "FLOW_NOTIFY_UNKNOWN",
				Message: fmt.Sprintf("flow in team %q references unknown channel %q", team.Name, flow.Notify),
			})
		}

		// 3. Empty or missing when
		if len(flow.When) == 0 {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "FLOW_WHEN_EMPTY",
				Message: fmt.Sprintf("flow in team %q has no conditions (when block is empty)", team.Name),
			})
		}

		// 4. Invalid matcher syntax
		for k, v := range flow.When {
			if err := validateMatcher(k, v); err != nil {
				diags = append(diags, diag.Diagnostic{
					Level:   diag.LevelError,
					Code:    "FLOW_MATCHER_INVALID",
					Message: fmt.Sprintf("invalid matcher in flow (team %q): %v", team.Name, err),
				})
			}
		}

		// 5. Duplicate flow matcher sets
		hash := hashFlow(flow)
		if existing, ok := seenCombos[hash]; ok {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelWarn,
				Code:    "FLOW_DUPLICATE",
				Message: fmt.Sprintf("flow in team %q duplicates another flow: %s", team.Name, existing),
			})
		} else {
			seenCombos[hash] = fmt.Sprintf("notify=%v when=%v", flow.Notify, flow.When)
		}
	}

	return diags
}

// validateMatcher checks that the matcher key/value is syntactically valid.
// This is a placeholder — expand with proper Alertmanager matcher parsing.
func validateMatcher(key string, val any) error {
	// For now, reject empty keys and values
	if key == "" {
		return fmt.Errorf("matcher key is empty")
	}
	if val == nil || val == "" {
		return fmt.Errorf("matcher %q has empty value", key)
	}
	return nil
}

// hashFlow creates a simple hash key for a flow based on notify + when map.
func hashFlow(flow dsl.Flow) string {
	return fmt.Sprintf("notify=%v when=%v", flow.Notify, flow.When)
}

// Merge combines multiple diagnostic slices and sorts them
func Merge(diags ...[]diag.Diagnostic) []diag.Diagnostic {
	var all []diag.Diagnostic
	for _, d := range diags {
		if len(d) > 0 {
			all = append(all, d...)
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Level != all[j].Level {
			return all[i].Level > all[j].Level // ERROR > WARN > INFO
		}
		if all[i].File != all[j].File {
			return all[i].File < all[j].File
		}
		if all[i].Line != all[j].Line {
			return all[i].Line < all[j].Line
		}
		return all[i].Message < all[j].Message
	})
	return all
}

// ExitCode returns the exit code based on diagnostics and strict mode
// 0 = no issues
// 2 = warnings only (strict=false)
// 3 = errors found
func ExitCode(diags []diag.Diagnostic, strict bool) int {
	var hasWarn, hasErr bool
	for _, d := range diags {
		switch d.Level {
		case diag.LevelWarn:
			hasWarn = true
		case diag.LevelError:
			hasErr = true
		}
	}

	if hasErr {
		return 3
	}
	if hasWarn {
		if strict {
			return 3 // treat warnings as errors
		}
		return 2
	}
	return 0
}
