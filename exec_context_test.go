package sl

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetExecContext(t *testing.T) {
	// Get the current working directory
	cwd, err := os.Getwd()
	require.NoError(t, err)

	execCtx := testCaller()
	assert.Equal(t, cwd+"/exec_context_test.go", execCtx.File)
	assert.Equal(t, 15, execCtx.Line)
	assert.Equal(t, "github.com/seantcanavan/zerolog-json-structured-logs.TestGetExecContext", execCtx.Function)
}

// this wraps GetExecContext() to the correct depth to get the stack trace of the caller,
// not the stack trace for the go standard library (caller - 1)
func testCaller() ExecContext {
	return getExecContext()
}
