package validators

import (
	"fmt"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

type EmailValidator struct{}

func (EmailValidator) Validate(ch dsl.Channel) []diag.Diagnostic {
	var diags []diag.Diagnostic
	for _, cfg := range ch.Configs {
		to, ok := cfg["to"].([]any)
		if !ok || len(to) == 0 {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "CHANNEL_EMAIL_NO_TO",
				Message: fmt.Sprintf("email channel %q missing 'to' list", ch.Name),
			})
		}
	}
	return diags
}

func init() {
	RegisterChannelValidator("email", EmailValidator{})
}
