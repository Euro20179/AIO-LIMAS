package util

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"
)

type CancelSignal chan bool

func SetTimeout(duration time.Duration, cb func(...any) any, args ...any) CancelSignal {
	cancel := make(chan bool)
	signalSent := false
	go func() {
		go func() {
			time.Sleep(duration)
			println("DONE SLEEP")
			if !signalSent {
				println("RUN")
				cancel <- false
			} else {
				println("NOTHING")
			}
		}()

		if <- cancel {
			signalSent = true
		} else {
			signalSent = true
			cb(args...)
		}
		close(cancel)
	}()

	return cancel
}

func IsNumeric(text []byte) bool {
	matched, _ := regexp.Match("^[0-9]+$", text)
	return matched
}

func WError(w http.ResponseWriter, status int, format string, args ...any) {
	w.WriteHeader(status)
	fmt.Fprintf(w, format, args...)

	// also write to stderr
	fmt.Fprintf(os.Stderr, format, args...)
}

func Assert(condition bool, failFmt string, args ...string) {
	if !condition {
		panic(fmt.Sprintf(failFmt, args))
	}
}
