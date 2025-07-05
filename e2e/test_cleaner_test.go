//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// UT4.1: Create TestCleaner manager
func TestNewTestCleaner_CreatesCleanupManager(t *testing.T) {
	cleaner := NewTestCleaner()
	assert.NotNil(t, cleaner)
	assert.Empty(t, cleaner.cleanupFunctions)
}

// UT4.2: Register cleanup functions
func TestRegisterCleanup_AddsCleanupFunction(t *testing.T) {
	cleaner := NewTestCleaner()
	called := false

	cleaner.RegisterCleanup(func() error {
		called = true
		return nil
	})

	assert.Len(t, cleaner.cleanupFunctions, 1)

	cleaner.ExecuteCleanup()
	assert.True(t, called)
}

// UT4.3: Execute all cleanup operations
func TestExecuteCleanup_RunsAllCleanupOperations(t *testing.T) {
	cleaner := NewTestCleaner()
	count := 0

	cleaner.RegisterCleanup(func() error { count++; return nil })
	cleaner.RegisterCleanup(func() error { count++; return nil })
	cleaner.RegisterCleanup(func() error { count++; return nil })

	err := cleaner.ExecuteCleanup()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}