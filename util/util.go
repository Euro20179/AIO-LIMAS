package util

import "regexp"

func IsNumeric(text []byte) bool {
	matched, _ := regexp.Match("^[0-9]+$", text)
	return matched
}
