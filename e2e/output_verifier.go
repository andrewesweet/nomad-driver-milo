//go:build e2e

package e2e

import (
	"fmt"
	"strings"

	"github.com/hashicorp/nomad/api"
)

// OutputVerifier validates job outputs and behavior
type OutputVerifier struct {
	client *api.Client
}

// NewOutputVerifier creates a new OutputVerifier with the given Nomad API client
func NewOutputVerifier(client *api.Client) *OutputVerifier {
	return &OutputVerifier{
		client: client,
	}
}

// VerifyLogContent checks if logs contain expected content
func (ov *OutputVerifier) VerifyLogContent(logs, expectedContent string) error {
	if !strings.Contains(logs, expectedContent) {
		return fmt.Errorf("expected content not found: %q not in logs", expectedContent)
	}
	return nil
}