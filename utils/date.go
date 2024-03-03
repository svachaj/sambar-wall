package utils

import "strings"

func NormalizeDate(value string) string {
	return strings.ReplaceAll(value, " ", "")
}
