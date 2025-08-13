package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nyambati/fuse/internal/diag"
)

// ToMatchers converts a Fuse `when:` map into Alertmanager matcher strings.
// Option D rules:
// - value with no prefix        -> equals     (key="value")
// - value with "!" prefix       -> not equals (key!="value")
// - value with "~" prefix       -> regex      (key=~"pattern")
// - value with "!~" prefix      -> regex-not  (key!~"pattern")
func ToMatchers(when map[string]string) ([]string, []diag.Diagnostic) {
	if len(when) == 0 {
		return nil, nil
	}
	var (
		out   = make([]string, 0, len(when))
		diags []diag.Diagnostic
	)
	for k, raw := range when {
		key := strings.TrimSpace(k)
		if key == "" {
			diags = append(diags, diag.Diagnostic{
				Level:   diag.LevelError,
				Code:    "MATCH_EMPTY_KEY",
				Message: "matcher key is empty",
			})
			continue
		}

		m, d := parseOneMatcher(key, raw)
		if len(d) > 0 {
			diags = append(diags, d...)
		}
		if m != "" {
			out = append(out, m)
		}
	}
	return out, diags
}

// parseOneMatcher applies Option‑D prefix rules to a single key/value.
func parseOneMatcher(key, val string) (string, []diag.Diagnostic) {
	val = strings.TrimSpace(val)
	if val == "" {
		return "", []diag.Diagnostic{{
			Level:   diag.LevelError,
			Code:    "MATCH_EMPTY_VALUE",
			Message: fmt.Sprintf("matcher %q has empty value", key),
		}}
	}

	switch {
	case strings.HasPrefix(val, "!~"):
		pat := strings.TrimSpace(strings.TrimPrefix(val, "!~"))
		if err := validateRegex(pat); err != nil {
			return "", []diag.Diagnostic{{
				Level:   diag.LevelError,
				Code:    "MATCH_REGEX_INVALID",
				Message: fmt.Sprintf("invalid regex for %q: %v", key, err),
			}}
		}
		return fmt.Sprintf(`%s!~"%s"`, key, pat), nil

	case strings.HasPrefix(val, "~"):
		pat := strings.TrimSpace(strings.TrimPrefix(val, "~"))
		if err := validateRegex(pat); err != nil {
			return "", []diag.Diagnostic{{
				Level:   diag.LevelError,
				Code:    "MATCH_REGEX_INVALID",
				Message: fmt.Sprintf("invalid regex for %q: %v", key, err),
			}}
		}
		return fmt.Sprintf(`%s=~"%s"`, key, pat), nil

	case strings.HasPrefix(val, "!"):
		v := strings.TrimSpace(strings.TrimPrefix(val, "!"))
		if v == "" {
			return "", []diag.Diagnostic{{
				Level:   diag.LevelError,
				Code:    "MATCH_NOT_EMPTY",
				Message: fmt.Sprintf("matcher %q uses '!' but has no value", key),
			}}
		}
		return fmt.Sprintf(`%s!="%s"`, key, escapeQuotes(v)), nil

	default:
		// equals
		return fmt.Sprintf(`%s="%s"`, key, escapeQuotes(val)), nil
	}
}

func validateRegex(p string) error {
	// Allow empty? No — treat empty as invalid to avoid surprising catch-alls.
	if strings.TrimSpace(p) == "" {
		return fmt.Errorf("empty pattern")
	}
	_, err := regexp.Compile(p)
	return err
}

func escapeQuotes(s string) string {
	// Minimal escaping for AM matcher string form.
	return strings.ReplaceAll(s, `"`, `\"`)
}
