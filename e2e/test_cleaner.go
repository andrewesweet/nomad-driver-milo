//go:build e2e

package e2e

// TestCleaner manages cleanup operations for tests
type TestCleaner struct {
	cleanupFunctions []func() error
}

// NewTestCleaner creates a new TestCleaner
func NewTestCleaner() *TestCleaner {
	return &TestCleaner{
		cleanupFunctions: make([]func() error, 0),
	}
}

// RegisterCleanup adds a cleanup function to be executed later
func (tc *TestCleaner) RegisterCleanup(fn func() error) {
	tc.cleanupFunctions = append(tc.cleanupFunctions, fn)
}

// ExecuteCleanup runs all registered cleanup functions
func (tc *TestCleaner) ExecuteCleanup() error {
	for _, fn := range tc.cleanupFunctions {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}