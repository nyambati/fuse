package validators

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

// validateFlows checks notify presence, channel existence, when block validity, and duplicates.
func ValidateFlows(team dsl.Team) []diag.Diagnostic {
	var diags []diag.Diagnostic

	// Build a set of channels in this team
	channelSet := make(map[string]struct{})
	for _, ch := range team.Channels {
		channelSet[ch.Name] = struct{}{}
	}

	// Track seen (notify, when) combinations for duplicate detection
	signatures := make(map[string]string)

	for idx, flow := range team.Flows {
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
		for _, m := range flow.When {
			if err := validateMatcher(m); err != nil {
				diags = append(diags, diag.Diagnostic{
					Level:   diag.LevelError,
					Code:    "FLOW_MATCHER_INVALID",
					Message: fmt.Sprintf("invalid matcher in flow (team %q): %v", team.Name, err),
				})
			}
		}

		// Create a signature for duplicate detection
		sig := flowSignature(flow)
		if prev, ok := signatures[sig]; ok {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelWarn,
				Code:    "FLOW_DUPLICATE",
				Message: fmt.Sprintf("duplicate flow matcher set and notify found for flow %s and %d", prev, idx+1),
			})
		} else {
			signatures[sig] = fmt.Sprintf("flow %d", idx+1)
		}
	}

	return diags
}

// validateMatcher checks that the matcher key/value is syntactically valid.
// This is a placeholder — expand with proper Alertmanager matcher parsing.
// validateMatcher checks syntax for our DSL `when` block matchers.
func validateMatcher(m dsl.Matcher) error {
	if m.Label == "" {
		return fmt.Errorf("matcher label is empty")
	}
	switch m.Op {
	case "=", "!=", "=~", "!~":
	default:
		return fmt.Errorf("matcher %q has invalid op %q", m.Label, m.Op)
	}
	if m.Value == "" {
		return fmt.Errorf("matcher %q has empty value", m.Label)
	}
	if m.Op == "=~" || m.Op == "!~" {
		if _, err := regexp.Compile(m.Value); err != nil {
			return fmt.Errorf("matcher %q has invalid regex %q: %v", m.Label, m.Value, err)
		}
	}
	return nil
}

func flowSignature(f dsl.Flow) string {
	// Sort matchers for deterministic comparison
	matcherCopy := append([]dsl.Matcher(nil), f.When...)
	sort.Slice(matcherCopy, func(i, j int) bool {
		if matcherCopy[i].Label != matcherCopy[j].Label {
			return matcherCopy[i].Label < matcherCopy[j].Label
		}
		if matcherCopy[i].Op != matcherCopy[j].Op {
			return matcherCopy[i].Op < matcherCopy[j].Op
		}
		return matcherCopy[i].Value < matcherCopy[j].Value
	})

	var matcherStrs []string
	for _, m := range matcherCopy {
		matcherStrs = append(matcherStrs, fmt.Sprintf("%s%s%s", m.Label, m.Op, m.Value))
	}

	return fmt.Sprintf("notify=%s;when=%v", f.Notify, matcherStrs)
}
