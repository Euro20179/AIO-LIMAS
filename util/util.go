package util

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
)

func IsNumeric(text []byte) bool {
	matched, _ := regexp.Match("^[0-9]+$", text)
	return matched
}

func WError(w http.ResponseWriter, status int, format string, args ...any) {
	w.WriteHeader(status)
	fmt.Fprintf(w, format, args...)

	//also write to stderr
	fmt.Fprintf(os.Stderr, format, args...)
}

