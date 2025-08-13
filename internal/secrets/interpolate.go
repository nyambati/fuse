package secrets

import (
	"fmt"
	"regexp"
	"strings"
)

// Placeholder format: ${VAR_NAME}
// Allowed characters: letters, digits, underscore.
var placeholderRe = regexp.MustCompile(`\$\{([A-Za-z0-9_]+)\}`)

// InterpolateString replaces ${VAR} placeholders using the given Provider.
// It returns the interpolated string, a slice of missing variable names,
// and an error only for non-recoverable issues (regex, provider failure).
// Missing variables DO NOT produce an error; they are collected and left intact
// so the caller can decide whether to warn or fail (e.g., --strict).
func InterpolateString(s string, p Provider) (string, []string, error) {
	if s == "" {
		return s, nil, nil
	}
	missing := []string{}

	out := placeholderRe.ReplaceAllStringFunc(s, func(m string) string {
		key := placeholderRe.FindStringSubmatch(m)[1]
		val, err := p.Resolve(key)
		if err == nil {
			return val
		}
		if err == ErrNotFound {
			missing = append(missing, key)
			return m // keep placeholder intact
		}
		// Other provider errors: keep placeholder and bubble up via caller (we track via a marker)
		missing = append(missing, key)
		return m
	})

	return out, missing, nil
}

// InterpolateMapString applies InterpolateString to every value in a map[string]string.
// It returns the updated map (copied), a combined list of missing keys, and error.
func InterpolateMapString(in map[string]string, p Provider) (map[string]string, []string, error) {
	if len(in) == 0 {
		return map[string]string{}, nil, nil
	}
	out := make(map[string]string, len(in))
	var missing []string

	for k, v := range in {
		iv, miss, err := InterpolateString(v, p)
		if err != nil {
			// propagate error, but still collect what we have so far
			out[k] = v
			missing = append(missing, miss...)
			return out, missing, err
		}
		out[k] = iv
		if len(miss) > 0 {
			missing = append(missing, miss...)
		}
	}
	// Deduplicate missing
	if len(missing) > 1 {
		missing = dedup(missing)
	}
	return out, missing, nil
}

func dedup(ss []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// HasPlaceholders returns true if the string contains any ${VAR} placeholders.
func HasPlaceholders(s string) bool {
	return placeholderRe.MatchString(s)
}

// ListPlaceholders returns unique placeholder keys contained in s.
func ListPlaceholders(s string) []string {
	matches := placeholderRe.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}
	keys := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			keys = append(keys, m[1])
		}
	}
	return dedup(keys)
}

// InterpolateSlice applies InterpolateString to each item in a slice.
func InterpolateSlice(in []string, p Provider) ([]string, []string, error) {
	if len(in) == 0 {
		return nil, nil, nil
	}
	out := make([]string, len(in))
	var missing []string
	for i, v := range in {
		iv, miss, err := InterpolateString(v, p)
		if err != nil {
			out[i] = v
			missing = append(missing, miss...)
			return out, dedup(missing), err
		}
		out[i] = iv
		missing = append(missing, miss...)
	}
	return out, dedup(missing), nil
}

// InterpolateStructFields is a helper to interpolate known common fields often used in channels.
// You can expand this as needed in parse layer.
type ChannelLike struct {
	WebhookURL string
	URL        string
	RoutingKey string
	ServiceKey string
	Email      string
	Phone      string
	Extra      map[string]string
}

// InterpolateChannelFields applies interpolation to common fields plus Extra.
func InterpolateChannelFields(c ChannelLike, p Provider) (ChannelLike, []string, error) {
	var allMissing []string
	var err error

	if c.WebhookURL, allMissing, err = interpolateField(c.WebhookURL, allMissing, p); err != nil {
		return c, allMissing, err
	}
	if c.URL, allMissing, err = interpolateField(c.URL, allMissing, p); err != nil {
		return c, allMissing, err
	}
	if c.RoutingKey, allMissing, err = interpolateField(c.RoutingKey, allMissing, p); err != nil {
		return c, allMissing, err
	}
	if c.ServiceKey, allMissing, err = interpolateField(c.ServiceKey, allMissing, p); err != nil {
		return c, allMissing, err
	}
	if c.Email, allMissing, err = interpolateField(c.Email, allMissing, p); err != nil {
		return c, allMissing, err
	}
	if c.Phone, allMissing, err = interpolateField(c.Phone, allMissing, p); err != nil {
		return c, allMissing, err
	}

	if c.Extra != nil {
		var miss []string
		c.Extra, miss, err = InterpolateMapString(c.Extra, p)
		if err != nil {
			return c, append(allMissing, miss...), err
		}
		allMissing = append(allMissing, miss...)
	}

	if len(allMissing) > 1 {
		allMissing = dedup(allMissing)
	}
	return c, allMissing, nil
}

func interpolateField(val string, acc []string, p Provider) (string, []string, error) {
	iv, miss, err := InterpolateString(val, p)
	if err != nil {
		return val, append(acc, miss...), err
	}
	return iv, append(acc, miss...), nil
}

// Redact replaces any ${VAR} pattern with "***" without resolving it.
// Useful for logs or --redact in build step.
func Redact(s string) string {
	return placeholderRe.ReplaceAllString(s, "***")
}

// RedactMap applies Redact to all values.
func RedactMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = Redact(v)
	}
	return out
}

// ExpandError is a convenience to pretty-print missing variables.
func ExpandError(missing []string) error {
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing secrets: %s", strings.Join(dedup(missing), ", "))
}
