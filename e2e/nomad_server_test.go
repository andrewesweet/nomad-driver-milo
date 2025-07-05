//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UT1.1: Create NomadServer instance
func TestNewNomadServer_CreatesInstance(t *testing.T) {
	server := NewNomadServer(t)
	assert.NotNil(t, server)
	assert.NotNil(t, server.config)
	assert.Equal(t, "", server.configPath) // not generated yet
}

// UT1.2: Generate dynamic configuration
func TestGenerateConfig_CreatesDynamicPorts(t *testing.T) {
	server := NewNomadServer(t)
	err := server.GenerateConfig()
	require.NoError(t, err)
	assert.NotEqual(t, 0, server.config.HTTPPort)
	assert.NotEqual(t, 0, server.config.RPCPort)
	assert.NotEqual(t, 0, server.config.SerfPort)
	assert.True(t, isPortAvailable(server.config.HTTPPort))
}

// UT1.3: Generate configuration file from template
func TestGenerateConfigFile_CreatesValidConfig(t *testing.T) {
	server := NewNomadServer(t)
	server.GenerateConfig()
	
	err := server.GenerateConfigFile()
	require.NoError(t, err)
	assert.NotEqual(t, "", server.configPath)
	assert.FileExists(t, server.configPath)
}

// UT1.4: Start Nomad process
func TestStart_LaunchesNomadProcess(t *testing.T) {
	server := NewNomadServer(t)
	server.GenerateConfig()
	server.GenerateConfigFile()
	
	err := server.Start()
	require.NoError(t, err)
	assert.NotNil(t, server.process)
	assert.NotNil(t, server.process.Process)
	assert.True(t, isProcessRunning(server.process.Process.Pid))
}