package utils

import "strings"

func OsaEscape(string string) string {
	string = strings.ReplaceAll(string, "\"", "")
	return string
}
