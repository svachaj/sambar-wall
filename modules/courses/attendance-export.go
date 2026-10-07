package courses

import (
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/modules/courses/models"
	"github.com/svachaj/sambar-wall/utils"
	"github.com/xuri/excelize/v2"
)

const attendanceAllRecordsSheet = "Všechny záznamy"

type attendanceExportCourse struct {
	label        string
	sheetName    string
	participants []types.AttendanceSheetRow
}

// buildAttendanceWorkbook creates one matrix sheet per course (children x lesson dates)
// and a flat sheet with every recorded attendance.
func buildAttendanceWorkbook(participants []types.AttendanceSheetRow, records []types.AttendanceExportRecord) (*excelize.File, error) {
	// attendance status by application form and lesson date
	statusByApp := map[int]map[string]string{}
	for _, record := range records {
		if statusByApp[record.ApplicationFormID] == nil {
			statusByApp[record.ApplicationFormID] = map[string]string{}
		}
		statusByApp[record.ApplicationFormID][record.LessonDate.Format(utils.ISODateLayout)] = models.AttendanceStatus(true, record.Present)
	}

	// group participants by course, keeping the day/time order from the query
	courses := []*attendanceExportCourse{}
	courseByID := map[int]*attendanceExportCourse{}
	usedSheetNames := map[string]bool{strings.ToLower(attendanceAllRecordsSheet): true}
	for _, p := range participants {
		course, ok := courseByID[p.CourseID]
		if !ok {
			label := attendanceCourseLabel(p)
			course = &attendanceExportCourse{label: label, sheetName: uniqueSheetName(label, usedSheetNames)}
			courseByID[p.CourseID] = course
			courses = append(courses, course)
		}
		course.participants = append(course.participants, p)
	}

	f := excelize.NewFile()
	headerStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, err
	}
	percentStyle, err := f.NewStyle(&excelize.Style{NumFmt: 9})
	if err != nil {
		return nil, err
	}

	for _, course := range courses {
		if _, err := f.NewSheet(course.sheetName); err != nil {
			return nil, err
		}
		if err := writeAttendanceCourseSheet(f, course, statusByApp, headerStyle, percentStyle); err != nil {
			return nil, err
		}
	}

	if _, err := f.NewSheet(attendanceAllRecordsSheet); err != nil {
		return nil, err
	}
	if err := writeAttendanceAllRecordsSheet(f, courses, statusByApp, headerStyle); err != nil {
		return nil, err
	}

	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}
	f.SetActiveSheet(0)

	return f, nil
}

func writeAttendanceCourseSheet(f *excelize.File, course *attendanceExportCourse, statusByApp map[int]map[string]string, headerStyle, percentStyle int) error {
	sheet := course.sheetName

	// lesson dates with at least one record for this course
	dateSet := map[string]bool{}
	for _, p := range course.participants {
		for date := range statusByApp[p.ApplicationFormID] {
			dateSet[date] = true
		}
	}
	dates := make([]string, 0, len(dateSet))
	for date := range dateSet {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	header := []interface{}{"Účastník", "Rodič", "Telefon"}
	for _, date := range dates {
		header = append(header, formatExportDate(date))
	}
	header = append(header, "Účast", "%")
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		return err
	}
	lastCol, _ := excelize.CoordinatesToCellName(len(header), 1)
	if err := f.SetCellStyle(sheet, "A1", lastCol, headerStyle); err != nil {
		return err
	}

	for i, p := range course.participants {
		statuses := statusByApp[p.ApplicationFormID]
		row := []interface{}{p.LastName + " " + p.FirstName, utils.StringFromStringPointer(p.ParentName), utils.StringFromStringPointer(p.ParentPhone)}

		presentCount, recordedCount := 0, 0
		for _, date := range dates {
			status, ok := statuses[date]
			if !ok {
				status = models.ATTENDANCE_STATUS_UNSET
			} else {
				recordedCount++
				if status == models.ATTENDANCE_STATUS_PRESENT {
					presentCount++
				}
			}
			row = append(row, models.AttendanceStatusLabel(status))
		}

		row = append(row, presentCount)
		if recordedCount > 0 {
			row = append(row, float64(presentCount)/float64(recordedCount))
		} else {
			row = append(row, "")
		}

		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			return err
		}
	}

	if len(course.participants) > 0 {
		percentCol, _ := excelize.ColumnNumberToName(len(header))
		if err := f.SetCellStyle(sheet, percentCol+"2", percentCol+itoa(len(course.participants)+1), percentStyle); err != nil {
			return err
		}
	}

	if err := f.SetColWidth(sheet, "A", "C", 22); err != nil {
		return err
	}
	return f.SetPanes(sheet, &excelize.Panes{Freeze: true, XSplit: 1, YSplit: 1, TopLeftCell: "B2", ActivePane: "bottomRight"})
}

