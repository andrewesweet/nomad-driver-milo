package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

// LogStreamer represents a component that can stream logs from crun
type LogStreamer struct {
	stdout io.ReadCloser
	stderr io.ReadCloser
	cmd    *exec.Cmd
	wg     sync.WaitGroup
}

// NewLogStreamer creates a new log streamer for a crun command
func NewLogStreamer(crunArgs []string) (*LogStreamer, error) {
	cmd := exec.Command(crunArgs[0], crunArgs[1:]...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	return &LogStreamer{
		stdout: stdout,
		stderr: stderr,
		cmd:    cmd,
	}, nil
}

// Start begins the crun process and starts streaming logs
func (ls *LogStreamer) Start(stdoutHandler, stderrHandler func(string)) error {
	if err := ls.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start crun: %w", err)
	}
	
	// Stream stdout
	ls.wg.Add(1)
	go func() {
		defer ls.wg.Done()
		scanner := bufio.NewScanner(ls.stdout)
		for scanner.Scan() {
			stdoutHandler(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading stdout: %v", err)
		}
	}()
	
	// Stream stderr
	ls.wg.Add(1)
	go func() {
		defer ls.wg.Done()
		scanner := bufio.NewScanner(ls.stderr)
		for scanner.Scan() {
			stderrHandler(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading stderr: %v", err)
		}
	}()
	
	return nil
}

// Wait waits for the process to complete
func (ls *LogStreamer) Wait() error {
	ls.wg.Wait()
	return ls.cmd.Wait()
}

// FileAndStreamLogger writes logs to both files and provides callbacks for streaming
type FileAndStreamLogger struct {
	stdoutFile *os.File
	stderrFile *os.File
	onStdout   func(string)
	onStderr   func(string)
}

// NewFileAndStreamLogger creates a logger that writes to files and calls callbacks
func NewFileAndStreamLogger(stdoutPath, stderrPath string, onStdout, onStderr func(string)) (*FileAndStreamLogger, error) {
	stdoutFile, err := os.Create(stdoutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout file: %w", err)
	}
	
	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		stdoutFile.Close()
		return nil, fmt.Errorf("failed to create stderr file: %w", err)
	}
	
	return &FileAndStreamLogger{
		stdoutFile: stdoutFile,
		stderrFile: stderrFile,
		onStdout:   onStdout,
		onStderr:   onStderr,
	}, nil
}

// HandleStdout processes stdout lines
func (fsl *FileAndStreamLogger) HandleStdout(line string) {
	fmt.Fprintf(fsl.stdoutFile, "%s\n", line)
	fsl.stdoutFile.Sync() // Ensure data is written immediately
	if fsl.onStdout != nil {
		fsl.onStdout(line)
	}
}

// HandleStderr processes stderr lines
func (fsl *FileAndStreamLogger) HandleStderr(line string) {
	fmt.Fprintf(fsl.stderrFile, "%s\n", line)
	fsl.stderrFile.Sync() // Ensure data is written immediately
	if fsl.onStderr != nil {
		fsl.onStderr(line)
	}
}

// Close closes the log files
func (fsl *FileAndStreamLogger) Close() {
	fsl.stdoutFile.Close()
	fsl.stderrFile.Close()
}

// Demo function showing integration pattern
func demonstrateCrunIntegration() {
	fmt.Println("=== Demonstrating crun integration pattern ===\n")
	
	// Simulate crun command (replace with actual crun command)
	crunCmd := []string{"sh", "-c", `
		echo "[2024-01-01 12:00:00] Container starting..."
		sleep 1
		echo "[2024-01-01 12:00:01] Loading application..." >&2
		sleep 1
		echo "[2024-01-01 12:00:02] Application started"
		for i in 1 2 3; do
			echo "[2024-01-01 12:00:0$((i+2))] Processing request $i"
			sleep 0.5
		done
		echo "[2024-01-01 12:00:06] Application shutting down"
	`}
	
	// Create log files
	stdoutPath := "/tmp/milo-task-stdout.log"
	stderrPath := "/tmp/milo-task-stderr.log"
	
	// Create file and stream logger
	logger, err := NewFileAndStreamLogger(
		stdoutPath,
		stderrPath,
		func(line string) {
			fmt.Printf("[STDOUT] %s\n", line)
		},
		func(line string) {
			fmt.Printf("[STDERR] %s\n", line)
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close()
	
	// Create log streamer
	streamer, err := NewLogStreamer(crunCmd)
	if err != nil {
		log.Fatal(err)
	}
	
	// Start streaming
	fmt.Println("Starting process and streaming logs...")
	startTime := time.Now()
	
	if err := streamer.Start(logger.HandleStdout, logger.HandleStderr); err != nil {
		log.Fatal(err)
	}
	
	// Wait for completion
	if err := streamer.Wait(); err != nil {
		log.Printf("Process exited with error: %v", err)
	}
	
	duration := time.Since(startTime)
	fmt.Printf("\nProcess completed in %v\n", duration)
	fmt.Printf("Logs saved to:\n  stdout: %s\n  stderr: %s\n", stdoutPath, stderrPath)
}

// Example of how this would integrate with Nomad's task handle
type TaskLogManager struct {
	streamer *LogStreamer
	logger   *FileAndStreamLogger
	taskID   string
}

func (tlm *TaskLogManager) StartLogging(crunCmd []string, stdoutPath, stderrPath string) error {
	// This would be called from the driver's StartTask method
	
	// Setup file and stream logger
	logger, err := NewFileAndStreamLogger(
		stdoutPath,
		stderrPath,
		func(line string) {
			// This could send to Nomad's log streaming API
			fmt.Printf("[Task %s] stdout: %s\n", tlm.taskID, line)
		},
		func(line string) {
			// This could send to Nomad's log streaming API
			fmt.Printf("[Task %s] stderr: %s\n", tlm.taskID, line)
		},
	)
	if err != nil {
		return err
	}
	tlm.logger = logger
	
	// Create and start log streamer
	streamer, err := NewLogStreamer(crunCmd)
	if err != nil {
		logger.Close()
		return err
	}
	tlm.streamer = streamer
	
	return streamer.Start(logger.HandleStdout, logger.HandleStderr)
}

func main() {
	fmt.Println("Testing crun output integration for Milo driver")
	fmt.Println("==============================================\n")
	
	demonstrateCrunIntegration()
	
	fmt.Println("\n=== Implementation Notes for Milo Driver ===")
	fmt.Println("1. Replace executor.ExecCommand with direct exec.Command for more control")
	fmt.Println("2. Use StdoutPipe() and StderrPipe() for real-time streaming")
	fmt.Println("3. Write to files while simultaneously streaming for Nomad API")
	fmt.Println("4. Use buffered scanner for line-by-line processing")
	fmt.Println("5. Implement proper cleanup in Stop/Destroy methods")
	fmt.Println("6. Consider buffering and rate limiting for high-volume logs")
}