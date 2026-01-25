package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Ensure interface compatibility
var _ Logger = &StdoutLogger{}

// StdoutLogger is a logger that writes log entries to standard output in JSON format
type StdoutLogger struct {
	enc *json.Encoder
	mu  sync.Mutex
}

// NewStdoutLogger creates a new instance of StdoutLogger
func NewStdoutLogger() *StdoutLogger {
	return &StdoutLogger{
		enc: json.NewEncoder(os.Stdout),
	}
}

// Fatal logs a critical error message and exits the application with the specified exit code
func (sl *StdoutLogger) Fatal(err error, code int, args ...any) {
	sl.LogError(err, LevelCritical, args...)
	os.Exit(code)
}

// LogError logs an error message with the specified log level
func (sl *StdoutLogger) LogError(err error, level Level, args ...any) {
	sl.Log(err.Error(), level, args...)
}

// Log writes a log entry with a message, log level, and optional additional arguments, which must be divisible by 2
func (sl *StdoutLogger) Log(msg string, level Level, args ...any) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if err := sl.enc.Encode(createEntry(msg, level, args...)); err != nil {
		fmt.Fprintf(os.Stderr, "skipped logging entry due to error: %+v. Message: %q, Level: %v\n", err, msg, level)
	}
}