func writeAttendanceAllRecordsSheet(f *excelize.File, courses []*attendanceExportCourse, statusByApp map[int]map[string]string, headerStyle int) error {
	sheet := attendanceAllRecordsSheet

	header := []interface{}{"Kurz", "Den", "Čas", "Datum lekce", "Účastník", "Rodič", "Stav"}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "G1", headerStyle); err != nil {
		return err
	}

	rowIdx := 2
	for _, course := range courses {
		for _, p := range course.participants {
			statuses := statusByApp[p.ApplicationFormID]
			dates := make([]string, 0, len(statuses))
			for date := range statuses {
				dates = append(dates, date)
			}
			sort.Strings(dates)

			for _, date := range dates {
				row := []interface{}{
					p.CourseName,
					p.CourseDays,
					p.CourseTimeFrom.Format("15:04") + "-" + p.CourseTimeTo.Format("15:04"),
					formatExportDate(date),
					p.LastName + " " + p.FirstName,
					utils.StringFromStringPointer(p.ParentName),
					models.AttendanceStatusLabel(statuses[date]),
				}
				cell, _ := excelize.CoordinatesToCellName(1, rowIdx)
				if err := f.SetSheetRow(sheet, cell, &row); err != nil {
					return err
				}
				rowIdx++
			}
		}
	}

	if err := f.SetColWidth(sheet, "A", "G", 20); err != nil {
		return err
	}
	if err := f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
		return err
	}
	return f.AutoFilter(sheet, "A1:G"+itoa(max(rowIdx-1, 1)), nil)
}

// attendanceCourseLabel returns e.g. "Po 16.00 Lezení děti" (":" is not allowed in sheet names).
func attendanceCourseLabel(p types.AttendanceSheetRow) string {
	day := p.CourseDays
	if days := utils.ParseCourseDays(p.CourseDayCode, p.CourseDays); len(days) > 0 {
		shortNames := make([]string, 0, len(days))
		for _, d := range days {
			shortNames = append(shortNames, utils.CzechWeekdayShortName(d))
		}
		day = strings.Join(shortNames, "+")
	}
	return strings.Join(strings.Fields(day+" "+p.CourseTimeFrom.Format("15.04")+" "+p.CourseName), " ")
}

// uniqueSheetName makes a valid Excel sheet name (max 31 chars, no []:*?/\) that is not used yet.
func uniqueSheetName(name string, used map[string]bool) string {
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`[]:*?/\`, r) {
			return '-'
		}
		return r
	}, name)
	name = strings.Trim(name, "' ")
	if name == "" {
		name = "Kurz"
	}

	candidate := truncateRunes(name, 31)
	for i := 2; used[strings.ToLower(candidate)]; i++ {
		suffix := " (" + itoa(i) + ")"
		candidate = truncateRunes(name, 31-utf8.RuneCountInString(suffix)) + suffix
	}
	used[strings.ToLower(candidate)] = true
	return candidate
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

func formatExportDate(isoDate string) string {
	date, err := time.Parse(utils.ISODateLayout, isoDate)
	if err != nil {
		return isoDate
	}
	return date.Format("2.1.2006")
}

func itoa(i int) string {
	return utils.StringFromInt(i)
}
