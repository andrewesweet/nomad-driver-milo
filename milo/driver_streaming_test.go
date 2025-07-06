package milo

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/stretchr/testify/require"
)

// TestDriverStreamingSimple tests that the driver captures output with the new streaming approach
func TestDriverStreamingSimple(t *testing.T) {
	t.Skip("Requires crun and Java - enable for manual testing")
	
	require := require.New(t)
	logger := hclog.NewNullLogger()
	
	// Create driver
	d := NewPlugin(logger).(*MiloDriverPlugin)
	d.config = &Config{}
	d.nomadConfig = &base.ClientDriverConfig{
		ClientMaxPort: 10000,
		ClientMinPort: 9000,
	}
	
	// Create test structure
	allocDir := t.TempDir()
	taskName := "test-streaming"
	taskDir := filepath.Join(allocDir, taskName)
	localDir := filepath.Join(taskDir, "local")
	require.NoError(os.MkdirAll(localDir, 0755))
	
	// Copy test JAR
	testJarSrc := filepath.Join("..", "test-artifacts", "hello-world.jar")
	testJarDst := filepath.Join(localDir, "hello-world.jar")
	
	// Check if test JAR exists
	if _, err := os.Stat(testJarSrc); err == nil {
		data, err := os.ReadFile(testJarSrc)
		require.NoError(err)
		require.NoError(os.WriteFile(testJarDst, data, 0644))
	} else {
		t.Skip("Test JAR not found, skipping")
	}
	
	// Create FIFOs
	stdoutPath := filepath.Join(taskDir, "stdout.fifo")
	stderrPath := filepath.Join(taskDir, "stderr.fifo")
	require.NoError(exec.Command("mkfifo", stdoutPath).Run())
	require.NoError(exec.Command("mkfifo", stderrPath).Run())
	
	// Prepare task config
	taskCfg := &drivers.TaskConfig{
		ID:         "test-task-streaming",
		Name:       taskName,
		AllocDir:   allocDir,
		StdoutPath: stdoutPath,
		StderrPath: stderrPath,
	}
	
	// Encode driver config
	driverConfig := map[string]interface{}{
		"dummy": "",
	}
	require.NoError(taskCfg.EncodeConcreteDriverConfig(&driverConfig))
	
	// Channel to capture output
	outputCh := make(chan string, 1)
	errCh := make(chan error, 1)
	
	// Start reader before starting task
	go func() {
		fifo, err := os.Open(stdoutPath)
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
	
	// Start the task
	handle, _, err := d.StartTask(taskCfg)
	require.NoError(err)
	require.NotNil(handle)
	
	// Wait for output
	select {
	case output := <-outputCh:
		t.Logf("Captured output: %s", output)
		require.Contains(output, "Hello", "Expected 'Hello' in output")
	case err := <-errCh:
		t.Fatalf("Error reading output: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for output")
	}
	
	// Clean up
	err = d.StopTask(taskCfg.ID, 1*time.Second, "SIGTERM")
	require.NoError(err)
	
	err = d.DestroyTask(taskCfg.ID, false)
	require.NoError(err)
}