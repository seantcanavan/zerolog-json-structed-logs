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
	pc, fileName, lineNumber, ok := runtime.Caller(caller)
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
	noduleName := strings.Join(parts[:len(parts)-1], "/")

	return ExecContext{
		File:     fileName,
		Function: functionName,
		Line:     lineNumber,
		Module:   noduleName,
		Package:  packageName,
	}
}
