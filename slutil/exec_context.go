package slutil

import (
	"runtime"
	"strings"
)

type ExecContext struct {
	File     string `json:"-"`
	Function string `json:"-"`
	Line     int    `json:"-"`
	Module   string `json:"-"`
	Package  string `json:"-"`
}

func GetExecContext(caller int) ExecContext {
	pc, file, line, ok := runtime.Caller(caller)
	if !ok {
		return ExecContext{}
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ExecContext{}
	}

	fullFunctionName := fn.Name()

	// Split the string by '/'
	parts := strings.Split(fullFunctionName, "/")

	// The last part contains the package and function name
	lastPart := parts[len(parts)-1]

	// Split the last part by '.'
	pkgFunc := strings.Split(lastPart, ".")
	packageName := pkgFunc[0]
	functionName := pkgFunc[1]

	// Rejoin the remaining parts to form the module name
	moduleName := strings.Join(parts[:len(parts)-1], "/")

	return ExecContext{
		File:     file,
		Function: functionName,
		Line:     line,
		Module:   moduleName,
		Package:  packageName,
	}
}
