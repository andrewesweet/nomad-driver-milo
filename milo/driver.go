// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package milo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/consul-template/signals"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/drivers/shared/eventer"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

const (
	// pluginName is the name of the plugin
	// this is used for logging and (along with the version) for uniquely
	// identifying plugin binaries fingerprinted by the client
	pluginName = "milo"

	// pluginVersion allows the client to identify and use newer versions of
	// an installed plugin
	pluginVersion = "v0.1.0"

	// fingerprintPeriod is the interval at which the plugin will send
	// fingerprint responses
	fingerprintPeriod = 30 * time.Second

	// taskHandleVersion is the version of task handle which this plugin sets
	// and understands how to decode
	// this is used to allow modification and migration of the task schema
	// used by the plugin
	taskHandleVersion = 1
)

var (
	// pluginInfo describes the plugin
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDriver,
		PluginApiVersions: []string{drivers.ApiVersion010},
		PluginVersion:     pluginVersion,
		Name:              pluginName,
	}

	// configSpec is the specification of the plugin's configuration
	// this is used to validate the configuration specified for the plugin
	// on the client.
	// this is not global, but can be specified on a per-client basis.
	configSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		// TODO: define plugin's agent configuration schema.
		//
		// The schema should be defined using HCL specs and it will be used to
		// validate the agent configuration provided by the user in the
		// `plugin` stanza (https://www.nomadproject.io/docs/configuration/plugin.html).
		//
		// For example, for the schema below a valid configuration would be:
		//
		//   plugin "hello-driver-plugin" {
		//     config {
		//       shell = "fish"
		//     }
		//   }
		"shell": hclspec.NewDefault(
			hclspec.NewAttr("shell", "string", false),
			hclspec.NewLiteral(`"bash"`),
		),
	})

	// taskConfigSpec is the specification of the plugin's configuration for
	// a task
	// this is used to validated the configuration specified for the plugin
	// when a job is submitted.
	taskConfigSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		// Minimal dummy config to make validation pass
		"dummy": hclspec.NewDefault(
			hclspec.NewAttr("dummy", "string", false),
			hclspec.NewLiteral(`""`),
		),
	})

	// capabilities indicates what optional features this driver supports
	// this should be set according to the target run time.
	capabilities = &drivers.Capabilities{
		// TODO: set plugin's capabilities
		//
		// The plugin's capabilities signal Nomad which extra functionalities
		// are supported. For a list of available options check the docs page:
		// https://godoc.org/github.com/hashicorp/nomad/plugins/drivers#Capabilities
		SendSignals: true,
		Exec:        false,
	}
)

// Config contains configuration information for the plugin
type Config struct {
	// TODO: create decoded plugin configuration struct
	//
	// This struct is the decoded version of the schema defined in the
	// configSpec variable above. It's used to convert the HCL configuration
	// passed by the Nomad agent into Go contructs.
	Shell string `codec:"shell"`
}

// TaskConfig contains configuration information for a task that runs with
// this plugin
type TaskConfig struct {
	// Dummy field to match the taskConfigSpec
	Dummy string `codec:"dummy"`
}

// TaskState is the runtime state which is encoded in the handle returned to
// Nomad client.
// This information is needed to rebuild the task state and handler during
// recovery.
type TaskState struct {
	TaskConfig *drivers.TaskConfig
	StartedAt  time.Time
	Pid        int

	// Note: We don't store ReattachConfig since we're using direct exec.Command
	// which doesn't support reattachment. In a production driver, you might
	// want to use a different approach that supports recovery.
}

// MiloDriverPlugin is an example driver plugin. When provisioned in a job,
// the taks will output a greet specified by the user.
type MiloDriverPlugin struct {
	// eventer is used to handle multiplexing of TaskEvents calls such that an
	// event can be broadcast to all callers
	eventer *eventer.Eventer

	// config is the plugin configuration set by the SetConfig RPC
	config *Config

	// nomadConfig is the client config from Nomad
	nomadConfig *base.ClientDriverConfig

	// tasks is the in memory datastore mapping taskIDs to driver handles
	tasks *taskStore

	// ctx is the context for the driver. It is passed to other subsystems to
	// coordinate shutdown
	ctx context.Context

	// signalShutdown is called when the driver is shutting down and cancels
	// the ctx passed to any subsystems
	signalShutdown context.CancelFunc

	// logger will log to the Nomad agent
	logger hclog.Logger
}

