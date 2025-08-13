package validate

import (
	"fmt"
	"sort"

	"github.com/nyambati/fuse/internal/types"
)

// Options for validation behavior
type Options struct {
	Strict bool
}

// Project runs semantic validation on a loaded DSL project and the derived AM config.
// amc is intentionally typed as any to avoid import cycles; richer checks can be added later.
func Project(proj types.Project, amc any, opts Options) []types.Diagnostic {
	var diags []types.Diagnostic

	// ---- Basic project-level checks (skeleton) ----

	// Ensure project root is set.
	if proj.Root == "" {
		diags = append(diags, types.Diagnostic{
			Level:   types.LevelError,
			Code:    "PROJ_ROOT_EMPTY",
			Message: "project root not set",
		})
	}

	// Teams must have unique names (discovery should guarantee, but double-check).
	seen := map[string]struct{}{}
	for _, t := range proj.Teams {
		if t.Name == "" {
			diags = append(diags, types.Diagnostic{
				Level:   types.LevelWarn,
				Code:    "TEAM_NAME_EMPTY",
				Message: "a team folder has an empty name",
				File:    t.Path,
			})
			continue
		}
		if _, ok := seen[t.Name]; ok {
			diags = append(diags, types.Diagnostic{
				Level:   types.LevelError,
				Code:    "TEAM_NAME_DUP",
				Message: fmt.Sprintf("duplicate team name %q", t.Name),
				File:    t.Path,
			})
		}
		seen[t.Name] = struct{}{}
	}

	// TODO (next steps):
	// - Validate silence_windows names uniqueness (global vs team shadowing -> warn)
	// - Validate channels: unique names within a team, required params per type
	// - Validate flows: notify exists, references to channels & silence_windows resolve
	// - Validate inhibitors: fields present, matcher syntax sanity
	// - Time/duration parsing checks for wait/group/repeat

	return diags
}

// Merge combines multiple diagnostic slices and sorts them
func Merge(diags ...[]types.Diagnostic) []types.Diagnostic {
	var all []types.Diagnostic
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
func ExitCode(diags []types.Diagnostic, strict bool) int {
	var hasWarn, hasErr bool
	for _, d := range diags {
		switch d.Level {
		case types.LevelWarn:
			hasWarn = true
		case types.LevelError:
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
