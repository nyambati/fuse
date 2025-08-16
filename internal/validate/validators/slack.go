package validators

import (
	"fmt"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

type SlackValidator struct{}

func (SlackValidator) Validate(ch dsl.Channel) []diag.Diagnostic {
	var diags []diag.Diagnostic
	for _, cfg := range ch.Configs {
		channel, ok := cfg["channel"].(string)
		if !ok || channel == "" {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "CHANNEL_SLACK_NO_CHANNEL",
				Message: fmt.Sprintf("slack channel %q missing channel", ch.Name),
			})
		}
	}
	return diags
}

func init() {
	RegisterChannelValidator("slack", SlackValidator{})
}
