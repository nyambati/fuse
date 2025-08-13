package types

// Level is the severity of a diagnostic.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Diagnostic represents a single validation message.
type Diagnostic struct {
	Level   Level  `json:"level"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	// Optional: Column, Hint, etc.
}
