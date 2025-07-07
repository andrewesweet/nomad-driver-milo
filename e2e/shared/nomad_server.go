package shared

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hashicorp/nomad/api"
)

// LiveNomadServer represents a live Nomad server instance for testing
type LiveNomadServer struct {
	cmd         *exec.Cmd
	ctx         context.Context
	cancel      context.CancelFunc
	httpPort    int
	rpcPort     int
	serfPort    int
	timeout     time.Duration
	configPath  string
	dataDir     string
	client      *api.Client
}

// NewLiveNomadServer creates a new LiveNomadServer instance
func NewLiveNomadServer() *LiveNomadServer {
	httpPort := getFreePort()
	rpcPort := getFreePort()
	serfPort := getFreePort()
	
	ctx, cancel := context.WithCancel(context.Background())
	
	server := &LiveNomadServer{
		ctx:      ctx,
		cancel:   cancel,
		httpPort: httpPort,
		rpcPort:  rpcPort,
		serfPort: serfPort,
		timeout:  30 * time.Second,
	}
	
	return server
}

// Start starts the Nomad server
func (s *LiveNomadServer) Start() error {
	// Ensure plugin symlink exists
	pluginDir := "/tmp/nomad-plugins"
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin dir: %v", err)
	}
	
	// Create symlink from nomad-driver-milo to milo if needed
	miloPath := filepath.Join(pluginDir, "milo")
	targetPath := filepath.Join(pluginDir, "nomad-driver-milo")
	if _, err := os.Lstat(targetPath); os.IsNotExist(err) {
		if err := os.Symlink(miloPath, targetPath); err != nil && !os.IsExist(err) {
			return fmt.Errorf("failed to create plugin symlink: %v", err)
		}
	}
	
	// Create temporary directories
	var err error
	s.dataDir, err = os.MkdirTemp("", "nomad-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	
	// Create config file
	if err := s.createConfig(); err != nil {
		return fmt.Errorf("failed to create config: %v", err)
	}
	
	// Create log file for Nomad output
	logFile, err := os.Create(filepath.Join(s.dataDir, "nomad.log"))
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	
	// Start Nomad server
	s.cmd = exec.CommandContext(s.ctx, "nomad", "agent", "-config", s.configPath)
	s.cmd.Env = append(os.Environ(), "NOMAD_DISABLE_UPDATE_CHECK=1")
	s.cmd.Stdout = logFile
	s.cmd.Stderr = logFile
	
	// Start the process
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start nomad: %v", err)
	}
	
	// Create API client
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("http://127.0.0.1:%d", s.httpPort)
	s.client, err = api.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	
	// Wait for server to be ready
	return s.waitForReady()
}

// Stop stops the Nomad server and cleans up resources
func (s *LiveNomadServer) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}
	
	// Cleanup temp directories
	if s.dataDir != "" {
		os.RemoveAll(s.dataDir)
	}
	if s.configPath != "" {
		os.Remove(s.configPath)
	}
	
	return nil
}

