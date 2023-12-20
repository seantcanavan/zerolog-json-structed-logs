package sl

import (
	"context"
	"testing"
)

// TestFromCtxSafe contains individual test cases for different types
func TestFromCtxSafe(t *testing.T) {
	// Test for int
	t.Run("Int", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "intKey", 42)
		if got := fromCtxSafe[int](ctx, "intKey"); got != 42 {
			t.Errorf("fromCtxSafe[int] = %v, want %v", got, 42)
		}
	})

	// Test for string
	t.Run("String", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "stringKey", "hello")
		if got := fromCtxSafe[string](ctx, "stringKey"); got != "hello" {
			t.Errorf("fromCtxSafe[string] = %v, want %v", got, "hello")
		}
	})

	// Test for bool
	t.Run("Bool", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "boolKey", true)
		if got := fromCtxSafe[bool](ctx, "boolKey"); !got {
			t.Errorf("fromCtxSafe[bool] = %v, want %v", got, true)
		}
	})

	// Test for map[string]string
	t.Run("MapStringString", func(t *testing.T) {
		testMap := map[string]string{"key": "value"}
		ctx := context.WithValue(context.Background(), "mapKey", testMap)
		if got := fromCtxSafe[map[string]string](ctx, "mapKey"); got["key"] != "value" {
			t.Errorf("fromCtxSafe[map[string]string] = %v, want %v", got, testMap)
		}
	})

	// Test for map[string][]string
	t.Run("MapStringSliceString", func(t *testing.T) {
		testMap := map[string][]string{"key": {"value1", "value2"}}
		ctx := context.WithValue(context.Background(), "mapSliceKey", testMap)
		if got := fromCtxSafe[map[string][]string](ctx, "mapSliceKey"); len(got["key"]) != 2 || got["key"][0] != "value1" || got["key"][1] != "value2" {
			t.Errorf("fromCtxSafe[map[string][]string] = %v, want %v", got, testMap)
		}
	})

	// Test for byte (alias for uint8)
	t.Run("Byte", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "byteKey", byte('a'))
		if got := fromCtxSafe[byte](ctx, "byteKey"); got != byte('a') {
			t.Errorf("fromCtxSafe[byte] = %v, want %v", got, byte('a'))
		}
	})

	// Test for rune (alias for int32)
	t.Run("Rune", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "runeKey", rune('a'))
		if got := fromCtxSafe[rune](ctx, "runeKey"); got != rune('a') {
			t.Errorf("fromCtxSafe[rune] = %v, want %v", got, rune('a'))
		}
	})

	// Test for float32
	t.Run("Float32", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "float32Key", float32(3.14))
		if got := fromCtxSafe[float32](ctx, "float32Key"); got != float32(3.14) {
			t.Errorf("fromCtxSafe[float32] = %v, want %v", got, float32(3.14))
		}
	})

	// Test for float64
	t.Run("Float64", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "float64Key", 3.14159)
		if got := fromCtxSafe[float64](ctx, "float64Key"); got != 3.14159 {
			t.Errorf("fromCtxSafe[float64] = %v, want %v", got, 3.14159)
		}
	})

	// Test for uint
	t.Run("Uint", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "uintKey", uint(42))
		if got := fromCtxSafe[uint](ctx, "uintKey"); got != uint(42) {
			t.Errorf("fromCtxSafe[uint] = %v, want %v", got, uint(42))
		}
	})

	// Test for uint16
	t.Run("Uint16", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "uint16Key", uint16(42))
		if got := fromCtxSafe[uint16](ctx, "uint16Key"); got != uint16(42) {
			t.Errorf("fromCtxSafe[uint16] = %v, want %v", got, uint16(42))
		}
	})

	// Test for uint32
	t.Run("Uint32", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "uint32Key", uint32(42))
		if got := fromCtxSafe[uint32](ctx, "uint32Key"); got != uint32(42) {
			t.Errorf("fromCtxSafe[uint32] = %v, want %v", got, uint32(42))
		}
	})

	// Test for uint64
	t.Run("Uint64", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "uint64Key", uint64(42))
		if got := fromCtxSafe[uint64](ctx, "uint64Key"); got != uint64(42) {
			t.Errorf("fromCtxSafe[uint64] = %v, want %v", got, uint64(42))
		}
	})

	// Test for int8
	t.Run("Int8", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "int8Key", int8(42))
		if got := fromCtxSafe[int8](ctx, "int8Key"); got != int8(42) {
			t.Errorf("fromCtxSafe[int8] = %v, want %v", got, int8(42))
		}
	})

	// Test for int16
	t.Run("Int16", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "int16Key", int16(42))
		if got := fromCtxSafe[int16](ctx, "int16Key"); got != int16(42) {
			t.Errorf("fromCtxSafe[int16] = %v, want %v", got, int16(42))
		}
	})

	// Test for int64
	t.Run("Int64", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "int64Key", int64(42))
		if got := fromCtxSafe[int64](ctx, "int64Key"); got != int64(42) {
			t.Errorf("fromCtxSafe[int64] = %v, want %v", got, int64(42))
		}
	})
}
