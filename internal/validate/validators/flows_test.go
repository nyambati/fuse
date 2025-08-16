package validators_test

import (
	"testing"

	"github.com/nyambati/fuse/internal/dsl"
	"github.com/nyambati/fuse/internal/validate/validators"
	"github.com/stretchr/testify/assert"
)

func TestValidateFlows(t *testing.T) {
	tests := []struct {
		name     string
		team     dsl.Team
		wantCode []string
	}{
		{
			name: "valid flow",
			team: dsl.Team{
				Name: "payments",
				Channels: []dsl.Channel{
					{Name: "slack-payments"},
				},
				Flows: []dsl.Flow{
					{
						Notify: "slack-payments",
						When: []dsl.Matcher{
							{Label: "severity", Op: "=", Value: "critical"},
						},
					},
				},
			},
			wantCode: nil, // no errors/warnings
		},
		{
			name: "missing notify",
			team: dsl.Team{
				Name: "team1",
				Channels: []dsl.Channel{
					{Name: "slack-team1"},
				},
				Flows: []dsl.Flow{
					{
						When: []dsl.Matcher{
							{Label: "severity", Op: "=", Value: "critical"},
						},
					},
				},
			},
			wantCode: []string{"FLOW_NOTIFY_EMPTY"},
		},
		{
			name: "unknown notify channel",
			team: dsl.Team{
				Name:     "team2",
				Channels: []dsl.Channel{},
				Flows: []dsl.Flow{
					{
						Notify: "non-existent",
						When: []dsl.Matcher{
							{Label: "severity", Op: "=", Value: "critical"},
						},
					},
				},
			},
			wantCode: []string{"FLOW_NOTIFY_UNKNOWN"},
		},
		{
			name: "empty when block",
			team: dsl.Team{
				Name: "team3",
				Channels: []dsl.Channel{
					{Name: "slack-team3"},
				},
				Flows: []dsl.Flow{
					{
						Notify: "slack-team3",
					},
				},
			},
			wantCode: []string{"FLOW_WHEN_EMPTY"},
		},
		{
			name: "invalid matcher op",
			team: dsl.Team{
				Name: "team4",
				Channels: []dsl.Channel{
					{Name: "slack-team4"},
				},
				Flows: []dsl.Flow{
					{
						Notify: "slack-team4",
						When: []dsl.Matcher{
							{Label: "severity", Op: "badop", Value: "critical"},
						},
					},
				},
			},
			wantCode: []string{"FLOW_MATCHER_INVALID"},
		},
		{
			name: "invalid regex",
			team: dsl.Team{
				Name: "team5",
				Channels: []dsl.Channel{
					{Name: "slack-team5"},
				},
				Flows: []dsl.Flow{
					{
						Notify: "slack-team5",
						When: []dsl.Matcher{
							{Label: "instance", Op: "=~", Value: "("}, // invalid regex
						},
					},
				},
			},
			wantCode: []string{"FLOW_MATCHER_INVALID"},
		},
		{
			name: "duplicate flows",
			team: dsl.Team{
				Name: "team6",
				Channels: []dsl.Channel{
					{Name: "slack-team6"},
				},
				Flows: []dsl.Flow{
					{
						Notify: "slack-team6",
						When: []dsl.Matcher{
							{Label: "severity", Op: "=", Value: "critical"},
						},
					},
					{
						Notify: "slack-team6",
						When: []dsl.Matcher{
							{Label: "severity", Op: "=", Value: "critical"},
						},
					},
				},
			},
			wantCode: []string{"FLOW_DUPLICATE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validators.ValidateFlows(tt.team)
			var codes []string
			for _, d := range diags {
				codes = append(codes, d.Code)
			}
			for _, want := range tt.wantCode {
				assert.Contains(t, codes, want, "expected diagnostic code %q not found", want)
			}
			if tt.wantCode == nil {
				assert.Empty(t, diags, "expected no diagnostics")
			}
		})
	}
}