// IsRunning checks if the Nomad server process is running
func (s *LiveNomadServer) IsRunning() bool {
	if s.cmd == nil || s.cmd.Process == nil {
		return false
	}
	
	// Check if process is still running by sending signal 0
	err := s.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

// IsAccessible checks if the Nomad server API is accessible
func (s *LiveNomadServer) IsAccessible() bool {
	if s.client == nil {
		return false
	}
	
	// Try to connect to API
	_, err := s.client.Status().Leader()
	return err == nil
}

// GetLoadedPlugins returns a list of loaded driver plugins
func (s *LiveNomadServer) GetLoadedPlugins() []string {
	if s.client == nil {
		return nil
	}
	
	// Get node status to check plugins
	nodes, _, err := s.client.Nodes().List(nil)
	if err != nil || len(nodes) == 0 {
		return nil
	}
	
	node, _, err := s.client.Nodes().Info(nodes[0].ID, nil)
	if err != nil {
		return nil
	}
	
	var plugins []string
	if node.Drivers != nil {
		for driver := range node.Drivers {
			plugins = append(plugins, driver)
		}
	}
	
	return plugins
}

// GetHTTPPort returns the HTTP port the server is listening on
func (s *LiveNomadServer) GetHTTPPort() int {
	return s.httpPort
}

// SetPluginLoadTimeout sets the timeout for waiting for plugins to load
func (s *LiveNomadServer) SetPluginLoadTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// GetDataDir returns the data directory path
func (s *LiveNomadServer) GetDataDir() string {
	return s.dataDir
}

// GetClient returns the Nomad API client
func (s *LiveNomadServer) GetClient() *api.Client {
	return s.client
}

// SubmitJarJobFromGit submits a job that runs a JAR file from git repository using the Milo driver
func (s *LiveNomadServer) SubmitJarJobFromGit(jobID, gitRepoURL string) error {
	if s.client == nil {
		return fmt.Errorf("client not initialized")
	}
	
	// Create job specification
	job := &api.Job{
		ID:   &jobID,
		Name: &jobID,
		Type: stringToPtr("batch"),
		TaskGroups: []*api.TaskGroup{
			{
				Name: stringToPtr("app"),
				Tasks: []*api.Task{
					{
						Name:   "java-app",
						Driver: "milo",
						Artifacts: []*api.TaskArtifact{
							{
								GetterSource:  stringToPtr(gitRepoURL),
								RelativeDest:  stringToPtr("local/"),
							},
						},
						Config: map[string]interface{}{},
					},
				},
			},
		},
	}
	
	// Submit the job
	_, _, err := s.client.Jobs().Register(job, nil)
	return err
}

// SubmitJarJobFromHTTP submits a job using HTTP artifact source
func (s *LiveNomadServer) SubmitJarJobFromHTTP(jobID, jarName, httpArtifactURL string) error {
	if s.client == nil {
		return fmt.Errorf("client not initialized")
	}

	artifactSource := fmt.Sprintf("%s/%s", httpArtifactURL, jarName)
	job := &api.Job{
		ID:   &jobID,
		Name: &jobID,
		Type: stringToPtr("batch"),
		TaskGroups: []*api.TaskGroup{
			{
				Name: stringToPtr("app"),
				Tasks: []*api.Task{
					{
						Name:   "java-app",
						Driver: "milo",
						Artifacts: []*api.TaskArtifact{
							{
								GetterSource:  stringToPtr(artifactSource),
								RelativeDest:  stringToPtr("local/"),
							},
						},
						Config: map[string]interface{}{
							"dummy": "",
						},
					},
				},
			},
		},
	}

	_, _, err := s.client.Jobs().Register(job, nil)
	return err
}

// WaitForJobCompletion waits for a job to complete within the given timeout
func (s *LiveNomadServer) WaitForJobCompletion(jobID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		status, _, err := s.client.Jobs().Summary(jobID, nil)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		
		// Check if job is complete (all tasks finished)
		if status.Summary != nil {
			var totalComplete, totalFailed, totalLost, totalRunning, totalStarting, totalQueued int
			for _, taskGroup := range status.Summary {
				totalComplete += taskGroup.Complete
				totalFailed += taskGroup.Failed
				totalLost += taskGroup.Lost
				totalRunning += taskGroup.Running
				totalStarting += taskGroup.Starting
				totalQueued += taskGroup.Queued
			}
			
			total := totalComplete + totalFailed + totalLost
			if total > 0 && totalRunning == 0 && totalStarting == 0 && totalQueued == 0 {
				return nil
			}
		}
		
		time.Sleep(1 * time.Second)
	}
	
	return fmt.Errorf("job did not complete within %v", timeout)
}

// GetJobStatus returns the status of a job
func (s *LiveNomadServer) GetJobStatus(jobID string) (*api.JobSummary, error) {
	status, _, err := s.client.Jobs().Summary(jobID, nil)
	return status, err
}

// GetJobLogs returns the logs for a specific task in a job
func (s *LiveNomadServer) GetJobLogs(jobID, taskName string) (string, error) {
	// Get allocations for the job
	allocs, _, err := s.client.Jobs().Allocations(jobID, false, nil)
	if err != nil {
		return "", err
	}
	
	if len(allocs) == 0 {
		return "", fmt.Errorf("no allocations found for job %s", jobID)
	}
	
	// Get detailed allocation info
	alloc, _, err := s.client.Allocations().Info(allocs[0].ID, nil)
	if err != nil {
		return "", err
	}
	
	// Get logs from the allocation
	logsChan, errChan := s.client.AllocFS().Logs(alloc, false, taskName, "stdout", "start", 0, nil, nil)
	
	var logOutput string
	select {
	case logFrame := <-logsChan:
		if logFrame != nil {
			logOutput += string(logFrame.Data)
		}
	case err := <-errChan:
		if err != nil {
			return "", err
		}
	case <-time.After(10 * time.Second):
		// Timeout reading logs
		break
	}
	
	// Read any additional log frames
	for {
		select {
		case logFrame := <-logsChan:
			if logFrame != nil {
				logOutput += string(logFrame.Data)
			} else {
				// Channel closed
				return logOutput, nil
			}
		case err := <-errChan:
			if err != nil {
				return logOutput, err
			}
		case <-time.After(1 * time.Second):
			// No more data available
			return logOutput, nil
		}
	}
}

