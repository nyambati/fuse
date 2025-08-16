package secrets

import (
	"os"
	"strings"
)

// EnvProvider resolves secrets from environment variables.
type EnvProvider struct{}

func (p *EnvProvider) Resolve(key string) (string, error) {
	if v, ok := os.LookupEnv(strings.ToUpper(strings.TrimSpace(key))); ok {
		return v, nil
	}
	return "", ErrNotFound
}
