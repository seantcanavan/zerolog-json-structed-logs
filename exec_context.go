package sl

import (
	"runtime"
)

type execContext struct {
	File     string `json:"-"`
	Line     int    `json:"-"`
	Function string `json:"-"`
}

func getExecContext() execContext {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return execContext{
		File:     frame.File,
		Line:     frame.Line,
		Function: frame.Function,
	}
}
