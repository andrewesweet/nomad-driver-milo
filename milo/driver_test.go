package milo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlugin(t *testing.T) {
	logger := hclog.NewNullLogger()
	plugin := NewPlugin(logger)

	require.NotNil(t, plugin)

	// Plugin is already of type drivers.DriverPlugin, no need to assert
}

func TestPluginInfo(t *testing.T) {
	logger := hclog.NewNullLogger()
	plugin := NewPlugin(logger)

	info, err := plugin.PluginInfo()
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, base.PluginTypeDriver, info.Type)
	assert.Equal(t, pluginName, info.Name)
	assert.Equal(t, pluginVersion, info.PluginVersion)
	assert.Contains(t, info.PluginApiVersions, drivers.ApiVersion010)
}

func TestConfigSchema(t *testing.T) {
	logger := hclog.NewNullLogger()
	plugin := NewPlugin(logger)

	schema, err := plugin.ConfigSchema()
	require.NoError(t, err)
	require.NotNil(t, schema)
}

func TestTaskConfigSchema(t *testing.T) {
	logger := hclog.NewNullLogger()
	plugin := NewPlugin(logger)

	schema, err := plugin.TaskConfigSchema()
	require.NoError(t, err)
	require.NotNil(t, schema)
}

func TestCapabilities(t *testing.T) {
	logger := hclog.NewNullLogger()
	plugin := NewPlugin(logger)

	caps, err := plugin.Capabilities()
	require.NoError(t, err)
	require.NotNil(t, caps)

	assert.True(t, caps.SendSignals)
	assert.False(t, caps.Exec)
}

func TestSetConfig(t *testing.T) {
	logger := hclog.NewNullLogger()
	plugin := NewPlugin(logger)

	// Test configuration with empty config (should fail because no shell specified)
	cfg := &base.Config{
		PluginConfig: []byte{},
	}
	err := plugin.SetConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid shell")

	// Test valid configuration with explicit shell
	validConfig := map[string]interface{}{
		"shell": "bash",
	}
	var configBytes []byte
	err = base.MsgPackEncode(&configBytes, validConfig)
	require.NoError(t, err)

	cfg2 := &base.Config{
		PluginConfig: configBytes,
	}
	err = plugin.SetConfig(cfg2)
	assert.NoError(t, err)
}

// === CYCLE 4: Test validation pipeline functions ===
// Phase: RED
func TestValidateTaskArtifacts(t *testing.T) {
	// Given a task directory with a JAR file
	taskDirPath := "/tmp/test-task-validate"
	localDir := filepath.Join(taskDirPath, "local")
	jarPath := filepath.Join(localDir, "test-app.jar")

	// Create the directory structure
	err := os.MkdirAll(localDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(taskDirPath)

	// Create a test JAR file
	content := "PK\x03\x04" // JAR file magic bytes
	err = os.WriteFile(jarPath, []byte(content), 0600)
	require.NoError(t, err)

	// When we validate artifacts in the pipeline
	artifactPath, err := FindArtifactInTaskDir(taskDirPath)
	require.NoError(t, err)

	// The JAR file should be found and validation should pass
	err = ValidateArtifactExtension(artifactPath)
	require.NoError(t, err, "JAR file should be valid")

	// But if we manually test a non-jar file
	err = ValidateArtifactExtension("/tmp/script.py")
	assert.Error(t, err, "non-jar files should fail validation")
}
