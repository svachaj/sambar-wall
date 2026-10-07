package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const ISODateLayout = "2006-01-02"

func NormalizeDate(value string) string {
	value = strings.ReplaceAll(value, " ", "")
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return ""
	}

	return parts[2] + "-" + parts[1] + "-" + parts[0]
}

// ParseISODate parses a date in the YYYY-MM-DD format (as sent by <input type="date">).
func ParseISODate(value string) (time.Time, error) {
	return time.ParseInLocation(ISODateLayout, value, time.Local)
}

var czechWeekdayNames = map[time.Weekday]string{
	time.Monday:    "pondělí",
	time.Tuesday:   "úterý",
	time.Wednesday: "středa",
	time.Thursday:  "čtvrtek",
	time.Friday:    "pátek",
	time.Saturday:  "sobota",
	time.Sunday:    "neděle",
}

// CzechWeekdayName returns the lower-case Czech name of the weekday, e.g. "pondělí".
func CzechWeekdayName(day time.Weekday) string {
	return czechWeekdayNames[day]
}

var czechWeekdayShortNames = map[time.Weekday]string{
	time.Monday:    "Po",
	time.Tuesday:   "Út",
	time.Wednesday: "St",
	time.Thursday:  "Čt",
	time.Friday:    "Pá",
	time.Saturday:  "So",
	time.Sunday:    "Ne",
}

// CzechWeekdayShortName returns the two-letter Czech abbreviation of the weekday, e.g. "Po".
func CzechWeekdayShortName(day time.Weekday) string {
	return czechWeekdayShortNames[day]
}

// FormatCzechDateWithWeekday formats a date as e.g. "pondělí 5. 10. 2026".
func FormatCzechDateWithWeekday(date time.Time) string {
	return CzechWeekdayName(date.Weekday()) + " " + date.Format("2. 1. 2006")
}

// czechDayTokens maps normalized (lower-case, no diacritics) day names and abbreviations to weekdays.
var czechDayTokens = map[string]time.Weekday{
	"po": time.Monday, "pondeli": time.Monday,
	"ut": time.Tuesday, "utery": time.Tuesday,
	"st": time.Wednesday, "streda": time.Wednesday,
	"ct": time.Thursday, "ctvrtek": time.Thursday,
	"pa": time.Friday, "patek": time.Friday,
	"so": time.Saturday, "sobota": time.Saturday,
	"ne": time.Sunday, "nedele": time.Sunday,
}

var czechDiacriticsReplacer = strings.NewReplacer(
	"á", "a", "č", "c", "ď", "d", "é", "e", "ě", "e", "í", "i", "ň", "n",
	"ó", "o", "ř", "r", "š", "s", "ť", "t", "ú", "u", "ů", "u", "ý", "y", "ž", "z",
)

var courseDayCodeRegex = regexp.MustCompile(`^(?i)(?:CD)?([1-7])$`)

// ParseCourseDays resolves the weekdays a course takes place on from its day name
// (e.g. "Pondělí", "Po + St") and, as a fallback, from its day code ("CD1"/"1" = Monday ... "CD7"/"7" = Sunday).
func ParseCourseDays(code string, name string) []time.Weekday {
	days := []time.Weekday{}
	seen := map[time.Weekday]bool{}

	normalized := czechDiacriticsReplacer.Replace(strings.ToLower(name))
	tokens := strings.FieldsFunc(normalized, func(r rune) bool { return !unicode.IsLetter(r) })
	for _, token := range tokens {
		if day, ok := czechDayTokens[token]; ok && !seen[day] {
			seen[day] = true
			days = append(days, day)
		}
	}

	if len(days) == 0 {
		if match := courseDayCodeRegex.FindStringSubmatch(strings.TrimSpace(code)); match != nil {
			n, _ := strconv.Atoi(match[1])
			days = append(days, time.Weekday(n%7))
		}
	}

	return days
}

var dayMonthRegex = regexp.MustCompile(`(\d{1,2})\.\s*(\d{1,2})\.?`)

// courseDateRangeContains reports whether the date falls into a course "day" given as a date or a date range
// without a year (e.g. "2.4.", "20.7. - 24.7.", "28.12. - 2.1."), as used for camps and one-day events.
func courseDateRangeContains(name string, date time.Time) bool {
	matches := dayMonthRegex.FindAllStringSubmatch(name, -1)
	if len(matches) == 0 || len(matches) > 2 {
		return false
	}

	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	toDate := func(m []string, year int) time.Time {
		d, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		return time.Date(year, time.Month(mo), d, 0, 0, 0, 0, time.UTC)
	}

	// try the range starting in the date's year and in the previous year (for ranges over New Year)
	for _, year := range []int{day.Year(), day.Year() - 1} {
		from := toDate(matches[0], year)
		to := from
		if len(matches) == 2 {
			to = toDate(matches[1], year)
			if to.Before(from) {
				to = to.AddDate(1, 0, 0)
			}
		}
		if !day.Before(from) && !day.After(to) {
			return true
		}
	}
	return false
}

// CourseDayMatchesDate reports whether a course with the given day code/name takes place on the date:
// either on its weekday (regular courses) or within its date range (camps, one-day events).
func CourseDayMatchesDate(code string, name string, date time.Time) bool {
	days := ParseCourseDays(code, name)
	for _, day := range days {
		if day == date.Weekday() {
			return true
		}
	}
	if len(days) == 0 {
		return courseDateRangeContains(name, date)
	}
	return false
}

// SchoolYearStart returns 1 September of the school year the given date belongs to.
func SchoolYearStart(date time.Time) time.Time {
	year := date.Year()
	if date.Month() < time.September {
		year--
	}
	return time.Date(year, time.September, 1, 0, 0, 0, 0, date.Location())
}
