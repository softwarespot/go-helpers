package helpers

import "os"

// IsPiped determines whether the application is piped to another application or not
func IsPiped() bool {
	// Taken from URL: https://rosettacode.org/wiki/Check_output_device_is_a_terminal#Go
	if info, _ := os.Stdout.Stat(); (info.Mode() & os.ModeCharDevice) != 0 {
		return false
	}
	return true
}
