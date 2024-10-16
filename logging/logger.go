package logging

// Logger is an interface, as it allows different logging adapters to be created e.g.
// a file logger, a logger which sends to Slack etc.
type Logger interface {
	// NOTE: This MUST exit the process
	Fatal(err error, code int, args ...any)
	LogError(err error, level Level, args ...any)
	Log(msg string, level Level, args ...any)
}
