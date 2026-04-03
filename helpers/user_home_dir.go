package helpers

import "os"

// UserHomeDir returns the user's home directory as a string.
//
// This function is a wrapper around os.UserHomeDir() and is designed
// to simplify the retrieval of the home directory without requiring
// the caller to handle the error explicitly. If an error occurs,
// it returns an empty string, which can be handled accordingly by
// the caller. This approach avoids panicking on error, allowing
// for more graceful error management in applications
func UserHomeDir() string {
	// There is no need to panic on error, as an empty string can be handled
	// accordingly
	dir, _ := os.UserHomeDir()
	return dir
}