// NewPlugin returns a new example driver plugin
func NewPlugin(logger hclog.Logger) drivers.DriverPlugin {
	ctx, cancel := context.WithCancel(context.Background())
	logger = logger.Named(pluginName)

	return &MiloDriverPlugin{
		eventer:        eventer.NewEventer(ctx, logger),
		config:         &Config{},
		tasks:          newTaskStore(),
		ctx:            ctx,
		signalShutdown: cancel,
		logger:         logger,
	}
}

// PluginInfo returns information describing the plugin.
func (d *MiloDriverPlugin) PluginInfo() (*base.PluginInfoResponse, error) {
	return pluginInfo, nil
}

// ConfigSchema returns the plugin configuration schema.
func (d *MiloDriverPlugin) ConfigSchema() (*hclspec.Spec, error) {
	return configSpec, nil
}

// SetConfig is called by the client to pass the configuration for the plugin.
func (d *MiloDriverPlugin) SetConfig(cfg *base.Config) error {
	var config Config
	if len(cfg.PluginConfig) != 0 {
		if err := base.MsgPackDecode(cfg.PluginConfig, &config); err != nil {
			return err
		}
	}

	// Save the configuration to the plugin
	d.config = &config

	// TODO: parse and validated any configuration value if necessary.
	//
	// If your driver agent configuration requires any complex validation
	// (some dependency between attributes) or special data parsing (the
	// string "10s" into a time.Interval) you can do it here and update the
	// value in d.config.
	//
	// In the example below we check if the shell specified by the user is
	// supported by the plugin.
	shell := d.config.Shell
	if shell != "bash" && shell != "fish" {
		return fmt.Errorf("invalid shell %s", d.config.Shell)
	}

	// Save the Nomad agent configuration
	if cfg.AgentConfig != nil {
		d.nomadConfig = cfg.AgentConfig.Driver
	}

	// TODO: initialize any extra requirements if necessary.
	//
	// Here you can use the config values to initialize any resources that are
	// shared by all tasks that use this driver, such as a daemon process.

	return nil
}

// TaskConfigSchema returns the HCL schema for the configuration of a task.
func (d *MiloDriverPlugin) TaskConfigSchema() (*hclspec.Spec, error) {
	return taskConfigSpec, nil
}

// Capabilities returns the features supported by the driver.
func (d *MiloDriverPlugin) Capabilities() (*drivers.Capabilities, error) {
	return capabilities, nil
}

// Fingerprint returns a channel that will be used to send health information
// and other driver specific node attributes.
func (d *MiloDriverPlugin) Fingerprint(ctx context.Context) (<-chan *drivers.Fingerprint, error) {
	ch := make(chan *drivers.Fingerprint)
	go d.handleFingerprint(ctx, ch)
	return ch, nil
}

// handleFingerprint manages the channel and the flow of fingerprint data.
func (d *MiloDriverPlugin) handleFingerprint(ctx context.Context, ch chan<- *drivers.Fingerprint) {
	defer close(ch)

	// Nomad expects the initial fingerprint to be sent immediately
	ticker := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			// after the initial fingerprint we can set the proper fingerprint
			// period
			ticker.Reset(fingerprintPeriod)
			ch <- d.buildFingerprint()
		}
	}
}

