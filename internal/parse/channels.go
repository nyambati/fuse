package parse

import (
	"fmt"
	"strings"

	"github.com/nyambati/fuse/internal/am"
	"github.com/nyambati/fuse/internal/types"
)

// BuildReceivers maps team channels into AM receivers.
// Does not deduplicate — that’s handled later in validation.
// Channel.Type determines which AM config array gets populated.
func BuildReceivers(proj types.Project) ([]am.Receiver, []types.Diagnostic) {
	var (
		out   []am.Receiver
		diags []types.Diagnostic
	)

	for _, team := range proj.Teams {
		for idx, ch := range team.Channels {
			name := strings.TrimSpace(ch.Name)
			if name == "" {
				diags = append(diags, types.Diagnostic{
					Level:   types.LevelError,
					Code:    "CHAN_NAME_EMPTY",
					Message: fmt.Sprintf("team %q channel[%d] has empty name", team.Name, idx),
					File:    team.Path,
				})
				continue
			}

			r := am.Receiver{Name: name}
			switch strings.ToLower(strings.TrimSpace(ch.Type)) {
			case "slack":
				cfg := map[string]any{}
				if ch.WebhookURL != "" {
					cfg["api_url"] = ch.WebhookURL
				}
				for k, v := range ch.Params {
					cfg[k] = v
				}
				if len(cfg) > 0 {
					r.SlackConfigs = []map[string]any{cfg}
				} else {
					diags = append(diags, types.Diagnostic{
						Level:   types.LevelWarn,
						Code:    "CHAN_SLACK_EMPTY",
						Message: fmt.Sprintf("slack channel %q in team %q has no webhook_url or params", name, team.Name),
					})
				}

			case "pagerduty":
				cfg := map[string]any{}
				if ch.RoutingKey != "" {
					cfg["routing_key"] = ch.RoutingKey
				}
				for k, v := range ch.Params {
					cfg[k] = v
				}
				if len(cfg) > 0 {
					r.PagerdutyConfig = []map[string]any{cfg}
				}

			case "webhook":
				cfg := map[string]any{}
				if ch.URL != "" {
					cfg["url"] = ch.URL
				}
				for k, v := range ch.Params {
					cfg[k] = v
				}
				if len(cfg) > 0 {
					r.WebhookConfigs = []map[string]any{cfg}
				}

			case "email":
				cfg := map[string]any{}
				if ch.Email != "" {
					cfg["to"] = ch.Email
				}
				for k, v := range ch.Params {
					cfg[k] = v
				}
				if len(cfg) > 0 {
					r.EmailConfigs = []map[string]any{cfg}
				}

			default:
				diags = append(diags, types.Diagnostic{
					Level:   types.LevelError,
					Code:    "CHAN_TYPE_UNKNOWN",
					Message: fmt.Sprintf("unknown channel type %q for channel %q in team %q", ch.Type, name, team.Name),
				})
			}

			out = append(out, r)
		}
	}

	return out, diags
}
