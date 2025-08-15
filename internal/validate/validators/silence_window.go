package validators

import (
	"fmt"
	"strings"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

// SilenceWindowsValidator validates silence windows in the project.
type SilenceWindowsValidator struct {
	project dsl.Project
}

// NewSilenceWindowsValidator creates a new SilenceWindowsValidator.
func NewSilenceWindowsValidator(proj dsl.Project) Validator {
	return SilenceWindowsValidator{project: proj}
}

// Validate runs all silence window checks for global and team configs.
func (v SilenceWindowsValidator) Validate() []diag.Diagnostic {
	var diags []diag.Diagnostic

	// Track global names for uniqueness and shadowing detection
	globalNames := map[string]struct{}{}
	for _, sw := range v.project.SilenceWindows {
		diags = append(diags, validateOneSilenceWindow(sw, "global", globalNames)...)
	}

	// Validate each team's silence windows
	for _, t := range v.project.Teams {
		teamNames := map[string]struct{}{}
		for _, sw := range t.SilenceWindows {
			if _, exists := globalNames[sw.Name]; exists && strings.TrimSpace(sw.Name) != "" {
				diags = append(diags, diag.Diagnostic{
					Level:   diag.LevelWarn,
					Code:    "SILENCE_NAME_SHADOW",
					Message: fmt.Sprintf("team %q silence window %q shadows a global silence window", t.Name, sw.Name),
					File:    t.Path,
				})
			}
			diags = append(diags, validateOneSilenceWindow(sw, t.Name, teamNames)...)
		}
	}

	return diags
}

func validateOneSilenceWindow(sw dsl.SilenceWindow, scope string, seen map[string]struct{}) []diag.Diagnostic {
	var diags []diag.Diagnostic

	name := strings.TrimSpace(sw.Name)
	if name == "" {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "SILENCE_NO_NAME",
			Message: fmt.Sprintf("silence window in %s has no name", scope),
		})
	} else {
		if _, exists := seen[name]; exists {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "SILENCE_DUP_NAME",
				Message: fmt.Sprintf("duplicate silence window %q in %s", name, scope),
			})
		}
		seen[name] = struct{}{}
	}

	if strings.TrimSpace(sw.Time) == "" {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "SILENCE_NO_TIME",
			Message: fmt.Sprintf("silence window %q in %s has no time", name, scope),
		})
	}

	return diags
}
