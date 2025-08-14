package types

import "github.com/nyambati/fuse/internal/am"

//
// ===== v0.1 DSL Types (skeleton) =====
//

// Project is the in-memory representation of a Fuse project DSL.
type Project struct {
	Root           string
	Global         Global
	RootRoute      *am.Route
	SilenceWindows []SilenceWindow
	Inhibitors     []Inhibitor
	Teams          []Team
}

// Global mirrors Alertmanager's global section (keep flexible for now).
type Global map[string]any

// SilenceWindow defines a named recurring mute period.
type SilenceWindow struct {
	Name        string   `yaml:"name"`
	Time        string   `yaml:"time"`
	Enabled     bool     `yaml:"enabled"`
	Weekdays    []string `yaml:"weekdays"`
	DaysOfMonth []string `yaml:"days_of_month"`
	Months      []string `yaml:"months"`
	Years       []string `yaml:"years"`
	Timezone    string   `yaml:"timezone"`
}

// Channel represents a notification destination.
type Channel struct {
	Name    string           `yaml:"name"`
	Type    string           `yaml:"type"`
	Configs []map[string]any `yaml:"configs,omitempty"`
}

// Flow is a single routing rule inside flows.yaml.
type Flow struct {
	Notify        string            // normalized: always a slice (string in YAML expands to 1 item)
	When          map[string]string `yaml:"when"`
	GroupBy       []string          `yaml:"group_by,omitempty"`
	WaitFor       string            `yaml:"wait_for,omitempty"`
	GroupInterval string            `yaml:"group_interval,omitempty"`
	RepeatAfter   string            `yaml:"repeat_after,omitempty"`
	SilenceWhen   []string          `yaml:"silence_when,omitempty"`
	Continue      *bool             `yaml:"continue,omitempty"`
}

// Inhibitor represents a simplified inhibit rule.
type Inhibitor struct {
	Name     string            `yaml:"name"`
	If       map[string]string `yaml:"if"`
	Suppress map[string]string `yaml:"suppress"`
	When     []string          `yaml:"when"`
}

// Team collects a team's DSL files.
type Team struct {
	Name           string
	Path           string
	Channels       []Channel
	Flows          []Flow
	SilenceWindows []SilenceWindow
	// Future: Alerts, Templates references, etc.
}
