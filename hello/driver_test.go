package hello

import (
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
	
	// Test that the plugin implements the DriverPlugin interface
	_, ok := plugin.(drivers.DriverPlugin)
	assert.True(t, ok, "NewPlugin should return a DriverPlugin")
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
	assert.Error(t, err)
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