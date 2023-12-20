package slutil

import "context"

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
