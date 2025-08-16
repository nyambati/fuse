package validators

import (
	"fmt"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

type TeamValidator struct {
	teams []dsl.Team
}

func (v TeamValidator) Validate() []diag.Diagnostic {
	var diags []diag.Diagnostic

	seen := map[string]struct{}{}

	for _, t := range v.teams {
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
		seen[t.Name] = struct{}{}
	}

	return diags
}

func NewTeamValidator(teams []dsl.Team) Validator {
	return TeamValidator{teams: teams}
}
