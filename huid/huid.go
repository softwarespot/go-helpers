package huid

import "time"

var nowFn = func() time.Time {
	return time.Now().UTC()
}

// New generates a unique identifier based on the current date and time.
// The format of git the identifier is "YYYYMMDD-HHMMSS".
// IMPORTANT: IDs generated within the same second will be identical, as it's expected to generated at
// the speed of human interaction
func New() string {
	return nowFn().Format("20060102-150405")
}

// SetNowFunc allows overriding the function which is used to get the current time.
// IMPORTANT: This function is not thread-safe and should only be used on initialization or in tests.
func SetNowFunc(fn func() time.Time) {
	nowFn = fn
}