// buildFingerprint returns the driver's fingerprint data
func (d *MiloDriverPlugin) buildFingerprint() *drivers.Fingerprint {
	fp := &drivers.Fingerprint{
		Attributes:        map[string]*structs.Attribute{},
		Health:            drivers.HealthStateHealthy,
		HealthDescription: drivers.DriverHealthy,
	}

	// TODO: implement fingerprinting logic to populate health and driver
	// attributes.
	//
	// Fingerprinting is used by the plugin to relay two important information
	// to Nomad: health state and node attributes.
	//
	// If the plugin reports to be unhealthy, or doesn't send any fingerprint
	// data in the expected interval of time, Nomad will restart it.
	//
	// Node attributes can be used to report any relevant information about
	// the node in which the plugin is running (specific library availability,
	// installed versions of a software etc.). These attributes can then be
	// used by an operator to set job constrains.
	//
	// In the example below we check if the shell specified by the user exists
	// in the node.
	shell := d.config.Shell

	cmd := exec.Command("which", shell)
	if err := cmd.Run(); err != nil {
		return &drivers.Fingerprint{
			Health:            drivers.HealthStateUndetected,
			HealthDescription: fmt.Sprintf("shell %s not found", shell),
		}
	}

	// We also set the shell and its version as attributes
	cmd = exec.Command(shell, "--version")
	if out, err := cmd.Output(); err != nil {
		d.logger.Warn("failed to find shell version: %v", err)
	} else {
		re := regexp.MustCompile(`[0-9]\.[0-9]\.[0-9]`)
		version := re.FindString(string(out))

		fp.Attributes["driver.milo.shell_version"] = structs.NewStringAttribute(version)
		fp.Attributes["driver.milo.shell"] = structs.NewStringAttribute(shell)
	}

	return fp
}

// StartTask returns a task handle and a driver network if necessary.
func (d *MiloDriverPlugin) StartTask(cfg *drivers.TaskConfig) (*drivers.TaskHandle, *drivers.DriverNetwork, error) {
	if _, ok := d.tasks.Get(cfg.ID); ok {
		return nil, nil, fmt.Errorf("task with ID %q already started", cfg.ID)
	}

	var driverConfig TaskConfig
	if err := cfg.DecodeDriverConfig(&driverConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to decode driver config: %v", err)
	}

	d.logger.Info("starting task", "driver_cfg", hclog.Fmt("%+v", driverConfig))
	handle := drivers.NewTaskHandle(taskHandleVersion)
	handle.Config = cfg

	// Validate artifacts before starting the task using comprehensive validator
	validator := NewArtifactValidator(cfg.TaskDir().Dir)
	artifactPath, err := validator.FindAndValidateArtifact()
	if err != nil {
		// Return the error directly since it already has user-friendly formatting
		return nil, nil, err
	}

	// Detect Java runtime on the host
	commonJavaPaths := []string{
		"/usr/lib/jvm/java-21-openjdk-amd64",
		"/usr/lib/jvm/java-17-openjdk-amd64",
		"/usr/lib/jvm/java-11-openjdk-amd64",
		"/usr/lib/jvm/java-17",
		"/usr/lib/jvm/java-11",
		"/usr/lib/jvm/java-8",
		"/opt/java",
		"/usr/java",
	}

	// Use common Java paths (test mode handled in BDD tests via environment)
	searchPaths := commonJavaPaths

	javaHome, err := DetectJavaRuntime(searchPaths)
	if err != nil {
		return nil, nil, fmt.Errorf("Java runtime detection failed: %v", err)
	}

	d.logger.Info("detected Java runtime", "java_home", javaHome)

	// Create container bundle for crun execution
	bundlePath := filepath.Join(cfg.TaskDir().Dir, "container-bundle")
	// Sanitize container ID by replacing slashes with hyphens (crun doesn't allow slashes in container IDs)
	sanitizedID := strings.ReplaceAll(cfg.ID, "/", "-")
	containerID := fmt.Sprintf("milo-task-%s", sanitizedID)

	// Convert host artifact path to container path
	// The task directory is mounted to /app, so convert the path accordingly
	relativeArtifactPath, err := filepath.Rel(cfg.TaskDir().Dir, artifactPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get relative artifact path: %v", err)
	}
	containerArtifactPath := filepath.Join("/app", relativeArtifactPath)

	// Create OCI specification for the JAR execution
	spec, err := CreateOCISpec(javaHome, containerArtifactPath, cfg.TaskDir().Dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OCI spec: %v", err)
	}

	// OCI spec logging for testing can be added later if needed

	// Create the container bundle
	if err := CreateContainerBundle(bundlePath, spec); err != nil {
		return nil, nil, fmt.Errorf("failed to create container bundle: %v", err)
	}

	// Generate crun command
	crunCmd := GenerateCrunCommand(bundlePath, containerID)
	d.logger.Info("executing JAR with crun", "command", crunCmd)

	// Create crun command using exec.Command for better streaming control
	// #nosec G204 - crunCmd is generated internally with validated inputs
	cmd := exec.Command(crunCmd[0], crunCmd[1:]...)

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start crun: %v", err)
	}

	// Create context for log streaming
	ctx, cancel := context.WithCancel(context.Background())

	// Create task handle
	h := &taskHandle{
		cmd:          cmd,
		pid:          cmd.Process.Pid,
		taskConfig:   cfg,
		procState:    drivers.TaskStateRunning,
		startedAt:    time.Now().Round(time.Millisecond),
		logger:       d.logger,
		ctx:          ctx,
		cancelFunc:   cancel,
		waitCh:       make(chan struct{}),
		stdoutStream: nil, // Will be set below
		stderrStream: nil, // Will be set below
	}

	// Create and start log streamers
	h.stdoutStream = NewLogStreamer(d.logger.Named("stdout"), cfg.StdoutPath, stdoutPipe)
	h.stderrStream = NewLogStreamer(d.logger.Named("stderr"), cfg.StderrPath, stderrPipe)

	// Start streaming goroutines
	go func() {
		if err := h.stdoutStream.Stream(h.ctx); err != nil {
			d.logger.Error("stdout streaming failed", "error", err)
		}
	}()

	go func() {
		if err := h.stderrStream.Stream(h.ctx); err != nil {
			d.logger.Error("stderr streaming failed", "error", err)
		}
	}()

	driverState := TaskState{
		Pid:        h.pid,
		TaskConfig: cfg,
		StartedAt:  h.startedAt,
	}

	if err := handle.SetDriverState(&driverState); err != nil {
		return nil, nil, fmt.Errorf("failed to set driver state: %v", err)
	}

	d.tasks.Set(cfg.ID, h)
	go h.run()
	return handle, nil, nil
}

