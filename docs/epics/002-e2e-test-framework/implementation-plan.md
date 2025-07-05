# Epic 002: Implementation Plan

## Technical Architecture

### Core Components

Based on Gemini's feedback and Go/Nomad best practices, the e2e test framework consists of:

#### 1. NomadTestServer
**Responsibility**: Manage Nomad server lifecycle for tests
**Key Features**:
- Use official Nomad Go API client (`github.com/hashicorp/nomad/api`)
- Dynamic port allocation to prevent conflicts
- Template-based configuration generation
- Robust process management with `t.Cleanup()`
- Capture Nomad agent stdout/stderr for debugging
- Plugin logs accessible through agent output

#### 2. JobRunner
**Responsibility**: Submit jobs and monitor execution
**Key Features**:
- API-based job submission (not CLI)
- Polling with timeouts using `require.Eventually()`
- Typed job status monitoring
- Log streaming via API client

#### 3. OutputVerifier
**Responsibility**: Validate job outputs and behavior
**Key Features**:
- Log content verification
- Exit code validation
- Resource usage checking
- Cleanup verification

#### 4. TestCleaner
**Responsibility**: Ensure clean state between tests
**Key Features**:
- Automatic process cleanup via `t.Cleanup()`
- Temporary file removal
- Port deallocation
- State reset between tests

### Directory Structure

```
e2e/
├── nomad_server.go          # Nomad server lifecycle management
├── nomad_server_test.go     # Unit tests for server management
├── job_runner.go            # Job submission and monitoring
├── job_runner_test.go       # Unit tests for job runner
├── output_verifier.go       # Output and behavior verification
├── output_verifier_test.go  # Unit tests for verification
├── test_cleaner.go          # Cleanup utilities
├── test_cleaner_test.go     # Unit tests for cleanup
├── e2e_test.go              # End-to-end test scenarios
├── fixtures/                # Test job specifications
│   ├── success.nomad
│   ├── failure.nomad
│   ├── timeout.nomad
│   └── resource-limit.nomad
└── templates/               # Nomad agent config templates
    ├── agent.hcl.tmpl
    └── plugin.hcl.tmpl
```

### Integration with Build System

#### GNUmakefile Integration
```makefile
.PHONY: test-e2e
test-e2e: build ## Run end-to-end tests
	go test -v -tags=e2e ./e2e/...

.PHONY: test-all
test-all: test-unit test-acceptance test-e2e ## Run all test types
```

#### Build Tags
Use `//go:build e2e` to separate expensive e2e tests from fast unit tests:
```go
//go:build e2e

package e2e

import "testing"

func TestJobExecution(t *testing.T) {
    // E2E test implementation
}
```

## Key Improvements from Gemini Feedback

### 1. Official Nomad Go API Client
**Instead of**: CLI commands via `os/exec`
**Use**: `github.com/hashicorp/nomad/api`

**Benefits**:
- Type safety with Go structs
- Better error handling
- More reliable than parsing CLI output
- Standard practice in Nomad ecosystem

**Implementation**:
```go
client, err := api.NewClient(api.DefaultConfig())
if err != nil {
    return nil, fmt.Errorf("failed to create Nomad client: %v", err)
}

// Submit job
_, _, err = client.Jobs().Register(jobSpec, nil)
```

### 2. Robust Process Management
**Instead of**: Manual process handling with `defer`
**Use**: `t.Cleanup()` for guaranteed cleanup

**Benefits**:
- Cleanup runs even on test panics
- More reliable than defer
- Standard Go testing pattern

**Implementation**:
```go
func startNomadServer(t *testing.T) *api.Client {
    cmd := exec.CommandContext(ctx, "nomad", "agent", "-config", configPath)
    
    // Capture stdout and stderr for debugging
    stdout, err := cmd.StdoutPipe()
    require.NoError(t, err)
    stderr, err := cmd.StderrPipe()
    require.NoError(t, err)
    
    err = cmd.Start()
    require.NoError(t, err)
    
    // Start log capture goroutines
    go func() {
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            t.Logf("NOMAD STDOUT: %s", scanner.Text())
        }
    }()
    go func() {
        scanner := bufio.NewScanner(stderr)
        for scanner.Scan() {
            t.Logf("NOMAD STDERR: %s", scanner.Text())
        }
    }()
    
    t.Cleanup(func() {
        if cmd.Process != nil {
            cmd.Process.Kill()
        }
        // Additional cleanup...
    })
    
    return client
}
```

### 3. Dynamic Port Allocation
**Instead of**: Hardcoded ports in configuration
**Use**: OS-assigned free ports with template generation

