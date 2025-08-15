package validators_test

import (
	"testing"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
	"github.com/nyambati/fuse/internal/validate/validators"
	"github.com/stretchr/testify/assert"
)

func TestValidateChannels(t *testing.T) {
	tests := []struct {
		name     string
		channels []dsl.Channel
		wantErrs []string // expected diag codes
	}{
		// core rules + slack
		{
			name: "valid single slack channel",
			channels: []dsl.Channel{
				{Name: "payments-slack", Type: "slack"},
			},
			wantErrs: nil,
		},
		{
			name: "missing name",
			channels: []dsl.Channel{
				{Name: "", Type: "slack"},
			},
			wantErrs: []string{"CHANNEL_NO_NAME"},
		},
		{
			name: "duplicate name",
			channels: []dsl.Channel{
				{Name: "payments-slack", Type: "slack"},
				{Name: "payments-slack", Type: "slack"},
			},
			wantErrs: []string{"CHANNEL_DUP_NAME"},
		},
		{
			name: "missing type",
			channels: []dsl.Channel{
				{Name: "payments-slack", Type: ""},
			},
			wantErrs: []string{"CHANNEL_NO_TYPE"},
		},
		{
			name: "unknown type",
			channels: []dsl.Channel{
				{Name: "payments-slack", Type: "pagerduty"},
			},
			wantErrs: []string{"CHANNEL_UNKNOWN_TYPE"},
		},

		// add more tests for channels here
		{
			name: "valid single email channel",
			channels: []dsl.Channel{
				{Name: "payments-email", Type: "email"},
			},
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			teams := []dsl.Team{
				{Name: "payments", Channels: tt.channels},
			}

			diags := validators.NewChannelsValidator(teams).Validate()

			var gotCodes []string
			for _, d := range diags {
				gotCodes = append(gotCodes, d.Code)
			}

			assert.ElementsMatch(t, tt.wantErrs, gotCodes)
			for _, d := range diags {
				assert.Contains(t, []diag.Level{diag.LevelError, diag.LevelWarn}, d.Level)
			}
		})
	}
}
