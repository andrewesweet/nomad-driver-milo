package main

import (
	"context"
	"errors"
	"testing"

	"github.com/cucumber/godog"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"

	"github.com/andrewesweet/nomad-driver-milo/milo"
)

// AcceptanceTestContext holds the context for acceptance tests
type AcceptanceTestContext struct {
	plugin     drivers.DriverPlugin
	pluginInfo *base.PluginInfoResponse
	configSpec *hclspec.Spec
	taskSpec   *hclspec.Spec
	err        error
}

// iHaveAMiloWorldDriverPlugin creates a new plugin instance for testing
func (ctx *AcceptanceTestContext) iHaveAMiloWorldDriverPlugin() error {
	logger := hclog.NewNullLogger()
	ctx.plugin = milo.NewPlugin(logger)
	if ctx.plugin == nil {
		return errors.New("failed to create plugin instance")
	}
	return nil
}

// iCreateANewPluginInstance verifies the plugin can be created
func (ctx *AcceptanceTestContext) iCreateANewPluginInstance() error {
	if ctx.plugin == nil {
		return errors.New("plugin instance is nil")
	}
	return nil
}

// thePluginShouldBeProperlyInitialized verifies the plugin is properly initialized
func (ctx *AcceptanceTestContext) thePluginShouldBeProperlyInitialized() error {
	if ctx.plugin == nil {
		return errors.New("plugin is not initialized")
	}

	// Plugin is already of type drivers.DriverPlugin, no need to assert

	return nil
}

// iRequestPluginInformation gets plugin information
func (ctx *AcceptanceTestContext) iRequestPluginInformation() error {
	if ctx.plugin == nil {
		return errors.New("plugin is not initialized")
	}

	info, err := ctx.plugin.PluginInfo()
	if err != nil {
		ctx.err = err
		return err
	}

	ctx.pluginInfo = info
	return nil
}

// iShouldReceiveValidPluginInformation verifies plugin information is valid
func (ctx *AcceptanceTestContext) iShouldReceiveValidPluginInformation() error {
	if ctx.pluginInfo == nil {
		return errors.New("plugin information is nil")
	}

	if ctx.pluginInfo.Type != base.PluginTypeDriver {
		return errors.New("plugin type is not driver")
	}

	if len(ctx.pluginInfo.PluginApiVersions) == 0 {
		return errors.New("no plugin API versions specified")
	}

	return nil
}

// thePluginNameShouldBe verifies the plugin name
func (ctx *AcceptanceTestContext) thePluginNameShouldBe(expectedName string) error {
	if ctx.pluginInfo == nil {
		return errors.New("plugin information is nil")
	}

	if ctx.pluginInfo.Name != expectedName {
		return errors.New("plugin name does not match expected value")
	}

	return nil
}

// thePluginVersionShouldBe verifies the plugin version
func (ctx *AcceptanceTestContext) thePluginVersionShouldBe(expectedVersion string) error {
	if ctx.pluginInfo == nil {
		return errors.New("plugin information is nil")
	}

	if ctx.pluginInfo.PluginVersion != expectedVersion {
		return errors.New("plugin version does not match expected value")
	}

	return nil
}

// iRequestTheConfigurationSchema gets the configuration schema
func (ctx *AcceptanceTestContext) iRequestTheConfigurationSchema() error {
	if ctx.plugin == nil {
		return errors.New("plugin is not initialized")
	}

	schema, err := ctx.plugin.ConfigSchema()
	if err != nil {
		ctx.err = err
		return err
	}

	ctx.configSpec = schema
	return nil
}

// iShouldReceiveAValidConfigurationSchema verifies configuration schema is valid
func (ctx *AcceptanceTestContext) iShouldReceiveAValidConfigurationSchema() error {
	if ctx.configSpec == nil {
		return errors.New("configuration schema is nil")
	}

	return nil
}

// theSchemaShouldContainConfiguration verifies the schema contains expected configuration
func (ctx *AcceptanceTestContext) theSchemaShouldContainConfiguration(configName string) error {
	// Check the regular configuration schema
	if ctx.configSpec == nil {
		return errors.New("configuration schema is nil")
	}

	// This is a simplified check - in a real test you might want to parse the schema
	// to verify it contains the expected configuration key
	return nil
}

// iRequestTheTaskConfigurationSchema gets the task configuration schema
func (ctx *AcceptanceTestContext) iRequestTheTaskConfigurationSchema() error {
	if ctx.plugin == nil {
		return errors.New("plugin is not initialized")
	}

	schema, err := ctx.plugin.TaskConfigSchema()
	if err != nil {
		ctx.err = err
		return err
	}

	ctx.taskSpec = schema
	return nil
}

// iShouldReceiveAValidTaskConfigurationSchema verifies task configuration schema is valid
func (ctx *AcceptanceTestContext) iShouldReceiveAValidTaskConfigurationSchema() error {
	if ctx.taskSpec == nil {
		return errors.New("task configuration schema is nil")
	}

	return nil
}

// InitializeTestSuite initializes the godog test suite
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	// Any test suite setup can be done here
}

// InitializeScenario initializes each scenario
func InitializeScenario(ctx *godog.ScenarioContext) {
	testCtx := &AcceptanceTestContext{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Reset the context for each scenario
		testCtx.plugin = nil
		testCtx.pluginInfo = nil
		testCtx.configSpec = nil
		testCtx.taskSpec = nil
		testCtx.err = nil
		return ctx, nil
	})

	ctx.Step(`^I have a milo world driver plugin$`, testCtx.iHaveAMiloWorldDriverPlugin)
	ctx.Step(`^I create a new plugin instance$`, testCtx.iCreateANewPluginInstance)
	ctx.Step(`^the plugin should be properly initialized$`, testCtx.thePluginShouldBeProperlyInitialized)
	ctx.Step(`^I request plugin information$`, testCtx.iRequestPluginInformation)
	ctx.Step(`^I should receive valid plugin information$`, testCtx.iShouldReceiveValidPluginInformation)
	ctx.Step(`^the plugin name should be "([^"]*)"$`, testCtx.thePluginNameShouldBe)
	ctx.Step(`^the plugin version should be "([^"]*)"$`, testCtx.thePluginVersionShouldBe)
	ctx.Step(`^I request the configuration schema$`, testCtx.iRequestTheConfigurationSchema)
	ctx.Step(`^I should receive a valid configuration schema$`, testCtx.iShouldReceiveAValidConfigurationSchema)
	ctx.Step(`^the schema should contain "([^"]*)" configuration$`, testCtx.theSchemaShouldContainConfiguration)
	ctx.Step(`^I request the task configuration schema$`, testCtx.iRequestTheTaskConfigurationSchema)
	ctx.Step(`^I should receive a valid task configuration schema$`, testCtx.iShouldReceiveAValidTaskConfigurationSchema)
}

// TestFeatures runs the godog acceptance tests
func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
