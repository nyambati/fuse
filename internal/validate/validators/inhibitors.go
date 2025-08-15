package validators

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

type InhibitorValidator struct {
	project dsl.Project
}

func NewInhibitorsValidator(proj dsl.Project) Validator {
	return InhibitorValidator{project: proj}
}

func (v InhibitorValidator) Validate() []diag.Diagnostic {
	var diags []diag.Diagnostic

	// Track global names for uniqueness
	globalNames := map[string]struct{}{}
	for _, inh := range v.project.Inhibitors {
		diags = append(diags, validateOneInhibitor(inh, "global", globalNames)...)
	}

	// Validate team inhibitors
	for _, t := range v.project.Teams {
		teamNames := map[string]struct{}{}
		for _, inh := range t.Inhibitors {
			if _, exists := globalNames[strings.TrimSpace(inh.Name)]; exists && strings.TrimSpace(inh.Name) != "" {
				diags = append(diags, diag.Diagnostic{
					Level:   diag.LevelWarn,
					Code:    "INHIBITOR_NAME_SHADOW",
					Message: fmt.Sprintf("team %q inhibitor %q shadows a global inhibitor", t.Name, inh.Name),
					File:    t.Path,
				})
			}
			diags = append(diags, validateOneInhibitor(inh, t.Name, teamNames)...)
		}
	}

	return diags
}

func validateOneInhibitor(inh dsl.Inhibitor, scope string, seen map[string]struct{}) []diag.Diagnostic {
	var diags []diag.Diagnostic

	// --- Name ---
	name := strings.TrimSpace(inh.Name)
	if name == "" {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "INHIBITOR_NO_NAME",
			Message: fmt.Sprintf("inhibitor in %s has no name", scope),
		})
	} else {
		if _, exists := seen[name]; exists {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "INHIBITOR_DUP_NAME",
				Message: fmt.Sprintf("duplicate inhibitor %q in %s", name, scope),
			})
		}
		seen[name] = struct{}{}
	}

	// --- If matchers ---
	if len(inh.If) == 0 {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "INHIBITOR_NO_IF",
			Message: fmt.Sprintf("inhibitor %q in %s has no 'if' matchers", name, scope),
		})
	} else {
		diags = append(diags, validateInhibitorMatchers(inh.If, name, scope, "if")...)
	}

	// --- Suppress matchers ---
	if len(inh.Suppress) == 0 {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "INHIBITOR_NO_SUPPRESS",
			Message: fmt.Sprintf("inhibitor %q in %s has no 'suppress' matchers", name, scope),
		})
	} else {
		diags = append(diags, validateInhibitorMatchers(inh.Suppress, name, scope, "suppress")...)
	}

	// --- When labels ---
	if len(inh.When) == 0 {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "INHIBITOR_NO_WHEN",
			Message: fmt.Sprintf("inhibitor %q in %s has no 'when' labels", name, scope),
		})
	} else {
		seenLabels := map[string]struct{}{}
		for _, lbl := range inh.When {
			l := strings.TrimSpace(lbl)
			if l == "" {
				diags = append(diags, diag.Diagnostic{
					Level:   diag.LevelError,
					Code:    "INHIBITOR_EMPTY_WHEN_LABEL",
					Message: fmt.Sprintf("inhibitor %q in %s has an empty label in 'when'", name, scope),
				})
				continue
			}
			if _, exists := seenLabels[l]; exists {
				diags = append(diags, diag.Diagnostic{
					Level:   diag.LevelError,
					Code:    "INHIBITOR_DUP_WHEN_LABEL",
					Message: fmt.Sprintf("inhibitor %q in %s has duplicate label %q in 'when'", name, scope, l),
				})
			}
			seenLabels[l] = struct{}{}
		}
	}

	return diags
}

func validateInhibitorMatchers(m map[string]string, inhName, scope, field string) []diag.Diagnostic {
	var diags []diag.Diagnostic
	for k, v := range m {
		key := strings.TrimSpace(k)
		val := strings.TrimSpace(v)

		if key == "" {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "MATCH_EMPTY_KEY",
				Message: fmt.Sprintf("inhibitor %q in %s has empty key in '%s' matchers", inhName, scope, field),
			})
		}
		if val == "" {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "MATCH_EMPTY_VALUE",
				Message: fmt.Sprintf("inhibitor %q in %s has empty value for key %q in '%s' matchers", inhName, scope, key, field),
			})
		}

		// Optional regex validation if we allow it in the future
		if strings.HasPrefix(val, "~") {
			if _, err := regexp.Compile(val[1:]); err != nil {
				diags = append(diags, diag.Diagnostic{
					Level:   diag.LevelError,
					Code:    "MATCH_REGEX_INVALID",
					Message: fmt.Sprintf("inhibitor %q in %s has invalid regex for key %q in '%s': %v", inhName, scope, key, field, err),
				})
			}
		}
	}
	return diags
}
