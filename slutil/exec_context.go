package slutil

import (
	"runtime"
)

type ExecContext struct {
	File     string `json:"-"`
	Line     int    `json:"-"`
	Function string `json:"-"`
}

func GetExecContext() ExecContext {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return ExecContext{
		File:     frame.File,
		Line:     frame.Line,
		Function: frame.Function,
	}
}
