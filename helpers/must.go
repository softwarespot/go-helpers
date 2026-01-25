package helpers

// Must returns the result if there is no error; otherwise, it panics with the provided error.
// This function is useful for handling results from functions that return an error,
// allowing you to simplify error handling in cases where you expect no errors to occur.
//
// Example usage:
//
//	value, err := SomeFunction()
//	result := Must(value, err) // If err is nil, result will contain the value; if err is non-nil, it will panic.
//
// This function is typically used in scenarios where a failure is considered fatal,
// and you want to terminate execution immediately rather than handle the error gracefully
func Must[T any](res T, err error) T {
	if err != nil {
		panic(err)
	}
	return res
}
