package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestParseCourseDays(t *testing.T) {
	tests := []struct {
		code string
		name string
		want []time.Weekday
	}{
		{"1", "Pondělí", []time.Weekday{time.Monday}},
		{"", "úterý", []time.Weekday{time.Tuesday}},
		{"", "Po + St", []time.Weekday{time.Monday, time.Wednesday}},
		{"", "Čt, Pá", []time.Weekday{time.Thursday, time.Friday}},
		{"7", "", []time.Weekday{time.Sunday}},
		{"CD4", "", []time.Weekday{time.Thursday}},
		{"CD1", "Pondělí lano ", []time.Weekday{time.Monday}},
		{"1jednodenní", "2.4.", []time.Weekday{}},
		{"105", "20.7. - 24.7. ", []time.Weekday{}},
		{"3", "unknown", []time.Weekday{time.Wednesday}},
		{"x", "unknown", []time.Weekday{}},
	}

	for _, tt := range tests {
		got := ParseCourseDays(tt.code, tt.name)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("ParseCourseDays(%q, %q) = %v, want %v", tt.code, tt.name, got, tt.want)
		}
	}
}

func TestCourseDayMatchesDate(t *testing.T) {
	monday := time.Date(2026, time.October, 5, 0, 0, 0, 0, time.Local)
	if !CourseDayMatchesDate("1", "Pondělí", monday) {
		t.Error("expected Monday course to match Monday date")
	}
	if CourseDayMatchesDate("2", "Úterý", monday) {
		t.Error("expected Tuesday course not to match Monday date")
	}
}

func TestCourseDayMatchesDateRange(t *testing.T) {
	date := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.Local) }
	tests := []struct {
		name string
		date time.Time
		want bool
	}{
		{"20.7. - 24.7. ", date(2026, time.July, 20), true},
		{"20.7. - 24.7. ", date(2026, time.July, 24), true},
		{"20.7. - 24.7. ", date(2026, time.July, 25), false},
		{"21.11.-22.11", date(2026, time.November, 22), true},
		{"2.4.", date(2027, time.April, 2), true},
		{"2.4.", date(2027, time.April, 3), false},
		{"28.12. - 2.1.", date(2027, time.January, 1), true},
		{"28.12. - 2.1.", date(2026, time.December, 29), true},
	}
	for _, tt := range tests {
		if got := CourseDayMatchesDate("105", tt.name, tt.date); got != tt.want {
			t.Errorf("CourseDayMatchesDate(%q, %v) = %v, want %v", tt.name, tt.date, got, tt.want)
		}
	}
}

func TestSchoolYearStart(t *testing.T) {
	tests := []struct {
		date time.Time
		want time.Time
	}{
		{time.Date(2026, time.October, 6, 0, 0, 0, 0, time.UTC), time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2027, time.March, 15, 0, 0, 0, 0, time.UTC), time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		if got := SchoolYearStart(tt.date); !got.Equal(tt.want) {
			t.Errorf("SchoolYearStart(%v) = %v, want %v", tt.date, got, tt.want)
		}
	}
}

func TestFormatCzechDateWithWeekday(t *testing.T) {
	got := FormatCzechDateWithWeekday(time.Date(2026, time.October, 5, 0, 0, 0, 0, time.Local))
	if got != "pondělí 5. 10. 2026" {
		t.Errorf("unexpected format: %q", got)
	}
}
