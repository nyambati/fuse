package types

//
// ===== v0.1 DSL Types (skeleton) =====
//

// Project is the in-memory representation of a Fuse project DSL.
type Project struct {
	Root           string
	Global         Global
	SilenceWindows []SilenceWindow
	Inhibitors     []Inhibitor
	Teams          []Team
}

// Global mirrors Alertmanager's global section (keep flexible for now).
type Global struct {
	// Raw holds decoded YAML for future mapping (keep permissive).
	Raw map[string]any
}

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
	Name   string            `yaml:"name"`
	Type   string            `yaml:"type"`
	Params map[string]string `yaml:"params,omitempty"` // generic bag; specific keys (e.g., webhook_url) can also be top-level later
	// Common well-known fields kept for convenience (optional at this stage)
	WebhookURL  string `yaml:"webhook_url,omitempty"`
	RoutingKey  string `yaml:"routing_key,omitempty"`
	URL         string `yaml:"url,omitempty"`
	ServiceKey  string `yaml:"service_key,omitempty"`
	Email       string `yaml:"email,omitempty"`
	PhoneNumber string `yaml:"phone_number,omitempty"`
}

// Flow is a single routing rule inside flows.yaml.
type Flow struct {
	Notify        []string          // normalized: always a slice (string in YAML expands to 1 item)
	When          map[string]string `yaml:"when"`
	GroupBy       []string          `yaml:"group_by,omitempty"`
	WaitFor       string            `yaml:"wait_for,omitempty"`
	GroupInterval string            `yaml:"group_interval,omitempty"`
	RepeatAfter   string            `yaml:"repeat_after,omitempty"`
	MuteWhen      []string          `yaml:"mute_when,omitempty"`
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
