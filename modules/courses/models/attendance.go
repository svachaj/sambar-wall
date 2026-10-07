package models

import "github.com/svachaj/sambar-wall/db/types"

const (
	ATTENDANCE_STATUS_PRESENT = "present"
	ATTENDANCE_STATUS_ABSENT  = "absent"
	ATTENDANCE_STATUS_UNSET   = "unset"
)

// AttendanceStatus maps a stored attendance record to one of the ATTENDANCE_STATUS_* values.
func AttendanceStatus(hasRecord bool, present bool) string {
	if !hasRecord {
		return ATTENDANCE_STATUS_UNSET
	}
	if present {
		return ATTENDANCE_STATUS_PRESENT
	}
	return ATTENDANCE_STATUS_ABSENT
}

// AttendanceStatusLabel returns the label used in the UI and exports.
func AttendanceStatusLabel(status string) string {
	switch status {
	case ATTENDANCE_STATUS_PRESENT:
		return "Byl"
	case ATTENDANCE_STATUS_ABSENT:
		return "Nebyl"
	default:
		return "—"
	}
}

type AttendancePageModel struct {
	// courses taking place on the weekday of LessonDate
	DayCourses []types.Course
	// all other active courses (for make-up lessons on a different day)
	OtherCourses     []types.Course
	SelectedCourseID int
	LessonDate       string
	LessonDateLabel  string
	PrevLessonDate   string
	NextLessonDate   string
	// the selected course does not take place on the weekday of LessonDate
	DayMismatch        bool
	SelectedCourseDays string
	Rows               []types.AttendanceSheetRow
	ExportDateFrom     string
	ExportDateTo       string
}
