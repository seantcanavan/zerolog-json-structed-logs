package sl

import "context"

func fromCtxSafe[T any](ctx context.Context, key interface{}) T {
	val, ok := ctx.Value(key).(T)
	if !ok {
		// Return zero value of T if the type does not match
		return *new(T)
	}
	return val
}

func uneraseMapStringArray(input map[string]any) map[string][]string {
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

func uneraseMapString(input map[string]any) map[string]string {
	res := make(map[string]string)

	for currentKey, currentVal := range input {
		res[currentKey] = currentVal.(string)
	}

	return res
}
