// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package milo

import (
	"context"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/nomad/drivers/shared/executor"
	"github.com/hashicorp/nomad/plugins/drivers"
)

// taskHandle should store all relevant runtime information
// such as process ID if this is a local task or other meta
// data if this driver deals with external APIs
type taskHandle struct {
	// stateLock syncs access to all fields below
	stateLock sync.RWMutex

	logger       hclog.Logger
	cmd          *exec.Cmd           // Direct command execution
	taskConfig   *drivers.TaskConfig
	procState    drivers.TaskState
	startedAt    time.Time
	completedAt  time.Time
	exitResult   *drivers.ExitResult

	// Process management
	pid        int
	ctx        context.Context
	cancelFunc context.CancelFunc
	waitCh     chan struct{}

	// Log streaming
	stdoutStream *LogStreamer
	stderrStream *LogStreamer

	// Legacy fields for compatibility (may be removed later)
	exec         executor.Executor
	pluginClient *plugin.Client
}

func (h *taskHandle) TaskStatus() *drivers.TaskStatus {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()

	return &drivers.TaskStatus{
		ID:          h.taskConfig.ID,
		Name:        h.taskConfig.Name,
		State:       h.procState,
		StartedAt:   h.startedAt,
		CompletedAt: h.completedAt,
		ExitResult:  h.exitResult,
		DriverAttributes: map[string]string{
			"pid": strconv.Itoa(h.pid),
		},
	}
}

func (h *taskHandle) IsRunning() bool {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()
	return h.procState == drivers.TaskStateRunning
}

func (h *taskHandle) run() {
	h.stateLock.Lock()
	if h.exitResult == nil {
		h.exitResult = &drivers.ExitResult{}
	}
	h.stateLock.Unlock()

	// Wait for the command to complete
	err := h.cmd.Wait()
	
	// Signal that process has finished
	close(h.waitCh)
	
	// Cancel log streaming context
	if h.cancelFunc != nil {
		h.cancelFunc()
	}

	h.stateLock.Lock()
	defer h.stateLock.Unlock()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			h.exitResult.ExitCode = exitErr.ExitCode()
			h.procState = drivers.TaskStateExited
		} else {
			h.exitResult.Err = err
			h.procState = drivers.TaskStateUnknown
		}
	} else {
		h.procState = drivers.TaskStateExited
		h.exitResult.ExitCode = 0
	}
	
	h.completedAt = time.Now()
}
