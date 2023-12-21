package slutil

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

func FromCtxSafe[T any](ctx context.Context, key interface{}) T {
	val, ok := ctx.Value(key).(T)
	if !ok {
		// Return zero value of T if the type does not match
		return *new(T)
	}
	return val
}

func UneraseMapStringArray(input map[string]any) map[string][]string {
	res := make(map[string][]string)

	for currentKey, currentVal := range input {
		var strs []string
		for _, currentAny := range currentVal.([]any) {
			strs = append(strs, currentAny.(string))
		}

		res[currentKey] = strs
	}

	return res
}

func UneraseMapString(input map[string]any) map[string]string {
	res := make(map[string]string)

	for currentKey, currentVal := range input {
		res[currentKey] = currentVal.(string)
	}

	return res
}

func GetCallerDetails(caller int) (string, string) {
	pc, _, _, ok := runtime.Caller(caller)
	if !ok {
		return "unknown", "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown", "unknown"
	}

	fullFunctionName := fn.Name()
	components := strings.Split(fullFunctionName, ".")
	fmt.Println(fmt.Sprintf("[%+v]", components))
	lastComponent := components[len(components)-1]

	packageName := strings.Join(components[:len(components)-1], ".")
	functionName := lastComponent[strings.LastIndex(lastComponent, ".")+1:]

	return packageName, functionName
}

func PrettyInfoMsgF(calleePkg, calleeFn string, extra any) string {
	return fmt.Sprintf("%s with result %+v", PrettyInfoMsg(calleePkg, calleeFn), extra)
}

func PrettyInfoMsg(calleePkg, calleeFn string) string {
	callerPkg, callerFn := GetCallerDetails(2)
	return fmt.Sprintf("%s.%s successfully called %s.%s", callerPkg, callerFn, calleePkg, calleeFn)
}

func PrettyErrMsg(calleePkg, calleeFn string) string {
	callerPkg, callerFn := GetCallerDetails(2)
	return fmt.Sprintf("%s.%s unsuccessfully called %s.%s", callerPkg, callerFn, calleePkg, calleeFn)
}

func PrettyErrMsgF(calleePkg, calleeFn string, extra any) string {
	return fmt.Sprintf("%s with result %+v", PrettyErrMsg(calleePkg, calleeFn), extra)
}

func PrettyErrMsgInternal() string {
	callerPkg, callerFn := GetCallerDetails(2)
	return fmt.Sprintf("%s.%s encountered an error", callerPkg, callerFn)
}

func PrettyErrMsgInternalF(extra any) string {
	return fmt.Sprintf("%s with result %+v", PrettyErrMsgInternal(), extra)
}
