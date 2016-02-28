package main

import (
	"testing"
	"time"
)

// fakeDate create date for testing.
func fakeDate(y, m, d int) time.Time {
	if y == 0 && m == 0 && d == 0 {
		return midnight(time.Time{})
	}
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// formatDate returns shorter date representation.
func formatDate(date time.Time) string {
	return date.Format(iso8601DateFormat)
}

func TestMidnight(t *testing.T) {
	in := time.Date(2016, 1, 23, 1, 2, 3, 4, time.UTC)
	want := time.Date(2016, 1, 23, 0, 0, 0, 0, time.UTC)
	if got := midnight(in); got != want {
		t.Errorf("midnight(%v) = %v; want %v", in, got, want)
	}
}

func TestParseDate(t *testing.T) {
	// invalid
	in := "abc"
	want := fakeDate(0, 0, 0)
	if got, err := parseDate(in); got != want || err == nil {
		t.Errorf("parseDate(%q) = %v, %v; want %v, error",
			in, formatDate(got), err, formatDate(want))
	}

	// valid
	in = "2016-01-23"
	want = fakeDate(2016, 1, 23)
	if got, err := parseDate(in); got != want || err != nil {
		t.Errorf("parseDate(%q) = %v, %v; want %v, nil",
			in, formatDate(got), err, formatDate(want))
	}
}
