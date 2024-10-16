package helpers

// ErrorString returns the string representation of an error.
// If the provided error is nil, it returns an empty string
func ErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
