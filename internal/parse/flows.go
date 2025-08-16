package parse

import (
	"fmt"

	"github.com/nyambati/fuse/internal/am"
	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

// BuildFlowRoutes converts all teams' flows into a flat slice of AM routes.
// Caller typically wraps these under a single root route:

func BuildFlowRoutes(proj dsl.Project) (am.Route, []diag.Diagnostic) {
	var diags []diag.Diagnostic

	for _, team := range proj.Teams {
		for idx, f := range team.Flows {
			rs, d := mapFlowToRoutes(team, idx, f)
			if len(d) > 0 {
				diags = append(diags, d...)
			}
			proj.RootRoute.Routes = append(proj.RootRoute.Routes, rs...)
		}
	}

	return proj.RootRoute, diags
}

func mapFlowToRoutes(team dsl.Team, idx int, f dsl.Flow) ([]am.Route, []diag.Diagnostic) {
	var (
		routes []am.Route
		diags  []diag.Diagnostic
	)

	// ---- notify must exist ----
	if f.Notify == "" {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "FLOW_NOTIFY_EMPTY",
			Message: fmt.Sprintf("flows[%d] has no notify target(s)", idx),
			File:    team.Path, // we don't have exact file/line yet
		})
		return routes, diags
	}

	// ---- matchers from when ----
	matchers, mDiags := ToMatchers(f.When)
	if len(mDiags) > 0 {
		diags = append(diags, mDiags...)
	}

	r := am.Route{
		Receiver:       f.Notify,
		GroupBy:        append([]string{}, f.GroupBy...),
		GroupWait:      f.WaitFor,
		GroupInterval:  f.GroupInterval,
		RepeatInterval: f.RepeatAfter,
		Matchers:       matchers,
		TimeIntervals:  append([]string{}, f.SilenceWhen...),
	}

	if f.Continue != nil {
		r.Continue = *f.Continue
	}

	routes = append(routes, r)

	return routes, diags
}