// RecoverTask recreates the in-memory state of a task from a TaskHandle.
func (d *MiloDriverPlugin) RecoverTask(handle *drivers.TaskHandle) error {
	if handle == nil {
		return errors.New("error: handle cannot be nil")
	}

	if _, ok := d.tasks.Get(handle.Config.ID); ok {
		return nil
	}

	var taskState TaskState
	if err := handle.GetDriverState(&taskState); err != nil {
		return fmt.Errorf("failed to decode task state from handle: %v", err)
	}

	var driverConfig TaskConfig
	if err := taskState.TaskConfig.DecodeDriverConfig(&driverConfig); err != nil {
		return fmt.Errorf("failed to decode driver config: %v", err)
	}

	// For now, we cannot recover tasks since we're using direct exec.Command
	// which doesn't support reattachment. In a production driver, we might:
	// 1. Use a supervisor process that can be reattached
	// 2. Store enough state to recreate the process monitoring
	// 3. Check if the process is still running via PID

	// Mark the task as lost since we can't reattach to it
	d.logger.Warn("cannot recover task - direct exec doesn't support reattachment",
		"task_id", handle.Config.ID,
		"pid", taskState.Pid)

	// We could check if the process is still running and create a minimal handle
	// but for now, we'll just return an error to indicate the task is lost
	return fmt.Errorf("task recovery not supported with direct exec.Command")
}

// WaitTask returns a channel used to notify Nomad when a task exits.
func (d *MiloDriverPlugin) WaitTask(ctx context.Context, taskID string) (<-chan *drivers.ExitResult, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	ch := make(chan *drivers.ExitResult)
	go d.handleWait(ctx, handle, ch)
	return ch, nil
}

func (d *MiloDriverPlugin) handleWait(ctx context.Context, handle *taskHandle, ch chan *drivers.ExitResult) {
	defer close(ch)

	// Wait for the task to complete
	select {
	case <-handle.waitCh:
		// Task completed, get the result
		handle.stateLock.RLock()
		result := handle.exitResult
		handle.stateLock.RUnlock()

		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case ch <- result:
			return
		}

	case <-ctx.Done():
		return
	case <-d.ctx.Done():
		return
	}
}

