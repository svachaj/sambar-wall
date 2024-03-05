package utils

import "strings"

func NormalizeDate(value string) string {
	value = strings.ReplaceAll(value, " ", "")
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return ""
	}

	return parts[2] + "-" + parts[1] + "-" + parts[0]
}
