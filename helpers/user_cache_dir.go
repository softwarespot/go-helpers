package helpers

import "os"

// UserCacheDir returns the user's cache directory as a string.
//
// This function is a wrapper around os.UserCacheDir() and is designed
// to simplify the retrieval of the cache directory without requiring
// the caller to handle the error explicitly. If an error occurs,
// it returns an empty string, which can be handled accordingly by
// the caller. This approach avoids panicking on error, allowing
// for more graceful error management in applications
func UserCacheDir() string {
	// There is no need to panic on error, as an empty string can be handled
	// accordingly
	dir, _ := os.UserCacheDir()
	return dir
}
