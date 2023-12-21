package slutil

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

	execCtx := GetExecContext(1)
	assert.Equal(t, "TestGetExecContext", execCtx.Function)
	assert.Equal(t, "github.com/seantcanavan/zerolog-json-structured-logs", execCtx.Module)
	assert.Equal(t, "slutil", execCtx.Package)
	assert.Equal(t, 15, execCtx.Line)
	assert.Equal(t, cwd+"/exec_context_test.go", execCtx.File)
}
