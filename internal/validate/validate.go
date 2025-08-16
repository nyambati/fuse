package validate

import (
	"sort"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
	"github.com/nyambati/fuse/internal/validate/validators"
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

	validators := []validators.Validator{
		validators.NewTeamValidator(proj.Teams),
		validators.NewFlowValidator(proj.Teams),
		validators.NewChannelsValidator(proj.Teams),
		validators.NewInhibitorsValidator(proj),
		validators.NewSilenceWindowsValidator(proj),
	}

	for _, v := range validators {
		diags = append(diags, v.Validate()...)
	}

	return diags
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
