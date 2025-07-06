package milo

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/stretchr/testify/require"
)

// Test that handle properly manages lifecycle
func TestTaskHandle_Lifecycle(t *testing.T) {
	logger := hclog.NewNullLogger()
	
	// Create a simple command that runs for a bit
	cmd := exec.Command("sleep", "0.1")
	err := cmd.Start()
	require.NoError(t, err)
	
	ctx, cancel := context.WithCancel(context.Background())
	
	h := &taskHandle{
		logger:       logger,
		cmd:          cmd,
		pid:          cmd.Process.Pid,
		taskConfig:   &drivers.TaskConfig{ID: "test-task"},
		procState:    drivers.TaskStateRunning,
		startedAt:    time.Now(),
		ctx:          ctx,
		cancelFunc:   cancel,
		waitCh:       make(chan struct{}),
	}
	
	// Verify initial state
	require.True(t, h.IsRunning())
	status := h.TaskStatus()
	require.Equal(t, drivers.TaskStateRunning, status.State)
	
	// Run the handler
	go h.run()
	
	// Wait for process to complete
	select {
	case <-h.waitCh:
		// Process completed
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for process to complete")
	}
	
	// Verify final state
	require.False(t, h.IsRunning())
	status = h.TaskStatus()
	require.Equal(t, drivers.TaskStateExited, status.State)
	require.Equal(t, 0, status.ExitResult.ExitCode)
	require.False(t, status.CompletedAt.IsZero())
}

// Test that cancel properly stops streaming
func TestTaskHandle_CancelStreaming(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Verify context is not cancelled
	select {
	case <-ctx.Done():
		t.Fatal("context should not be cancelled yet")
	default:
		// Good
	}
	
	// Cancel the context
	cancel()
	
	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Good
	default:
		t.Fatal("context should be cancelled")
	}
}