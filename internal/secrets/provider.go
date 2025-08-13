package secrets

import "fmt"

// ErrNotFound is returned when a secret placeholder cannot be resolved by a provider.
var ErrNotFound = fmt.Errorf("secret not found")

// Provider resolves secret keys (e.g., "SLACK_WEBHOOK").
type Provider interface {
	// Resolve returns the secret value for key.
	// If the key is not present, return ErrNotFound.
	Resolve(key string) (string, error)
}

// NewProvider creates a Provider based on name and an optional config path.
// Supported names (MVP): "env"
// Stubs for later: "sops", "vault", "ssm"
func NewProvider(name, configPath string) (Provider, error) {
	switch name {
	case "", "env":
		return &EnvProvider{}, nil
	case "sops":
		return nil, fmt.Errorf("secrets provider %q not implemented yet", name)
	case "vault":
		return nil, fmt.Errorf("secrets provider %q not implemented yet", name)
	case "ssm":
		return nil, fmt.Errorf("secrets provider %q not implemented yet", name)
	default:
		return nil, fmt.Errorf("unknown secrets provider %q", name)
	}
}
