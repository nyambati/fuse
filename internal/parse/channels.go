package parse

import (
	"fmt"
	"strings"

	"github.com/nyambati/fuse/internal/am"
	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/types"
)

// BuildReceivers maps team channels into AM receivers.
// Does not deduplicate — that’s handled later in validation.
// Channel.Type determines which AM config array gets populated.
func BuildReceivers(proj types.Project) ([]am.Receiver, []diag.Diagnostic) {
	var (
		receivers   []am.Receiver
		diagnostics []diag.Diagnostic
	)

	for _, team := range proj.Teams {
		for idx, channel := range team.Channels {
			receiver, diags := buildReceiver(team, idx, channel)
			if receiver != nil {
				receivers = append(receivers, *receiver)
			}
			diagnostics = append(diagnostics, diags...)
		}
	}

	return receivers, diagnostics
}

// buildReceiver processes a single channel and returns a receiver and any diagnostics.
func buildReceiver(team types.Team, idx int, channel types.Channel) (*am.Receiver, []diag.Diagnostic) {
	var diags []diag.Diagnostic

	// Validate and normalize the channel name
	name := strings.TrimSpace(channel.Name)
	if name == "" {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "CHAN_NAME_EMPTY",
			Message: fmt.Sprintf("team %q channel[%d] has empty name", team.Name, idx),
			File:    team.Path,
		})
		return nil, diags
	}

	receiver := am.Receiver{Name: name}

	if len(channel.Params) < 1 {
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelWarn,
			Code:    "CHAN_PARAMS_EMPTY",
			Message: fmt.Sprintf("%s channel %q in team %q has no params", channel.Type, name, team.Name),
		})
		return nil, diags
	}

	// Process channel configuration based on its type
	channelType := strings.ToLower(strings.TrimSpace(channel.Type))
	switch channelType {
	case "slack":
		receiver.SlackConfigs = []map[string]any{channel.Params}
	case "opsgenie":
		receiver.OpsgenieConfigs = []map[string]any{channel.Params}
	default:
		diags = append(diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "CHAN_TYPE_UNKNOWN",
			Message: fmt.Sprintf("unknown channel type %q for channel %q in team %q", channel.Type, name, team.Name),
		})
	}

	return &receiver, diags
}
