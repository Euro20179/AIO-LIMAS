package logging

import (
	"fmt"
	"os"
	"runtime"
	"slices"
)


type LLEVEL int

const (
	L_ERROR LLEVEL = iota
	L_WARN LLEVEL = iota
	L_INFO LLEVEL = iota
)

var OUTFILE = os.Stderr
var LOGLEVEL = L_ERROR

var wd = os.Getenv("PWD")

func log(level LLEVEL, text string, values ...any) {
	//going up 1 will only go up to whatever log wrapper function called this function
	_, file, line, _ := runtime.Caller(2)
	file = file[len(wd) + 1:]
	if LOGLEVEL >= level {
		fmt.Fprintf(OUTFILE, "%s:%d " + text + "\n", slices.Concat([]any{file, line}, values)...)
	}
}

func ELog(error error) {
	log(L_ERROR, "%s", error.Error())
}

func Error(text string, values ...any) {
	log(L_ERROR, text, values...)
}

func Warn(text string, values ...any) {
	log(L_WARN, text, values...)
}

func Info(text string, values ...any) {
	log(L_INFO, text, values...)
}
