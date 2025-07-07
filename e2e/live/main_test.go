//go:build live_e2e

package live

import (
	"fmt"
	"os"
	"testing"
	
	"github.com/andrewesweet/nomad-driver-milo/e2e/shared"
)

var (
	testServer        *shared.LiveNomadServer
	artifactServer    *shared.ArtifactServer
	artifactServerURL string
)

func TestMain(m *testing.M) {
	// 1. Start HTTP artifact server using shared infrastructure
	artifactServer = shared.NewArtifactServer("../../test-artifacts")
	artifactServerURL = artifactServer.URL()

	// 2. Start shared Nomad server
	testServer = shared.NewLiveNomadServer()
	if err := testServer.Start(); err != nil {
		artifactServer.Close()
		testServer.Stop()
		os.Exit(1)
	}

	// Set NOMAD_ADDR environment variable for all tests
	nomadAddr := fmt.Sprintf("http://127.0.0.1:%d", testServer.GetHTTPPort())
	os.Setenv("NOMAD_ADDR", nomadAddr)

	// 3. Run tests
	exitCode := m.Run()

	// 4. Cleanup
	artifactServer.Close()
	testServer.Stop()
	os.Exit(exitCode)
}

// generateTestJobID creates unique job IDs for parallel tests using shared helper
func generateTestJobID(t *testing.T) string {
	return shared.GenerateTestJobID(t.Name())
}