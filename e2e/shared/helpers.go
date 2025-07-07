package shared

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// GenerateTestJobID creates a unique job ID for a test
// It sanitizes the test name and adds a random suffix to ensure uniqueness
func GenerateTestJobID(testName string) string {
	// Sanitize test name for Nomad job ID requirements
	sanitized := SanitizeForNomadJobID(testName)
	
	// Add random suffix for uniqueness
	suffix := generateRandomSuffix()
	
	return fmt.Sprintf("%s-%s", sanitized, suffix)
}

// SanitizeForNomadJobID converts a string to be valid as a Nomad job ID
// Nomad job IDs must match: ^[a-zA-Z0-9-_]+$
func SanitizeForNomadJobID(input string) string {
	// Replace invalid characters with hyphens
	sanitized := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
		   (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, input)
	
	// Remove leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")
	
	// Collapse multiple hyphens
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}
	
	// Ensure it starts with a letter (Nomad requirement)
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "job-" + sanitized
	}
	
	// Limit length to 63 characters (leaving room for suffix)
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}
	
	return sanitized
}

// generateRandomSuffix creates a short random string for test uniqueness
func generateRandomSuffix() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	suffix := make([]byte, 6)
	for i := range suffix {
		suffix[i] = charset[r.Intn(len(charset))]
	}
	
	return string(suffix)
}

// WaitForCondition polls a condition function until it returns true or timeout
func WaitForCondition(timeout time.Duration, interval time.Duration, condition func() bool) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if condition() {
			return nil
		}
		time.Sleep(interval)
	}
	
	return fmt.Errorf("timeout after %v waiting for condition", timeout)
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}