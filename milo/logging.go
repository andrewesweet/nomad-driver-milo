package milo

import (
	"context"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/hashicorp/go-hclog"
)

// LogStreamer handles streaming from a reader to a FIFO
type LogStreamer struct {
	logger   hclog.Logger
	fifoPath string
	source   io.Reader
	writer   io.Writer // For testing, can be overridden
}

// NewLogStreamer creates a new log streamer
func NewLogStreamer(logger hclog.Logger, fifoPath string, source io.Reader) *LogStreamer {
	return &LogStreamer{
		logger:   logger,
		fifoPath: fifoPath,
		source:   source,
	}
}

// Stream starts streaming logs from source to FIFO
func (ls *LogStreamer) Stream(ctx context.Context) error {
	// Open FIFO for writing
	// This will block until a reader is attached
	fifo, err := os.OpenFile(ls.fifoPath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open FIFO %s: %w", ls.fifoPath, err)
	}
	defer fifo.Close()

	ls.writer = fifo
	return ls.streamToWriter(ctx)
}

// streamToWriter handles the actual streaming logic
func (ls *LogStreamer) streamToWriter(ctx context.Context) error {
	// Use io.Copy for efficient streaming
	// Default buffer size (32KB) is appropriate for most cases
	_, err := io.Copy(ls.writer, ls.source)

	if err != nil {
		// Check if it's a broken pipe error
		if isEPIPE(err) {
			// Log at debug level - this is normal when consumer disconnects
			ls.logger.Debug("log consumer disconnected", "error", err)
			// Don't return error for broken pipe - it's expected behavior
			return nil
		}
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	return nil
}

// isEPIPE checks if an error is a broken pipe error
func isEPIPE(err error) bool {
	if err == nil {
		return false
	}

	// Check for closed pipe error
	if err == io.ErrClosedPipe {
		return true
	}

	// Check for syscall EPIPE
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			return errno == syscall.EPIPE
		}
	}

	// Check if error message contains "broken pipe"
	return err.Error() == "write: broken pipe" || err.Error() == "broken pipe"
}
