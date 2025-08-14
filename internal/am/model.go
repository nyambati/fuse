package am

type Config struct {
	Global        map[string]any    `yaml:"global,omitempty"`
	Receivers     []Receiver        `yaml:"receivers,omitempty"`
	Route         Route             `yaml:"route,omitempty"`
	InhibitRules  []InhibitRule     `yaml:"inhibit_rules,omitempty"`
	TimeIntervals []TimeIntervalSet `yaml:"time_intervals,omitempty"`
}

type Receiver struct {
	Name            string           `yaml:"name"`
	SlackConfigs    []map[string]any `yaml:"slack_configs,omitempty"`
	WebhookConfigs  []map[string]any `yaml:"webhook_configs,omitempty"`
	OpsgenieConfigs []map[string]any `yaml:"opsgenie_configs,omitempty"`
}

type Route struct {
	Receiver       string   `yaml:"receiver,omitempty"`
	GroupBy        []string `yaml:"group_by,omitempty"`
	GroupWait      string   `yaml:"group_wait,omitempty"`
	GroupInterval  string   `yaml:"group_interval,omitempty"`
	RepeatInterval string   `yaml:"repeat_interval,omitempty"`
	Matchers       []string `yaml:"matchers,omitempty"`
	Continue       bool     `yaml:"continue,omitempty"`
	TimeIntervals  []string `yaml:"time_intervals,omitempty"`
	Routes         []Route  `yaml:"routes,omitempty"`
}

type InhibitRule struct {
	SourceMatchers map[string]string `yaml:"source_matchers,omitempty"`
	TargetMatchers map[string]string `yaml:"target_matchers,omitempty"`
	Equal          []string          `yaml:"equal,omitempty"`
}

type TimeIntervalSet struct {
	Name          string         `yaml:"name"`
	TimeIntervals []TimeInterval `yaml:"time_intervals"`
}

type TimeInterval struct {
	// New fields aligned with Alertmanager's interval schema
	Weekdays    []string `yaml:"weekdays,omitempty"`
	DaysOfMonth []string `yaml:"days_of_month,omitempty"`
	Months      []string `yaml:"months,omitempty"`
	Years       []string `yaml:"years,omitempty"`
	Location    string   `yaml:"location,omitempty"` // timezone

	Times []struct {
		Start string `yaml:"start_time"`
		End   string `yaml:"end_time"`
	} `yaml:"times,omitempty"`
}
