//go:build e2e

package e2e

import (
	"testing"

	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
)

// UT3.1: Create OutputVerifier with API client
func TestNewOutputVerifier_CreatesWithClient(t *testing.T) {
	client := &api.Client{} // mock client
	verifier := NewOutputVerifier(client)
	assert.NotNil(t, verifier)
	assert.Equal(t, client, verifier.client)
}

// UT3.3: Verify log content matches patterns
func TestVerifyLogContent_MatchesExpectedPatterns(t *testing.T) {
	verifier := NewOutputVerifier(nil)

	logs := "Hello from JAR\nExecution completed successfully\n"

	// Should match expected content
	err := verifier.VerifyLogContent(logs, "Hello from JAR")
	assert.NoError(t, err)

	// Should fail for unexpected content
	err = verifier.VerifyLogContent(logs, "Error occurred")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected content not found")
}