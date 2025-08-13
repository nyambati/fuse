package secrets

import "os"

// EnvProvider resolves secrets from environment variables.
type EnvProvider struct{}

func (p *EnvProvider) Resolve(key string) (string, error) {
	if v, ok := os.LookupEnv(key); ok {
		return v, nil
	}
	return "", ErrNotFound
}
