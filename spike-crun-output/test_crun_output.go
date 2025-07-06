package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Test 1: Simple command execution with combined output
func testSimpleExecution() {
	fmt.Println("=== Test 1: Simple execution with CombinedOutput ===")
	
	cmd := exec.Command("echo", "Hello from crun test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	fmt.Printf("Combined output: %s\n", output)
}

// Test 2: Separate stdout and stderr capture
func testSeparateStreams() {
	fmt.Println("\n=== Test 2: Separate stdout/stderr capture ===")
	
	// Test with a command that writes to both stdout and stderr
	cmd := exec.Command("sh", "-c", "echo 'stdout message'; echo 'stderr message' >&2")
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	fmt.Printf("Stdout: %s", stdout.String())
	fmt.Printf("Stderr: %s", stderr.String())
}

// Test 3: Real-time streaming output
func testStreamingOutput() {
	fmt.Println("\n=== Test 3: Real-time streaming output ===")
	
	// Command that produces output over time
	cmd := exec.Command("sh", "-c", "for i in 1 2 3 4 5; do echo \"Line $i\"; sleep 0.5; done")
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating stdout pipe: %v", err)
		return
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error creating stderr pipe: %v", err)
		return
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %v", err)
		return
	}
	
	// Read stdout in real-time
	var wg sync.WaitGroup
	wg.Add(2)
	
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("[STDOUT] %s\n", scanner.Text())
		}
	}()
	
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[STDERR] %s\n", scanner.Text())
		}
	}()
	
	wg.Wait()
	
	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command: %v", err)
	}
}

// Test 4: Writing to files (similar to Nomad's approach)
func testFileOutput() {
	fmt.Println("\n=== Test 4: Writing output to files ===")
	
	// Create temporary files for stdout and stderr
	stdoutFile, err := os.CreateTemp("", "stdout-*.log")
	if err != nil {
		log.Printf("Error creating stdout file: %v", err)
		return
	}
	defer os.Remove(stdoutFile.Name())
	defer stdoutFile.Close()
	
	stderrFile, err := os.CreateTemp("", "stderr-*.log")
	if err != nil {
		log.Printf("Error creating stderr file: %v", err)
		return
	}
	defer os.Remove(stderrFile.Name())
	defer stderrFile.Close()
	
	cmd := exec.Command("sh", "-c", "echo 'stdout line 1'; echo 'stderr line 1' >&2; echo 'stdout line 2'")
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	
	if err := cmd.Run(); err != nil {
		log.Printf("Error running command: %v", err)
		return
	}
	
	// Read back the files
	stdoutFile.Seek(0, 0)
	stderrFile.Seek(0, 0)
	
	stdoutContent, _ := io.ReadAll(stdoutFile)
	stderrContent, _ := io.ReadAll(stderrFile)
	
	fmt.Printf("Stdout file content:\n%s", stdoutContent)
	fmt.Printf("Stderr file content:\n%s", stderrContent)
}

// Test 5: Streaming to both console and files
func testTeeOutput() {
	fmt.Println("\n=== Test 5: Tee output (stream to console and files) ===")
	
	stdoutFile, err := os.CreateTemp("", "stdout-tee-*.log")
	if err != nil {
		log.Printf("Error creating stdout file: %v", err)
		return
	}
	defer os.Remove(stdoutFile.Name())
	defer stdoutFile.Close()
	
	cmd := exec.Command("sh", "-c", "for i in 1 2 3; do echo \"Progress: $i/3\"; sleep 0.3; done")
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating stdout pipe: %v", err)
		return
	}
	
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %v", err)
		return
	}
	
	// Create a TeeReader to write to both console and file
	tee := io.TeeReader(stdout, stdoutFile)
	scanner := bufio.NewScanner(tee)
	
	for scanner.Scan() {
		fmt.Printf("[LIVE] %s\n", scanner.Text())
	}
	
	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command: %v", err)
	}
	
	fmt.Printf("Output also saved to: %s\n", stdoutFile.Name())
}

// Test 6: Simulating crun with a simple container
func testCrunSimulation() {
	fmt.Println("\n=== Test 6: Simulating crun-like behavior ===")
	
	// Check if crun is available
	if _, err := exec.LookPath("crun"); err != nil {
		fmt.Println("crun not found, using simulation with regular commands")
		
		// Simulate a container-like process
		cmd := exec.Command("sh", "-c", `
			echo "[Container] Starting process..."
			echo "[Container] Running application" >&2
			for i in 1 2 3; do
				echo "[Container] Processing item $i"
				sleep 0.2
			done
			echo "[Container] Process complete"
		`)
		
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		
		startTime := time.Now()
		err := cmd.Run()
		duration := time.Since(startTime)
		
		if err != nil {
			log.Printf("Error: %v", err)
		}
		
		fmt.Printf("Execution time: %v\n", duration)
		fmt.Printf("Stdout:\n%s", stdout.String())
		fmt.Printf("Stderr:\n%s", stderr.String())
		return
	}
	
	fmt.Println("crun is available - would test with actual container here")
}

func main() {
	fmt.Println("Testing stdout/stderr capture behavior for crun integration")
	fmt.Println("============================================================")
	
	testSimpleExecution()
	testSeparateStreams()
	testStreamingOutput()
	testFileOutput()
	testTeeOutput()
	testCrunSimulation()
	
	fmt.Println("\n=== Summary ===")
	fmt.Println("1. CombinedOutput() - Simple but no stream separation")
	fmt.Println("2. Separate buffers - Good for post-processing")
	fmt.Println("3. Pipe + Scanner - Best for real-time streaming")
	fmt.Println("4. File redirection - What Nomad executor uses")
	fmt.Println("5. TeeReader - Stream to multiple destinations")
	fmt.Println("6. All methods work with subprocess execution")
}