package utils

import "strings"

func OsaEscape(input string) string {
	input = strings.ReplaceAll(input, "\"", "")
	return input
}
