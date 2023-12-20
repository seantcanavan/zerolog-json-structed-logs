package slutil

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestGetExecContext(t *testing.T) {
	// Get the current working directory
	cwd, err := os.Getwd()
	require.NoError(t, err)

	execCtx := testCaller()
	assert.Equal(t, cwd+"/exec_context_test.go", execCtx.File)
	assert.Equal(t, 16, execCtx.Line)
	assert.True(t, strings.HasSuffix(execCtx.Function, "zerolog-json-structured-logs/slutil.TestGetExecContext"))
}

// this wraps GetExecContext() to the correct depth to get the stack trace of the caller,
// not the stack trace for the go standard library (caller - 1)
func testCaller() ExecContext {
	return GetExecContext()
}
