package helpers

import "os"

// IsPiped determines whether the application is piped to another application or not
func IsPiped() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		// Ignore the error
		return false
	}

	// Taken from URL: https://rosettacode.org/wiki/Check_output_device_is_a_terminal#Go
	return (info.Mode() & os.ModeCharDevice) == 0
}
