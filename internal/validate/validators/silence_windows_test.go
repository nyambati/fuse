package validators_test

import (
	"testing"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
	"github.com/nyambati/fuse/internal/validate/validators"
	"github.com/stretchr/testify/assert"
)

func TestValidateSilenceWindows_CoreRules(t *testing.T) {
	tests := []struct {
		name      string
		global    []dsl.SilenceWindow
		team      dsl.Team
		wantCodes []string
	}{
		{
			name: "valid global and team silence windows",
			global: []dsl.SilenceWindow{
				{Name: "night-shift", Time: "22:00-06:00"},
			},
			team: dsl.Team{
				Name: "payments",
				SilenceWindows: []dsl.SilenceWindow{
					{Name: "maintenance", Time: "01:00-02:00"},
				},
			},
			wantCodes: nil,
		},
		{
			name: "missing name",
			global: []dsl.SilenceWindow{
				{Name: "", Time: "22:00-06:00"},
			},
			team:      dsl.Team{},
			wantCodes: []string{"SILENCE_NO_NAME"},
		},
		{
			name: "duplicate in global",
			global: []dsl.SilenceWindow{
				{Name: "night-shift", Time: "22:00-06:00"},
				{Name: "night-shift", Time: "23:00-07:00"},
			},
			team:      dsl.Team{},
			wantCodes: []string{"SILENCE_DUP_NAME"},
		},
		{
			name: "team shadows global",
			global: []dsl.SilenceWindow{
				{Name: "night-shift", Time: "22:00-06:00"},
			},
			team: dsl.Team{
				Name: "payments",
				SilenceWindows: []dsl.SilenceWindow{
					{Name: "night-shift", Time: "23:00-07:00"},
				},
			},
			wantCodes: []string{"SILENCE_NAME_SHADOW"},
		},
		{
			name: "missing time",
			global: []dsl.SilenceWindow{
				{Name: "night-shift", Time: ""},
			},
			team:      dsl.Team{},
			wantCodes: []string{"SILENCE_NO_TIME"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := dsl.Project{
				SilenceWindows: tt.global,
				Teams:          []dsl.Team{tt.team},
			}

			diags := validators.NewSilenceWindowsValidator(proj).Validate()

			var gotCodes []string
			for _, d := range diags {
				gotCodes = append(gotCodes, d.Code)
				assert.Contains(t, []diag.Level{diag.LevelError, diag.LevelWarn}, d.Level)
			}
			assert.ElementsMatch(t, tt.wantCodes, gotCodes)
		})
	}
}
