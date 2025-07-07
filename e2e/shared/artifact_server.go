package shared

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
)

// ArtifactServer provides HTTP access to test artifacts
type ArtifactServer struct {
	server  *httptest.Server
	baseDir string
}

// NewArtifactServer creates a new HTTP server for serving test artifacts
func NewArtifactServer(artifactDir string) *ArtifactServer {
	// Resolve absolute path to handle relative paths
	absPath, _ := filepath.Abs(artifactDir)
	
	server := httptest.NewServer(http.FileServer(http.Dir(absPath)))
	
	return &ArtifactServer{
		server:  server,
		baseDir: absPath,
	}
}

// URL returns the base URL of the artifact server
func (a *ArtifactServer) URL() string {
	return a.server.URL
}

// Close shuts down the artifact server
func (a *ArtifactServer) Close() {
	a.server.Close()
}

// GetArtifactURL returns the full URL for a specific artifact
func (a *ArtifactServer) GetArtifactURL(filename string) string {
	return a.server.URL + "/" + filename
}

// BaseDir returns the base directory being served
func (a *ArtifactServer) BaseDir() string {
	return a.baseDir
}