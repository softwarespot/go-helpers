package datetime

import (
	"iter"
	"time"
)

type Hour struct {
	Start time.Time
	End   time.Time
}

func Hours(hours int, end time.Time) iter.Seq[Hour] {
	return func(yield func(Hour) bool) {
		d := time.Duration(hours) * time.Hour
		from := end.Add(-d).Truncate(time.Hour)
		for !from.After(end) {
			to := from.Add(time.Hour - time.Second)
			if to.After(end) {
				to = end
			}

			if !yield(Hour{
				Start: from,
				End:   to,
			}) {
				return
			}
			from = from.Add(time.Hour)
		}
	}
}
