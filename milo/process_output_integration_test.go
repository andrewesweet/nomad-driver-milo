package milo

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

// TestLogStreamerIntegration tests the integration between exec.Command and LogStreamer
func TestLogStreamerIntegration_CapturesStdout(t *testing.T) {
	require := require.New(t)
	logger := hclog.NewNullLogger()

	// Create temporary directory
	tmpDir := t.TempDir()
	
	// Create FIFO
	fifoPath := filepath.Join(tmpDir, "stdout.fifo")
	require.NoError(exec.Command("mkfifo", fifoPath).Run())

	// Create a simple command that outputs to stdout
	cmd := exec.Command("echo", "Hello from stdout")
	
	// Get stdout pipe
	stdoutPipe, err := cmd.StdoutPipe()
	require.NoError(err)

	// Start the command
	require.NoError(cmd.Start())

	// Create context for streaming
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create log streamer
	streamer := NewLogStreamer(logger, fifoPath, stdoutPipe)

	// Channel to capture output
	outputCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Start streaming in goroutine
	go func() {
		if err := streamer.Stream(ctx); err != nil {
			errCh <- err
		}
	}()

	// Read from FIFO in another goroutine
	go func() {
		fifo, err := os.Open(fifoPath)
		if err != nil {
			errCh <- err
			return
		}
		defer fifo.Close()

		data, err := io.ReadAll(fifo)
		if err != nil {
			errCh <- err
			return
		}
		outputCh <- string(data)
	}()

	// Get output or error (before waiting for command)
	select {
	case output := <-outputCh:
		require.Equal("Hello from stdout\n", output)
	case err := <-errCh:
		t.Fatalf("Streaming error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for output")
	}

	// Now wait for command to complete
	require.NoError(cmd.Wait())
}

// TestLogStreamerIntegration tests stderr capture
func TestLogStreamerIntegration_CapturesStderr(t *testing.T) {
	require := require.New(t)
	logger := hclog.NewNullLogger()

	// Create temporary directory
	tmpDir := t.TempDir()
	
	// Create FIFO
	fifoPath := filepath.Join(tmpDir, "stderr.fifo")
	require.NoError(exec.Command("mkfifo", fifoPath).Run())

	// Create a command that outputs to stderr
	cmd := exec.Command("sh", "-c", "echo 'Error message' >&2")
	
	// Get stderr pipe
	stderrPipe, err := cmd.StderrPipe()
	require.NoError(err)

	// Start the command
	require.NoError(cmd.Start())

	// Create context for streaming
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create log streamer
	streamer := NewLogStreamer(logger, fifoPath, stderrPipe)

	// Channel to capture output
	outputCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Start streaming in goroutine
	go func() {
		if err := streamer.Stream(ctx); err != nil {
			errCh <- err
		}
	}()

	// Read from FIFO in another goroutine
	go func() {
		fifo, err := os.Open(fifoPath)
		if err != nil {
			errCh <- err
			return
		}
		defer fifo.Close()

		data, err := io.ReadAll(fifo)
		if err != nil {
			errCh <- err
			return
		}
		outputCh <- string(data)
	}()

	// Get output or error (before waiting for command)
	select {
	case output := <-outputCh:
		require.Equal("Error message\n", output)
	case err := <-errCh:
		t.Fatalf("Streaming error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for output")
	}

	// Now wait for command to complete
	require.NoError(cmd.Wait())
}

// TestLogStreamerIntegration tests concurrent stdout/stderr streaming
func TestLogStreamerIntegration_ConcurrentStreaming(t *testing.T) {
	require := require.New(t)
	logger := hclog.NewNullLogger()

	// Create temporary directory
	tmpDir := t.TempDir()
	
	// Create FIFOs
	stdoutFifo := filepath.Join(tmpDir, "stdout.fifo")
	stderrFifo := filepath.Join(tmpDir, "stderr.fifo")
	require.NoError(exec.Command("mkfifo", stdoutFifo).Run())
	require.NoError(exec.Command("mkfifo", stderrFifo).Run())

	// Create a command that outputs to both
	cmd := exec.Command("sh", "-c", "echo 'stdout line'; echo 'stderr line' >&2")
	
	// Get pipes
	stdoutPipe, err := cmd.StdoutPipe()
	require.NoError(err)
	stderrPipe, err := cmd.StderrPipe()
	require.NoError(err)

	// Start the command
	require.NoError(cmd.Start())

	// Create context for streaming
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create log streamers
	stdoutStreamer := NewLogStreamer(logger.Named("stdout"), stdoutFifo, stdoutPipe)
	stderrStreamer := NewLogStreamer(logger.Named("stderr"), stderrFifo, stderrPipe)

	// Channels to capture output
	stdoutCh := make(chan string, 1)
	stderrCh := make(chan string, 1)
	errCh := make(chan error, 2)

	// Start streaming goroutines
	go func() {
		if err := stdoutStreamer.Stream(ctx); err != nil {
			errCh <- err
		}
	}()

	go func() {
		if err := stderrStreamer.Stream(ctx); err != nil {
			errCh <- err
		}
	}()

	// Read from FIFOs
	go func() {
		fifo, err := os.Open(stdoutFifo)
		if err != nil {
			errCh <- err
			return
		}
		defer fifo.Close()

		data, err := io.ReadAll(fifo)
		if err != nil {
			errCh <- err
			return
		}
		stdoutCh <- string(data)
	}()

	go func() {
		fifo, err := os.Open(stderrFifo)
		if err != nil {
			errCh <- err
			return
		}
		defer fifo.Close()

		data, err := io.ReadAll(fifo)
		if err != nil {
			errCh <- err
			return
		}
		stderrCh <- string(data)
	}()

	// Collect outputs (before waiting for command)
	var stdout, stderr string
	timeout := time.After(2 * time.Second)

	for i := 0; i < 2; i++ {
		select {
		case out := <-stdoutCh:
			stdout = out
		case err := <-stderrCh:
			stderr = err
		case err := <-errCh:
			t.Fatalf("Streaming error: %v", err)
		case <-timeout:
			t.Fatal("Timeout waiting for outputs")
		}
	}

	// Verify outputs
	require.Equal("stdout line\n", stdout)
	require.Equal("stderr line\n", stderr)

	// Now wait for command to complete
	require.NoError(cmd.Wait())
}