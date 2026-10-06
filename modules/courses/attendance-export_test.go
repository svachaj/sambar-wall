package courses

import (
	"testing"
	"time"

	"github.com/svachaj/sambar-wall/db/types"
)

func TestBuildAttendanceWorkbook(t *testing.T) {
	timeFrom := time.Date(1900, 1, 1, 16, 0, 0, 0, time.UTC)
	timeTo := time.Date(1900, 1, 1, 17, 30, 0, 0, time.UTC)
	participants := []types.AttendanceSheetRow{
		{ApplicationFormID: 1, CourseID: 10, CourseName: "Lezení děti", CourseDays: "Pondělí", CourseDayCode: "1", CourseTimeFrom: timeFrom, CourseTimeTo: timeTo, FirstName: "Jan", LastName: "Novák"},
		{ApplicationFormID: 2, CourseID: 10, CourseName: "Lezení děti", CourseDays: "Pondělí", CourseDayCode: "1", CourseTimeFrom: timeFrom, CourseTimeTo: timeTo, FirstName: "Eva", LastName: "Malá"},
		{ApplicationFormID: 3, CourseID: 11, CourseName: "Lezení děti", CourseDays: "Pondělí", CourseDayCode: "1", CourseTimeFrom: timeFrom, CourseTimeTo: timeTo, FirstName: "Petr", LastName: "Velký"},
	}
	day1 := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	records := []types.AttendanceExportRecord{
		{ApplicationFormID: 1, LessonDate: day1, Present: true},
		{ApplicationFormID: 1, LessonDate: day2, Present: false},
		{ApplicationFormID: 2, LessonDate: day2, Present: true},
	}

	f, err := buildAttendanceWorkbook(participants, records)
	if err != nil {
		t.Fatal(err)
	}

	sheets := f.GetSheetList()
	want := []string{"Po 16.00 Lezení děti", "Po 16.00 Lezení děti (2)", attendanceAllRecordsSheet}
	if len(sheets) != len(want) {
		t.Fatalf("sheets = %v, want %v", sheets, want)
	}
	for i := range want {
		if sheets[i] != want[i] {
			t.Fatalf("sheets = %v, want %v", sheets, want)
		}
	}

	rows, _ := f.GetRows(sheets[0])
	// header: Účastník, Rodič, Telefon, 5.10.2026, 12.10.2026, Účast, %
	if got := rows[0][3]; got != "5.10.2026" {
		t.Errorf("first date header = %q", got)
	}
	// Novák: Byl, Nebyl -> 1 present
	if rows[1][3] != "Byl" || rows[1][4] != "Nebyl" || rows[1][5] != "1" {
		t.Errorf("unexpected row for Novák: %v", rows[1])
	}
	// Malá: not filled in on the first date
	if rows[2][3] != "—" || rows[2][4] != "Byl" {
		t.Errorf("unexpected row for Malá: %v", rows[2])
	}

	allRows, _ := f.GetRows(attendanceAllRecordsSheet)
	if len(allRows) != 1+len(records) {
		t.Errorf("flat sheet has %d rows, want %d", len(allRows), 1+len(records))
	}
}

func TestUniqueSheetName(t *testing.T) {
	used := map[string]bool{}
	long := "Po 16:00 Velmi dlouhý název kurzu pro děti [pokročilí]"
	first := uniqueSheetName(long, used)
	second := uniqueSheetName(long, used)

	if len([]rune(first)) > 31 || len([]rune(second)) > 31 {
		t.Errorf("sheet names too long: %q, %q", first, second)
	}
	if first == second {
		t.Errorf("expected unique names, got %q twice", first)
	}
	for _, r := range first {
		if r == '[' || r == ']' || r == ':' {
			t.Errorf("invalid character in %q", first)
		}
	}
}
