package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
)

// MiloFIFOStreamer demonstrates how to integrate FIFO streaming into the Milo driver
type MiloFIFOStreamer struct {
	taskID     string
	stdoutPath string
	stderrPath string
	cmd        *exec.Cmd
	logger     hclog.Logger
	
	// FIFO files for writing
	stdoutFIFO *os.File
	stderrFIFO *os.File
	
	// Streaming state
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewMiloFIFOStreamer(taskID, stdoutPath, stderrPath string, cmd *exec.Cmd, logger hclog.Logger) *MiloFIFOStreamer {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &MiloFIFOStreamer{
		taskID:     taskID,
		stdoutPath: stdoutPath,
		stderrPath: stderrPath,
		cmd:        cmd,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// StartStreaming begins streaming logs from the command to FIFOs
func (mfs *MiloFIFOStreamer) StartStreaming() error {
	mfs.logger.Info("starting FIFO streaming", "task_id", mfs.taskID)

	// Open FIFOs for writing
	var err error
	mfs.stdoutFIFO, err = os.OpenFile(mfs.stdoutPath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open stdout FIFO: %v", err)
	}

	mfs.stderrFIFO, err = os.OpenFile(mfs.stderrPath, os.O_WRONLY, 0)
	if err != nil {
		mfs.stdoutFIFO.Close()
		return fmt.Errorf("failed to open stderr FIFO: %v", err)
	}

	mfs.logger.Info("opened FIFOs for streaming", "stdout", mfs.stdoutPath, "stderr", mfs.stderrPath)

	// Get pipes from the command
	stdout, err := mfs.cmd.StdoutPipe()
	if err != nil {
		mfs.cleanup()
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := mfs.cmd.StderrPipe()
	if err != nil {
		mfs.cleanup()
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	// Start the command
	if err := mfs.cmd.Start(); err != nil {
		mfs.cleanup()
		return fmt.Errorf("failed to start command: %v", err)
	}

	mfs.logger.Info("started command for streaming", "pid", mfs.cmd.Process.Pid)

	// Start streaming goroutines
	mfs.wg.Add(2)
	go mfs.streamPipeToFIFO(stdout, mfs.stdoutFIFO, "stdout")
	go mfs.streamPipeToFIFO(stderr, mfs.stderrFIFO, "stderr")

	return nil
}

// streamPipeToFIFO streams data from a pipe to a FIFO with proper error handling
func (mfs *MiloFIFOStreamer) streamPipeToFIFO(pipe io.ReadCloser, fifo *os.File, streamType string) {
	defer mfs.wg.Done()
	defer pipe.Close()

	mfs.logger.Debug("started streaming", "stream", streamType)

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		select {
		case <-mfs.ctx.Done():
			mfs.logger.Debug("streaming cancelled", "stream", streamType)
			return
		default:
		}

		line := scanner.Text()
		
		// Add timestamp to log line as driver would
		timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
		logLine := fmt.Sprintf("%s %s\n", timestamp, line)

		// Write to FIFO
		if _, err := fifo.WriteString(logLine); err != nil {
			mfs.logger.Error("failed to write to FIFO", "stream", streamType, "error", err)
			return
		}

		// Force flush to ensure immediate delivery
		if err := fifo.Sync(); err != nil {
			mfs.logger.Error("failed to sync FIFO", "stream", streamType, "error", err)
		}

		mfs.logger.Debug("streamed log line", "stream", streamType, "line", line)
	}

	if err := scanner.Err(); err != nil {
		mfs.logger.Error("error reading from pipe", "stream", streamType, "error", err)
	}

	mfs.logger.Debug("finished streaming", "stream", streamType)
}

// Wait waits for the command to complete and returns the exit result
func (mfs *MiloFIFOStreamer) Wait() (*drivers.ExitResult, error) {
	err := mfs.cmd.Wait()
	mfs.wg.Wait() // Wait for streaming to complete

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return &drivers.ExitResult{
				ExitCode: exitError.ExitCode(),
				Signal:   0,
			}, nil
		}
		return &drivers.ExitResult{
			Err: err,
		}, err
	}

	return &drivers.ExitResult{
		ExitCode: 0,
		Signal:   0,
	}, nil
}

// Stop stops the streaming and cleans up resources
func (mfs *MiloFIFOStreamer) Stop() {
	mfs.logger.Info("stopping FIFO streaming", "task_id", mfs.taskID)
	
	mfs.cancel()
	
	if mfs.cmd != nil && mfs.cmd.Process != nil {
		mfs.cmd.Process.Kill()
	}
	
	mfs.wg.Wait()
	mfs.cleanup()
}

func (mfs *MiloFIFOStreamer) cleanup() {
	if mfs.stdoutFIFO != nil {
		mfs.stdoutFIFO.Close()
	}
	if mfs.stderrFIFO != nil {
		mfs.stderrFIFO.Close()
	}
}

// TestMiloIntegration demonstrates how this would integrate into the Milo driver
func TestMiloIntegration() error {
	fmt.Println("=== Milo Driver FIFO Integration Test ===")
	
	// Create test directory structure
	tempDir, err := os.MkdirTemp("", "milo-fifo-test-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create FIFOs as Nomad would
	stdoutPath := filepath.Join(tempDir, "stdout.fifo")
	stderrPath := filepath.Join(tempDir, "stderr.fifo")

	if err := syscall.Mkfifo(stdoutPath, 0666); err != nil {
		return fmt.Errorf("failed to create stdout FIFO: %v", err)
	}
	if err := syscall.Mkfifo(stderrPath, 0666); err != nil {
		return fmt.Errorf("failed to create stderr FIFO: %v", err)
	}

	fmt.Printf("Created FIFOs:\n")
	fmt.Printf("  Stdout: %s\n", stdoutPath)
	fmt.Printf("  Stderr: %s\n", stderrPath)

	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "milo-fifo-test",
		Level: hclog.Debug,
	})

	// Create a test command that produces both stdout and stderr
	cmd := exec.Command("sh", "-c", `
		echo "Starting Java application..."
		echo "Loading configuration..." >&2
		echo "Application started successfully"
		echo "Processing request 1"
		echo "Warning: Memory usage high" >&2
		echo "Processing request 2"
		echo "Finished processing"
	`)

	// Create the FIFO streamer
	streamer := NewMiloFIFOStreamer("test-task", stdoutPath, stderrPath, cmd, logger)

	// Start Nomad readers in background (simulate Nomad reading from FIFOs)
	var readerWG sync.WaitGroup
	readerWG.Add(2)

	// Stdout reader
	go func() {
		defer readerWG.Done()
		file, err := os.OpenFile(stdoutPath, os.O_RDONLY, 0)
		if err != nil {
			fmt.Printf("Failed to open stdout FIFO for reading: %v\n", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fmt.Printf("[NOMAD-STDOUT] %s\n", scanner.Text())
		}
	}()

	// Stderr reader  
	go func() {
		defer readerWG.Done()
		file, err := os.OpenFile(stderrPath, os.O_RDONLY, 0)
		if err != nil {
			fmt.Printf("Failed to open stderr FIFO for reading: %v\n", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fmt.Printf("[NOMAD-STDERR] %s\n", scanner.Text())
		}
	}()

	// Give readers time to start
	time.Sleep(100 * time.Millisecond)

	// Start streaming
	if err := streamer.StartStreaming(); err != nil {
		return fmt.Errorf("failed to start streaming: %v", err)
	}

	// Wait for command to complete
	exitResult, err := streamer.Wait()
	if err != nil {
		return fmt.Errorf("command execution failed: %v", err)
	}

	fmt.Printf("Command completed with exit code: %d\n", exitResult.ExitCode)

	// Close FIFOs to signal readers
	streamer.cleanup()

	// Wait for readers to finish
	readerWG.Wait()

	fmt.Println("✓ Milo FIFO integration test completed successfully")
	return nil
}

// DemonstrateIntegrationWithCurrentDriver shows how to modify the current driver
func DemonstrateIntegrationWithCurrentDriver() {
	fmt.Println("=== Integration with Current Milo Driver ===")
	fmt.Println()

	fmt.Println("Current driver.go StartTask method uses:")
	fmt.Println("```go")
	fmt.Println("execCmd := &executor.ExecCommand{")
	fmt.Println("    Cmd:        crunCmd[0],")
	fmt.Println("    Args:       crunCmd[1:],")
	fmt.Println("    StdoutPath: cfg.StdoutPath,  // <- FIFO path from Nomad")
	fmt.Println("    StderrPath: cfg.StderrPath,  // <- FIFO path from Nomad")
	fmt.Println("}")
	fmt.Println("```")
	fmt.Println()

	fmt.Println("Recommended modification for real-time streaming:")
	fmt.Println("```go")
	fmt.Println("// Instead of using executor.ExecCommand, use direct exec.Command")
	fmt.Println("cmd := exec.Command(crunCmd[0], crunCmd[1:]...)")
	fmt.Println()
	fmt.Println("// Create FIFO streamer")
	fmt.Println("streamer := NewMiloFIFOStreamer(cfg.ID, cfg.StdoutPath, cfg.StderrPath, cmd, d.logger)")
	fmt.Println()
	fmt.Println("// Start streaming")
	fmt.Println("if err := streamer.StartStreaming(); err != nil {")
	fmt.Println("    return nil, nil, fmt.Errorf(\"failed to start log streaming: %v\", err)")
	fmt.Println("}")
	fmt.Println()
	fmt.Println("// Store streamer in task handle")
	fmt.Println("h := &taskHandle{")
	fmt.Println("    // ... existing fields ...")
	fmt.Println("    fifoStreamer: streamer,")
	fmt.Println("}")
	fmt.Println("```")
	fmt.Println()

	fmt.Println("Task handle run() method modification:")
	fmt.Println("```go")
	fmt.Println("func (h *taskHandle) run() {")
	fmt.Println("    // Wait for command completion with streaming")
	fmt.Println("    exitResult, err := h.fifoStreamer.Wait()")
	fmt.Println("    ")
	fmt.Println("    h.stateLock.Lock()")
	fmt.Println("    defer h.stateLock.Unlock()")
	fmt.Println("    ")
	fmt.Println("    if err != nil {")
	fmt.Println("        h.exitResult.Err = err")
	fmt.Println("        h.procState = drivers.TaskStateUnknown")
	fmt.Println("    } else {")
	fmt.Println("        h.exitResult = exitResult")
	fmt.Println("        h.procState = drivers.TaskStateExited")
	fmt.Println("    }")
	fmt.Println("    h.completedAt = time.Now()")
	fmt.Println("}")
	fmt.Println("```")
	fmt.Println()

	fmt.Println("Key benefits:")
	fmt.Println("1. Real-time log streaming to Nomad")
	fmt.Println("2. Proper error handling for broken pipes")
	fmt.Println("3. Clean integration with existing driver structure")
	fmt.Println("4. No dependency on external OCI libraries")
	fmt.Println("5. Compatible with Nomad's FIFO-based logging system")
}

func main() {
	fmt.Println("Milo Driver FIFO Integration Spike")
	fmt.Println("==================================")
	fmt.Println()

	if err := TestMiloIntegration(); err != nil {
		fmt.Printf("Test failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	DemonstrateIntegrationWithCurrentDriver()

	fmt.Println()
	fmt.Println("✓ Spike 3 completed successfully!")
	fmt.Println("Ready to implement Epic 001 User Story 002 - Nomad Log Streaming Integration")
}