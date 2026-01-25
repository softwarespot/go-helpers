package datetime

import (
	"testing"
	"time"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_Hours(t *testing.T) {
	tests := []struct {
		name  string
		to    time.Time
		hours int
		want  []Hour
	}{
		{
			name:  "0 full hours",
			to:    ParseAsDateTime("2025-12-20 12:20:00"),
			hours: 0,
			want: []Hour{
				{
					Start: ParseAsDateTime("2025-12-20 12:00:00"),
					End:   ParseAsDateTime("2025-12-20 12:20:00"),
				},
			},
		},
		{
			name:  "1 full hour",
			to:    ParseAsDateTime("2025-12-20 12:20:00"),
			hours: 1,
			want: []Hour{
				{
					Start: ParseAsDateTime("2025-12-20 11:00:00"),
					End:   ParseAsDateTime("2025-12-20 11:59:59"),
				},
				{
					Start: ParseAsDateTime("2025-12-20 12:00:00"),
					End:   ParseAsDateTime("2025-12-20 12:20:00"),
				},
			},
		},
		{
			name:  "Start of hour",
			to:    ParseAsDateTime("2025-12-20 12:00:00"),
			hours: 0,
			want: []Hour{
				{
					Start: ParseAsDateTime("2025-12-20 12:00:00"),
					End:   ParseAsDateTime("2025-12-20 12:00:00"),
				},
			},
		},
		{
			name:  "End of hour",
			to:    ParseAsDateTime("2025-12-20 12:59:59"),
			hours: 0,
			want: []Hour{
				{
					Start: ParseAsDateTime("2025-12-20 12:00:00"),
					End:   ParseAsDateTime("2025-12-20 12:59:59"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []Hour
			for hour := range Hours(tt.hours, tt.to) {
				got = append(got, hour)
			}
			testhelpers.AssertEqual(t, got, tt.want)
		})
	}
}
