package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nyambati/fuse/internal/am"
	"github.com/nyambati/fuse/internal/types"
)

var timeRangeRe = regexp.MustCompile(`^\s*([0-2]?\d:[0-5]\d)\s*-\s*([0-2]?\d:[0-5]\d)\s*$`)

// BuildTimeIntervals maps global + team silence_windows into AM time_intervals.
// Team-level windows can shadow global ones with the same name (warn).
func BuildTimeIntervals(proj types.Project) ([]am.TimeIntervalSet, []types.Diagnostic) {
	var (
		sets  []am.TimeIntervalSet
		diags []types.Diagnostic
	)

	seen := map[string]string{} // name -> scope ("global" or "team/<name>")

	add := func(scope string, sw types.SilenceWindow) {
		name := strings.TrimSpace(sw.Name)
		if name == "" {
			diags = append(diags, types.Diagnostic{
				Level:   types.LevelError,
				Code:    "SW_NAME_EMPTY",
				Message: fmt.Sprintf("%s silence window has empty name", scope),
			})
			return
		}

		if !sw.Enabled {
			diags = append(diags, types.Diagnostic{
				Level:   types.LevelInfo,
				Code:    "SW_DISABLED",
				Message: fmt.Sprintf("silence window %q is disabled; skipping", name),
			})
			return
		}

		if prev, ok := seen[name]; ok && prev != scope {
			diags = append(diags, types.Diagnostic{
				Level:   types.LevelWarn,
				Code:    "SW_SHADOW",
				Message: fmt.Sprintf("silence window %q from %s shadows %s", name, scope, prev),
			})
		}
		seen[name] = scope

		ti := am.TimeInterval{
			Weekdays:    cloneSlice(sw.Weekdays),
			DaysOfMonth: cloneSlice(sw.DaysOfMonth),
			Months:      cloneSlice(sw.Months),
			Years:       cloneSlice(sw.Years),
			Location:    strings.TrimSpace(sw.Timezone),
		}

		// Parse "HH:MM-HH:MM"
		if strings.TrimSpace(sw.Time) != "" {
			m := timeRangeRe.FindStringSubmatch(sw.Time)
			if len(m) != 3 {
				diags = append(diags, types.Diagnostic{
					Level:   types.LevelError,
					Code:    "SW_TIME_FORMAT",
					Message: fmt.Sprintf("silence window %q has invalid time range %q (expected HH:MM-HH:MM)", name, sw.Time),
				})
				// continue building other fields; skip adding times
			} else {
				ti.Times = []struct {
					Start string "yaml:\"start_time\""
					End   string "yaml:\"end_time\""
				}{
					{Start: m[1], End: m[2]},
				}
			}
		}

		// Minimal guards: at least one selector or time must be present
		if len(ti.Weekdays) == 0 &&
			len(ti.DaysOfMonth) == 0 &&
			len(ti.Months) == 0 &&
			len(ti.Years) == 0 &&
			len(ti.Times) == 0 {
			diags = append(diags, types.Diagnostic{
				Level:   types.LevelWarn,
				Code:    "SW_EMPTY_INTERVAL",
				Message: fmt.Sprintf("silence window %q has no constraints (weekdays/days/months/years/time); it would match everything", name),
			})
		}

		sets = append(sets, am.TimeIntervalSet{
			Name:          name,
			TimeIntervals: []am.TimeInterval{ti},
		})
	}

	// Global first
	for _, sw := range proj.SilenceWindows {
		add("global", sw)
	}
	// Teams next
	for _, t := range proj.Teams {
		for _, sw := range t.SilenceWindows {
			add("team/"+t.Name, sw)
		}
	}

	return sets, diags
}

func cloneSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