// GetTaskExitCode returns the exit code for a specific task
func (s *LiveNomadServer) GetTaskExitCode(jobID, taskName string) (int, error) {
	// Get allocations for the job
	allocs, _, err := s.client.Jobs().Allocations(jobID, false, nil)
	if err != nil {
		return -1, err
	}
	
	if len(allocs) == 0 {
		return -1, fmt.Errorf("no allocations found for job %s", jobID)
	}
	
	// Get task state from first allocation
	alloc, _, err := s.client.Allocations().Info(allocs[0].ID, nil)
	if err != nil {
		return -1, err
	}
	
	if taskState, exists := alloc.TaskStates[taskName]; exists {
		// Check if task is dead
		if taskState.State == "dead" && len(taskState.Events) > 0 {
			// Look for the last exit event
			for i := len(taskState.Events) - 1; i >= 0; i-- {
				event := taskState.Events[i]
				if event.Type == "Terminated" {
					return event.ExitCode, nil
				}
			}
			
			// If task is dead but no Terminated event found, check if it failed during setup
			// Look for events that indicate validation or setup failure
			for _, event := range taskState.Events {
				if event.Type == "Task Setup" && event.FailsTask {
					// Task failed during setup/validation, return -1 as a special exit code
					return -1, nil
				}
				if event.Type == "Driver Failure" {
					// Driver failed, likely during validation
					return -1, nil
				}
				// Check for validation errors
				if event.ValidationError != "" {
					// Task failed validation
					return -1, nil
				}
				// Check for setup errors
				if event.SetupError != "" {
					// Task failed during setup
					return -1, nil
				}
			}
			
			// Task is dead but no exit code found - treat as failure
			return -1, nil
		}
		
		// Task exists but is not dead yet
		if taskState.State == "pending" || taskState.State == "running" {
			return -1, fmt.Errorf("task %s is still %s", taskName, taskState.State)
		}
	}
	
	return -1, fmt.Errorf("task state not found for task %s", taskName)
}

// createConfig creates the Nomad configuration file
func (s *LiveNomadServer) createConfig() error {
	configFile, err := os.CreateTemp("", "nomad-config-*.hcl")
	if err != nil {
		return err
	}
	defer configFile.Close()
	
	s.configPath = configFile.Name()
	
	config := fmt.Sprintf(`
datacenter = "dc1"
data_dir = "%s"
log_level = "DEBUG"

server {
  enabled = true
  bootstrap_expect = 1
}

client {
  enabled = true
  servers = ["127.0.0.1:%d"]
}

plugin_dir = "/tmp/nomad-plugins"

plugin "milo" {
  config {
    shell = "bash"
  }
}

ports {
  http = %d
  rpc = %d
  serf = %d
}

addresses {
  http = "127.0.0.1"
  rpc = "127.0.0.1"
  serf = "127.0.0.1"
}

advertise {
  http = "127.0.0.1:%d"
  rpc = "127.0.0.1:%d"
  serf = "127.0.0.1:%d"
}
`, s.dataDir, s.rpcPort, s.httpPort, s.rpcPort, s.serfPort, s.httpPort, s.rpcPort, s.serfPort)
	
	_, err = configFile.WriteString(config)
	return err
}

// waitForReady waits for the server to be ready and plugins to be loaded
func (s *LiveNomadServer) waitForReady() error {
	// Wait for server to be accessible
	deadline := time.Now().Add(s.timeout)
	for time.Now().Before(deadline) {
		if s.IsAccessible() {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	
	if !s.IsAccessible() {
		// Try to read the log file for debugging
		logPath := filepath.Join(s.dataDir, "nomad.log")
		logContent, _ := os.ReadFile(logPath)
		return fmt.Errorf("server not accessible after %v. Last log output:\n%s", s.timeout, string(logContent))
	}
	
	// Log successful server startup
	fmt.Printf("Nomad server started successfully on port %d\n", s.httpPort)
	
	// Wait for plugin to load
	deadline = time.Now().Add(s.timeout)
	for time.Now().Before(deadline) {
		plugins := s.GetLoadedPlugins()
		for _, plugin := range plugins {
			if plugin == "milo" {
				fmt.Printf("Milo plugin loaded successfully\n")
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	
	// Log available plugins for debugging
	plugins := s.GetLoadedPlugins()
	return fmt.Errorf("plugin 'milo' not loaded after %v. Available plugins: %v", s.timeout, plugins)
}

// getFreePort returns a free port by creating a temporary listener
func getFreePort() int {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0
	}
	defer listener.Close()
	
	return listener.Addr().(*net.TCPAddr).Port
}

// stringToPtr converts a string to a pointer
func stringToPtr(s string) *string {
	return &s
}