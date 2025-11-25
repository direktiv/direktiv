package sched

import (
	"testing"
	"time"
)

func mustTime(y int, m time.Month, d, hh, mm, ss int) time.Time {
	return time.Date(y, m, d, hh, mm, ss, 0, time.Local)
}

func TestCalculateCronExpr(t *testing.T) {
	type tc struct {
		name     string
		cronExpr string
		start    time.Time
		want     time.Time
		wantErr  bool
	}

	tests := []tc{
		{
			name:     "6-field: every 2 seconds, simple increment",
			cronExpr: "*/2 * * * * *", // every 2 seconds
			start:    mustTime(2025, time.September, 17, 10, 0, 0),
			want:     mustTime(2025, time.September, 17, 10, 0, 2),
		},
		{
			name:     "6-field: every 30 seconds crosses minute boundary",
			cronExpr: "*/30 * * * * *", // 0,30
			start:    mustTime(2025, time.September, 17, 10, 0, 30),
			// schedule.Next is strictly after start, so next is 10:01:00
			want: mustTime(2025, time.September, 17, 10, 1, 0),
		},
		{
			name:     "6-field: spaces tolerated by strings.Fields",
			cronExpr: "*/10   *   *  *   *   *", // multiple spaces
			start:    mustTime(2025, time.September, 17, 10, 0, 9),
			want:     mustTime(2025, time.September, 17, 10, 0, 10),
		},
		{
			name:     "5-field: every 2 minutes (no seconds field)",
			cronExpr: "*/2 * * * *", // every 2 minutes
			start:    mustTime(2025, time.September, 17, 10, 0, 0),
			want:     mustTime(2025, time.September, 17, 10, 2, 0),
		},
		{
			name:     "5-field: next is strictly after start",
			cronExpr: "0 12 * * *", // 12:00 daily
			start:    mustTime(2025, time.September, 17, 12, 0, 0),
			// Next must be next day 12:00, not the same instant
			want: mustTime(2025, time.September, 18, 12, 0, 0),
		},
		{
			name:     "5-field: weekday filter, same day when before trigger",
			cronExpr: "0 12 * * MON-FRI",
			start:    mustTime(2025, time.September, 17, 11, 59, 59), // Wed
			want:     mustTime(2025, time.September, 17, 12, 0, 0),
		},
		{
			name:     "5-field: every 15 minutes, round up at 59s",
			cronExpr: "*/15 * * * *",
			start:    mustTime(2025, time.September, 17, 10, 14, 59),
			want:     mustTime(2025, time.September, 17, 10, 15, 0),
		},
		{
			name:     "6-field: invalid seconds value (61) -> error",
			cronExpr: "61 * * * * *",
			start:    mustTime(2025, time.September, 17, 10, 0, 0),
			wantErr:  true,
		},
		{
			name:     "parse error: nonsense spec",
			cronExpr: "not a cron",
			start:    mustTime(2025, time.September, 17, 10, 0, 0),
			wantErr:  true,
		},
		{
			name:     "6-field: exact hit must advance (strictly after start)",
			cronExpr: "0/20 * * * * *", // 0,20,40
			start:    mustTime(2025, time.September, 17, 10, 0, 40),
			want:     mustTime(2025, time.September, 17, 10, 1, 0),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateCronExpr(tt.cronExpr, tt.start)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (got=%v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("Next(%q, %s) = %s, want %s",
					tt.cronExpr, tt.start.Format(time.RFC3339), got.Format(time.RFC3339), tt.want.Format(time.RFC3339))
			}
		})
	}
}
