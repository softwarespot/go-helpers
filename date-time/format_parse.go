package datetime

import "time"

const (
	layoutDateTime = "2006-01-02 15:04:05"
	zeroDateTime   = "0000-00-00 00:00:00"
)

// FormatAsDateTime formats a given time.Time value into a string
// representation in the format "YYYY-MM-DD HH:MM:SS" based on the
// current local time zone
func FormatAsDateTime(t time.Time) string {
	if t.IsZero() {
		return zeroDateTime
	}
	return t.Format(layoutDateTime)
}

// ParseAsDateTime parses a string representation of date and time
// in the format "YYYY-MM-DD HH:MM:SS" into a time.Time value
// based on the local time zone
func ParseAsDateTime(tt string) time.Time {
	t, err := time.ParseInLocation(layoutDateTime, tt, time.Local)
	if err != nil {
		return time.Time{}
	}
	return t
}
