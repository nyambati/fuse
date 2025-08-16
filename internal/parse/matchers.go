package parse

import (
	"fmt"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
)

func ToMatchers(when []dsl.Matcher) ([]string, []diag.Diagnostic) {
	if len(when) == 0 {
		return nil, nil
	}
	var (
		out   = make([]string, 0, len(when))
		diags []diag.Diagnostic
	)

	for _, matcher := range when {
		m := fmt.Sprintf("%s %s \"%s\"", matcher.Label, matcher.Op, matcher.Value)
		if m != "" {
			out = append(out, m)
		}
	}
	return out, diags
}
