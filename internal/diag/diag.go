package diag

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
}

func Error(code, msg, file string) Diagnostic {
	return Diagnostic{
		Level:   LevelError,
		Code:    code,
		Message: msg,
		File:    file,
	}
}

func Warn(code, msg, file string) Diagnostic {
	return Diagnostic{
		Level:   LevelWarn,
		Code:    code,
		Message: msg,
		File:    file,
	}
}

func Info(code, msg, file string) Diagnostic {
	return Diagnostic{
		Level:   LevelInfo,
		Code:    code,
		Message: msg,
		File:    file,
	}
}
