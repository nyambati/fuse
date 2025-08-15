package validators_test

import (
	"testing"

	"github.com/nyambati/fuse/internal/dsl"
	"github.com/nyambati/fuse/internal/validate/validators"
	"github.com/stretchr/testify/assert"
)

func TestInhibitorValidator(t *testing.T) {
	tests := []struct {
		name   string
		proj   dsl.Project
		expect []string // expected diagnostic codes
	}{
		{
			name: "valid global and team inhibitors",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name: "suppress-warnings-when-critical",
						If: map[string]string{
							"severity": "critical",
						},
						Suppress: map[string]string{
							"severity": "warning",
						},
						When: []string{"alertname", "cluster"},
					},
				},
				Teams: []dsl.Team{
					{
						Name: "payments",
						Inhibitors: []dsl.Inhibitor{
							{
								Name: "team-inh",
								If:   map[string]string{"team": "payments"},
								Suppress: map[string]string{
									"team": "billing",
								},
								When: []string{"cluster"},
							},
						},
					},
				},
			},
			expect: nil,
		},
		{
			name: "missing name",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name:     "",
						If:       map[string]string{"severity": "critical"},
						Suppress: map[string]string{"severity": "warning"},
						When:     []string{"cluster"},
					},
				},
			},
			expect: []string{"INHIBITOR_NO_NAME"},
		},
		{
			name: "duplicate global inhibitor name",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name:     "dup",
						If:       map[string]string{"a": "b"},
						Suppress: map[string]string{"a": "c"},
						When:     []string{"x"},
					},
					{
						Name:     "dup",
						If:       map[string]string{"a": "b"},
						Suppress: map[string]string{"a": "c"},
						When:     []string{"x"},
					},
				},
			},
			expect: []string{"INHIBITOR_DUP_NAME"},
		},
		{
			name: "global shadows team",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name:     "shadowed",
						If:       map[string]string{"a": "b"},
						Suppress: map[string]string{"a": "c"},
						When:     []string{"x"},
					},
				},
				Teams: []dsl.Team{
					{
						Name: "team1",
						Inhibitors: []dsl.Inhibitor{
							{
								Name:     "shadowed",
								If:       map[string]string{"a": "b"},
								Suppress: map[string]string{"a": "c"},
								When:     []string{"x"},
							},
						},
					},
				},
			},
			expect: []string{"INHIBITOR_NAME_SHADOW"},
		},
		{
			name: "missing if",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name:     "no-if",
						If:       map[string]string{},
						Suppress: map[string]string{"a": "b"},
						When:     []string{"x"},
					},
				},
			},
			expect: []string{"INHIBITOR_NO_IF"},
		},
		{
			name: "missing suppress",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name:     "no-suppress",
						If:       map[string]string{"a": "b"},
						Suppress: map[string]string{},
						When:     []string{"x"},
					},
				},
			},
			expect: []string{"INHIBITOR_NO_SUPPRESS"},
		},
		{
			name: "missing when",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name:     "no-when",
						If:       map[string]string{"a": "b"},
						Suppress: map[string]string{"a": "b"},
						When:     []string{},
					},
				},
			},
			expect: []string{"INHIBITOR_NO_WHEN"},
		},
		{
			name: "duplicate when labels",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name:     "dup-when",
						If:       map[string]string{"a": "b"},
						Suppress: map[string]string{"a": "b"},
						When:     []string{"x", "x"},
					},
				},
			},
			expect: []string{"INHIBITOR_DUP_WHEN_LABEL"},
		},
		{
			name: "empty matcher key",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name:     "empty-key",
						If:       map[string]string{"": "val"},
						Suppress: map[string]string{"a": "b"},
						When:     []string{"x"},
					},
				},
			},
			expect: []string{"MATCH_EMPTY_KEY"},
		},
		{
			name: "empty matcher value",
			proj: dsl.Project{
				Inhibitors: []dsl.Inhibitor{
					{
						Name:     "empty-value",
						If:       map[string]string{"a": ""},
						Suppress: map[string]string{"a": "b"},
						When:     []string{"x"},
					},
				},
			},
			expect: []string{"MATCH_EMPTY_VALUE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validators.NewInhibitorsValidator(tt.proj).Validate()
			var gotCodes []string
			for _, d := range diags {
				gotCodes = append(gotCodes, d.Code)
			}
			for _, exp := range tt.expect {
				assert.Contains(t, gotCodes, exp, "expected code %s not found", exp)
			}
		})
	}
}
