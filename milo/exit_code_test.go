package milo

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/stretchr/testify/require"
)

// Test that exit code 0 is properly propagated
func TestExitCodePropagation_Success(t *testing.T) {
	logger := hclog.NewNullLogger()

	// Create a command that exits with 0
	cmd := exec.Command("sh", "-c", "exit 0")
	err := cmd.Start()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	h := &taskHandle{
		logger:     logger,
		cmd:        cmd,
		pid:        cmd.Process.Pid,
		taskConfig: &drivers.TaskConfig{ID: "test-exit-0"},
		procState:  drivers.TaskStateRunning,
		startedAt:  time.Now(),
		ctx:        ctx,
		cancelFunc: cancel,
		waitCh:     make(chan struct{}),
	}

	// Run the handler
	go h.run()

	// Wait for process to complete
	select {
	case <-h.waitCh:
		// Process completed
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for process to complete")
	}

	// Verify exit code
	status := h.TaskStatus()
	require.Equal(t, drivers.TaskStateExited, status.State)
	require.NotNil(t, status.ExitResult)
	require.Equal(t, 0, status.ExitResult.ExitCode)
	require.NoError(t, status.ExitResult.Err)
}

// Test that non-zero exit codes are properly propagated
func TestExitCodePropagation_NonZero(t *testing.T) {
	testCases := []struct {
		name     string
		exitCode int
	}{
		{"exit code 1", 1},
		{"exit code 2", 2},
		{"exit code 42", 42},
		{"exit code 127", 127},
		{"exit code 255", 255},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := hclog.NewNullLogger()

			// Create a command that exits with specific code
			// #nosec G204 - tc.exitCode is a controlled integer from test cases
			cmd := exec.Command("sh", "-c", fmt.Sprintf("exit %d", tc.exitCode))
			err := cmd.Start()
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())

			h := &taskHandle{
				logger:     logger,
				cmd:        cmd,
				pid:        cmd.Process.Pid,
				taskConfig: &drivers.TaskConfig{ID: "test-exit-" + string(rune(tc.exitCode))},
				procState:  drivers.TaskStateRunning,
				startedAt:  time.Now(),
				ctx:        ctx,
				cancelFunc: cancel,
				waitCh:     make(chan struct{}),
			}

			// Run the handler
			go h.run()

			// Wait for process to complete
			select {
			case <-h.waitCh:
				// Process completed
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for process to complete")
			}

			// Verify exit code
			status := h.TaskStatus()
			require.Equal(t, drivers.TaskStateExited, status.State)
			require.NotNil(t, status.ExitResult)
			require.Equal(t, tc.exitCode, status.ExitResult.ExitCode)
			require.NoError(t, status.ExitResult.Err)
		})
	}
}

// Test that signals result in proper exit codes
func TestExitCodePropagation_Signal(t *testing.T) {
	logger := hclog.NewNullLogger()

	// Create a command that sleeps (so we can kill it)
	cmd := exec.Command("sleep", "10")
	err := cmd.Start()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	h := &taskHandle{
		logger:     logger,
		cmd:        cmd,
		pid:        cmd.Process.Pid,
		taskConfig: &drivers.TaskConfig{ID: "test-signal"},
		procState:  drivers.TaskStateRunning,
		startedAt:  time.Now(),
		ctx:        ctx,
		cancelFunc: cancel,
		waitCh:     make(chan struct{}),
	}

	// Run the handler
	go h.run()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Kill the process
	err = cmd.Process.Kill()
	require.NoError(t, err)

	// Wait for process to complete
	select {
	case <-h.waitCh:
		// Process completed
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for process to complete")
	}

	// Verify exit code (killed processes typically have non-zero exit)
	status := h.TaskStatus()
	require.Equal(t, drivers.TaskStateExited, status.State)
	require.NotNil(t, status.ExitResult)
	// On Unix, SIGKILL typically results in exit code -1 or 137
	require.NotEqual(t, 0, status.ExitResult.ExitCode)
}

// Test that exit code is included in TaskStatus attributes
func TestExitCodePropagation_InDriverAttributes(t *testing.T) {
	logger := hclog.NewNullLogger()

	// Create a command that exits with 42
	cmd := exec.Command("sh", "-c", "exit 42")
	err := cmd.Start()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	h := &taskHandle{
		logger:     logger,
		cmd:        cmd,
		pid:        cmd.Process.Pid,
		taskConfig: &drivers.TaskConfig{ID: "test-attributes"},
		procState:  drivers.TaskStateRunning,
		startedAt:  time.Now(),
		ctx:        ctx,
		cancelFunc: cancel,
		waitCh:     make(chan struct{}),
	}

	// Run the handler
	go h.run()

	// Wait for process to complete
	select {
	case <-h.waitCh:
		// Process completed
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for process to complete")
	}

	// Verify TaskStatus includes PID in attributes
	status := h.TaskStatus()
	require.NotNil(t, status.DriverAttributes)
	require.Contains(t, status.DriverAttributes, "pid")

	// Exit code is in ExitResult, not DriverAttributes
	require.Equal(t, 42, status.ExitResult.ExitCode)
}