**Benefits**:
- Prevents port conflicts in CI
- Enables parallel test execution
- More reliable in shared environments

**Implementation**:
```go
func allocatePorts() (httpPort, rpcPort, serfPort int, err error) {
    // Use net.Listen to get free ports
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
```

### 4. Polling with Timeouts
**Instead of**: `time.Sleep()` for waiting
**Use**: `require.Eventually()` for reliable state checking

**Benefits**:
- Eliminates flaky tests
- Faster when conditions are met early
- Clear timeout behavior

**Implementation**:
```go
func waitForJobCompletion(t *testing.T, client *api.Client, jobID string) {
    require.Eventually(t, func() bool {
        allocs, _, err := client.Jobs().Allocations(jobID, false, nil)
        if err != nil {
            return false
        }
        
        for _, alloc := range allocs {
            if alloc.ClientStatus == "complete" {
                return true
            }
        }
        return false
    }, 30*time.Second, 500*time.Millisecond, "job never completed")
}
```

## ATDD Implementation Strategy

Following `docs/agent/atdd.md` methodology:

### Phase 1: Specification by Example
Create acceptance tests with concrete Gherkin scenarios:

```gherkin
Feature: E2E Test Framework
  As a developer
  I want to run end-to-end tests
  So that I can verify complete system behavior

Scenario: Basic job execution
  Given a Nomad server is running
  When I submit a job with a simple JAR
  Then the job should complete successfully
  And the logs should contain expected output
```

### Phase 2: Test Breakdown
Map each acceptance test to unit tests:

**AT1: Basic job execution**
- UT1: NomadServer can start and become ready
- UT2: JobRunner can submit job via API
- UT3: JobRunner can monitor job status
- UT4: OutputVerifier can retrieve and validate logs
- UT5: TestCleaner removes all resources

### Phase 3: TDD Implementation
For each unit test:
1. **RED**: Write failing test
2. **GREEN**: Write minimal code to pass
3. **REFACTOR**: Improve design while keeping tests green
4. Cycle time: Maximum 5 minutes

### Phase 4: Integration Verification
Ensure acceptance tests pass after unit implementation.

## Test Scenarios

### Happy Path Tests
1. **Simple JAR Execution**: Basic "Hello World" JAR
2. **JAR with Arguments**: JAR that processes command line arguments
3. **Long-Running JAR**: JAR that runs for several seconds
4. **Resource-Constrained JAR**: JAR with memory/CPU limits

### Error Handling Tests
1. **Missing JAR**: Job with invalid artifact URL
2. **Invalid JAR**: Corrupted or non-executable JAR
3. **Java Runtime Missing**: System without Java
4. **Resource Exceeded**: JAR that exceeds memory limits

### Lifecycle Tests
1. **Job Stop**: Manually stopping a running job
2. **Job Restart**: Restarting a failed job
3. **Multiple Jobs**: Running multiple jobs concurrently
4. **Cleanup Verification**: Ensuring all resources are cleaned up

## Performance Considerations

### Test Execution Time
- **Target**: Complete test suite in < 5 minutes
- **Strategy**: Parallel execution where possible
- **Optimization**: Reuse Nomad server for compatible tests

### Resource Usage
- **Memory**: Limit concurrent tests to prevent exhaustion
- **Disk**: Clean up temporary files aggressively
- **Network**: Use local file artifacts when possible

### CI/CD Integration
- **Docker Support**: Consider Testcontainers for perfect isolation
- **Dependency Management**: Clear documentation of required tools
- **Timeout Handling**: Fail fast on hung tests

## Future Enhancements

### Phase 2 Features
1. **Testcontainers Integration**: Perfect isolation via Docker
2. **Performance Benchmarking**: Track regression in execution times
3. **Cross-Platform Testing**: Windows and macOS support
4. **Matrix Testing**: Multiple Java and Nomad versions

### Advanced Scenarios
1. **Multi-Node Testing**: Nomad cluster with multiple nodes
2. **Service Discovery**: Jobs that register with Consul
3. **Vault Integration**: Jobs that use Vault secrets
4. **CSI Volumes**: Jobs that mount persistent storage

## Risk Mitigation

### Technical Risks
1. **Flaky Tests**: Use robust polling and proper cleanup
2. **Resource Leaks**: Comprehensive `t.Cleanup()` implementation
3. **Environment Dependencies**: Clear setup documentation

### Operational Risks
1. **CI Failures**: Provide debugging information in test output
2. **Maintenance Burden**: Keep tests simple and well-documented
3. **Version Compatibility**: Pin dependencies and test regularly