// StopTask stops a running task with the given signal and within the timeout window.
func (d *MiloDriverPlugin) StopTask(taskID string, timeout time.Duration, signal string) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	// Stop the command using signal
	if handle.cmd != nil && handle.cmd.Process != nil {
		// Convert signal name to syscall.Signal
		sig := signals.SignalLookup[signal]
		if sig == nil {
			// Default to SIGTERM if signal not found
			sig = signals.SignalLookup["SIGTERM"]
		}

		// Send signal to the process
		if err := handle.cmd.Process.Signal(sig); err != nil {
			d.logger.Warn("failed to send signal", "signal", signal, "error", err)
		}

		// Wait for timeout
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-handle.waitCh:
			// Process exited gracefully
			return nil
		case <-timer.C:
			// Timeout exceeded, force kill
			if err := handle.cmd.Process.Kill(); err != nil {
				d.logger.Error("failed to kill process", "error", err)
			}
		}
	}

	// Cancel log streaming
	if handle.cancelFunc != nil {
		handle.cancelFunc()
	}

	return nil
}

// DestroyTask cleans up and removes a task that has terminated.
func (d *MiloDriverPlugin) DestroyTask(taskID string, force bool) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	if handle.IsRunning() && !force {
		return errors.New("cannot destroy running task")
	}

	// If force is set, kill the process immediately
	if force && handle.cmd != nil && handle.cmd.Process != nil {
		if err := handle.cmd.Process.Kill(); err != nil {
			handle.logger.Error("failed to force kill process", "err", err)
		}
	}

	// Cancel log streaming context
	if handle.cancelFunc != nil {
		handle.cancelFunc()
	}

	// Clean up any container resources
	// The container should already be cleaned up by crun when process exits
	// but we can add additional cleanup here if needed

	d.tasks.Delete(taskID)
	return nil
}

// InspectTask returns detailed status information for the referenced taskID.
func (d *MiloDriverPlugin) InspectTask(taskID string) (*drivers.TaskStatus, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	return handle.TaskStatus(), nil
}

// TaskStats returns a channel which the driver should send stats to at the given interval.
func (d *MiloDriverPlugin) TaskStats(ctx context.Context, taskID string, interval time.Duration) (<-chan *drivers.TaskResourceUsage, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	// TODO: implement driver specific logic to send task stats.
	//
	// This function returns a channel that Nomad will use to listen for task
	// stats (e.g., CPU and memory usage) in a given interval. It should send
	// stats until the context is canceled or the task stops running.
	//
	// In the example below we use the Stats function provided by the executor,
	// but you can build a set of functions similar to the fingerprint process.
	return handle.exec.Stats(ctx, interval)
}

// TaskEvents returns a channel that the plugin can use to emit task related events.
func (d *MiloDriverPlugin) TaskEvents(ctx context.Context) (<-chan *drivers.TaskEvent, error) {
	return d.eventer.TaskEvents(ctx)
}

// SignalTask forwards a signal to a task.
// This is an optional capability.
func (d *MiloDriverPlugin) SignalTask(taskID string, signal string) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	// TODO: implement driver specific signal handling logic.
	//
	// The given signal must be forwarded to the target taskID. If this plugin
	// doesn't support receiving signals (capability SendSignals is set to
	// false) you can just return nil.
	sig := os.Interrupt
	if s, ok := signals.SignalLookup[signal]; ok {
		sig = s
	} else {
		d.logger.Warn("unknown signal to send to task, using SIGINT instead", "signal", signal, "task_id", handle.taskConfig.ID)

	}
	return handle.exec.Signal(sig)
}

// ExecTask returns the result of executing the given command inside a task.
// This is an optional capability.
func (d *MiloDriverPlugin) ExecTask(taskID string, cmd []string, timeout time.Duration) (*drivers.ExecTaskResult, error) {
	// TODO: implement driver specific logic to execute commands in a task.
	return nil, errors.New("This driver does not support exec")
}
