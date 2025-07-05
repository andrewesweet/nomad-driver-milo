//go:build live_e2e

package live

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	testServer        *LiveNomadServer
	artifactServerURL string
)

func TestMain(m *testing.M) {
	// 1. Start HTTP artifact server
	artifactServer := httptest.NewServer(http.FileServer(http.Dir("../../test-artifacts")))
	artifactServerURL = artifactServer.URL

	// 2. Start shared Nomad server
	t := &testing.T{}
	server := NewLiveNomadServer(t)
	if err := server.Start(); err != nil {
		artifactServer.Close()
		server.Stop()
		os.Exit(1)
	}
	testServer = server

	// 3. Run tests
	exitCode := m.Run()

	// 4. Cleanup
	artifactServer.Close()
	testServer.Stop()
	os.Exit(exitCode)
}

// sanitizeForNomadJobID converts test names to valid Nomad job IDs
func sanitizeForNomadJobID(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	sanitized := reg.ReplaceAllString(name, "-")
	sanitized = strings.Trim(sanitized, "-")
	return strings.ToLower(sanitized)
}

// generateTestJobID creates unique job IDs for parallel tests
func generateTestJobID(t *testing.T) string {
	rand.Seed(time.Now().UnixNano())
	baseName := sanitizeForNomadJobID(t.Name())
	return fmt.Sprintf("%s-%d", baseName, rand.Intn(10000))
}