package errors

// Arg represents a key/value attribute attached to an error.
// This is the same as structured logging [slog.Attr] patterns (e.g. slog)
type Arg struct {
	Key   string
	Value any
}

// Any creates a generic [Arg]
func Any(key string, value any) Arg {
	return Arg{
		Key:   key,
		Value: value,
	}
}

// Bool creates an Arg for a [bool] value
func Bool(key string, value bool) Arg {
	return Any(key, value)
}

// Int creates an Arg for an [int] value
func Int(key string, value int) Arg {
	return Any(key, value)
}

// Float64 creates an Arg for a [float64] value
func Float64(key string, value float64) Arg {
	return Any(key, value)
}

// String creates an Arg for a [string] value
func String(key, value string) Arg {
	return Any(key, value)
}
