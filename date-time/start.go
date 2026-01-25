package datetime

import "time"

func StartOfYear(t time.Time) time.Time {
	year := t.Year()
	return time.Date(year, time.January, 1, 0, 0, 0, 0, t.Location())
}

func StartOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
