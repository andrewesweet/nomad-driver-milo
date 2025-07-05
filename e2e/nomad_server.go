//go:build e2e

package e2e

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"text/template"
)

// ServerConfig holds Nomad server configuration
type ServerConfig struct {
	HTTPPort   int
	RPCPort    int
	SerfPort   int
	PluginDir  string
	PluginName string
}

// NomadTestServer manages Nomad server lifecycle for tests
type NomadTestServer struct {
	config     *ServerConfig
	configPath string
	process    *exec.Cmd
	t          *testing.T
}

// NewNomadServer creates a new NomadTestServer instance
func NewNomadServer(t *testing.T) *NomadTestServer {
	return &NomadTestServer{
		t:          t,
		config:     &ServerConfig{},
		configPath: "",
	}
}

// GenerateConfig generates dynamic configuration with free ports
func (s *NomadTestServer) GenerateConfig() error {
	httpPort, rpcPort, serfPort, err := allocatePorts()
	if err != nil {
		return fmt.Errorf("failed to allocate ports: %v", err)
	}

	s.config.HTTPPort = httpPort
	s.config.RPCPort = rpcPort
	s.config.SerfPort = serfPort
	s.config.PluginDir = "/tmp/nomad-plugins"
	s.config.PluginName = "nomad-driver-milo"

	return nil
}

// GenerateConfigFile creates a configuration file from template
func (s *NomadTestServer) GenerateConfigFile() error {
	// Read template
	templatePath := filepath.Join("templates", "agent.hcl.tmpl")
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %v", err)
	}

	// Parse template
	tmpl, err := template.New("agent").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Create temporary file
	tmpFile, err := ioutil.TempFile("", "nomad-agent-*.hcl")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	// Execute template
	err = tmpl.Execute(tmpFile, s.config)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	s.configPath = tmpFile.Name()
	s.t.Cleanup(func() {
		os.Remove(s.configPath)
	})

	return nil
}

// Start launches the Nomad server process
func (s *NomadTestServer) Start() error {
	ctx := context.Background()
	s.process = exec.CommandContext(ctx, "nomad", "agent", "-config", s.configPath)
	
	// Capture stdout and stderr for debugging
	stdout, err := s.process.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := s.process.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}
	
	// Start the process
	err = s.process.Start()
	if err != nil {
		return fmt.Errorf("failed to start nomad process: %v", err)
	}
	
	// Start log capture goroutines
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			s.t.Logf("NOMAD STDOUT: %s", scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			s.t.Logf("NOMAD STDERR: %s", scanner.Text())
		}
	}()
	
	// Setup cleanup
	s.t.Cleanup(func() {
		if s.process != nil && s.process.Process != nil {
			s.process.Process.Kill()
		}
	})
	
	return nil
}

// allocatePorts finds three available ports
func allocatePorts() (httpPort, rpcPort, serfPort int, err error) {
	listeners := make([]net.Listener, 3)
	defer func() {
		for _, l := range listeners {
			if l != nil {
				l.Close()
			}
		}
	}()

	for i := range listeners {
		listeners[i], err = net.Listen("tcp", ":0")
		if err != nil {
			return 0, 0, 0, err
		}
	}

	httpPort = listeners[0].Addr().(*net.TCPAddr).Port
	rpcPort = listeners[1].Addr().(*net.TCPAddr).Port
	serfPort = listeners[2].Addr().(*net.TCPAddr).Port

	return httpPort, rpcPort, serfPort, nil
}

// isPortAvailable checks if a port is available
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// isProcessRunning checks if a process is running
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}