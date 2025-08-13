package am

import "github.com/nyambati/fuse/internal/diag"

// "bytes"
// "os/exec"

// CheckWithAmtool optionally validates the rendered Alertmanager config with amtool.
// MVP: if amtoolPath == "", return no diagnostics. We'll implement this later.
func CheckWithAmtool(_ Config, amtoolPath string) []diag.Diagnostic {
	if amtoolPath == "" {
		return nil
	}
	// Placeholder stub; real implementation will:
	// - marshal Config to YAML
	// - exec amtool check-config <tmpfile>
	// - parse stdout/stderr into diagnostics
	return nil
}
