package validators

import (
	"fmt"
	"strings"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

type ChannelValidator interface {
	Validate(ch dsl.Channel) []diag.Diagnostic
}

var channelValidators = map[string]ChannelValidator{}

func RegisterChannelValidator(channelType string, v ChannelValidator) {
	channelValidators[channelType] = v
}

func ValidateChannels(teamName string, channels []dsl.Channel) []diag.Diagnostic {
	var diags []diag.Diagnostic
	seen := map[string]struct{}{}

	for _, ch := range channels {
		// --- Core: name required ---
		if strings.TrimSpace(ch.Name) == "" {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "CHANNEL_NO_NAME",
				Message: fmt.Sprintf("channel in team %q has no name", teamName),
			})
			continue
		}

		// --- Core: unique names ---
		if _, exists := seen[ch.Name]; exists {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "CHANNEL_DUP_NAME",
				Message: fmt.Sprintf("duplicate channel name %q in team %q", ch.Name, teamName),
			})
		}

		seen[ch.Name] = struct{}{}

		// --- Core: type required ---
		if strings.TrimSpace(ch.Type) == "" {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "CHANNEL_NO_TYPE",
				Message: fmt.Sprintf("channel %q in team %q has no type", ch.Name, teamName),
			})
			continue
		}

		// --- Type-specific validation ---
		v, ok := channelValidators[ch.Type]
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "CHANNEL_UNKNOWN_TYPE",
				Message: fmt.Sprintf("channel %q in team %q has unknown type %q", ch.Name, teamName, ch.Type),
			})
			continue
		}

		diags = append(diags, v.Validate(ch)...)
	}

	return diags
}
