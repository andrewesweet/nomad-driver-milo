# Epic 002: Unit Test Mapping

## Overview

This document maps each acceptance test to the unit tests required for TDD implementation. Following ATDD methodology from `docs/agent/atdd.md`, each acceptance test is broken down into the minimal unit tests needed to satisfy it.

**Note**: This document serves as a detailed planning guide for initial implementation. As development progresses, it should be treated as a living document that may evolve based on practical implementation needs. The core principle remains: "no production code without a failing test."

## Implementation Strategy

For each unit test:
1. **RED Phase**: Write failing test (max 5 minutes)
2. **GREEN Phase**: Write minimal code to pass (max 5 minutes)  
3. **REFACTOR Phase**: Improve design while keeping tests green
4. **Commit**: Separate commits for behavioral vs structural changes

## Story 1: Nomad Server Lifecycle Management

### Acceptance Test 1: Start Nomad server successfully

**UT1.1: Create NomadServer instance**
```go
func TestNewNomadServer_CreatesInstance(t *testing.T) {
    server := NewNomadServer(t)
    assert.NotNil(t, server)
    assert.NotNil(t, server.config)
    assert.Equal(t, "", server.configPath) // not generated yet
}
```

**UT1.2: Generate dynamic configuration**
```go
func TestGenerateConfig_CreatesDynamicPorts(t *testing.T) {
    server := NewNomadServer(t)
    err := server.GenerateConfig()
    assert.NoError(t, err)
    assert.NotEqual(t, 0, server.config.HTTPPort)
    assert.NotEqual(t, 0, server.config.RPCPort)
    assert.NotEqual(t, 0, server.config.SerfPort)
    assert.True(t, isPortAvailable(server.config.HTTPPort))
}
```

**UT1.3: Start Nomad process**
```go
func TestStart_LaunchesNomadProcess(t *testing.T) {
    server := NewNomadServer(t)
    server.GenerateConfig()
    
    err := server.Start()
    assert.NoError(t, err)
    assert.NotNil(t, server.process)
    assert.True(t, isProcessRunning(server.process.Pid))
}
```

**UT1.4: Wait for server readiness**
```go
func TestWaitForReady_PollsUntilReady(t *testing.T) {
    server := NewNomadServer(t)
    server.GenerateConfig()
    server.Start()
    
    err := server.WaitForReady(10 * time.Second)
    assert.NoError(t, err)
    assert.True(t, server.IsReady())
}
```

**UT1.5: Validate API connectivity**
```go
func TestIsReady_ChecksAPIConnectivity(t *testing.T) {
    server := NewNomadServer(t)
    // Mock server that responds to API calls
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"Name":"nomad"}`))
    }))
    defer mockServer.Close()
    
    server.config.HTTPAddr = mockServer.URL
    ready := server.IsReady()
    assert.True(t, ready)
}
```

### Acceptance Test 2: Stop Nomad server cleanly

**UT1.6: Stop Nomad process gracefully**
```go
func TestStop_TerminatesProcess(t *testing.T) {
    server := NewNomadServer(t)
    server.GenerateConfig()
    server.Start()
    server.WaitForReady(5 * time.Second)
    
    err := server.Stop()
    assert.NoError(t, err)
    assert.False(t, isProcessRunning(server.process.Pid))
}
```

**UT1.7: Clean up temporary files**
```go
func TestCleanup_RemovesTemporaryFiles(t *testing.T) {
    server := NewNomadServer(t)
    server.GenerateConfig()
    configPath := server.configPath
    
    // Verify file exists
    assert.FileExists(t, configPath)
    
    server.Cleanup()
    assert.NoFileExists(t, configPath)
}
```

**UT1.8: Release allocated ports**
```go
func TestReleasePorts_FreesAllocatedPorts(t *testing.T) {
    server := NewNomadServer(t)
    server.GenerateConfig()
    httpPort := server.config.HTTPPort
    
    server.ReleasePorts()
    assert.True(t, isPortAvailable(httpPort))
}
```

## Story 2: Job Submission and Monitoring

### Acceptance Test 1: Submit job successfully via API

**UT2.1: Create JobRunner with API client**
```go
func TestNewJobRunner_CreatesWithClient(t *testing.T) {
    client := &api.Client{} // mock client
    runner := NewJobRunner(client)
    assert.NotNil(t, runner)
    assert.Equal(t, client, runner.client)
}
```

**UT2.2: Submit job via Nomad API**
```go
func TestSubmitJob_RegistersJobViaAPI(t *testing.T) {
    mockClient := &MockNomadClient{}
    runner := NewJobRunner(mockClient)
    
    jobSpec := &api.Job{
        ID:   stringPtr("test-job"),
        Name: stringPtr("test-job"),
        Type: stringPtr("batch"),
    }
    
    mockClient.On("Register", jobSpec, mock.Anything).Return(&api.JobRegisterResponse{
        JobModifyIndex: 1,
    }, nil, nil)
    
    jobID, err := runner.SubmitJob(jobSpec)
    assert.NoError(t, err)
    assert.Equal(t, "test-job", jobID)
    mockClient.AssertExpectations(t)
}
```

**UT2.3: Validate job specification format**
```go
func TestValidateJobSpec_ChecksRequiredFields(t *testing.T) {
    runner := NewJobRunner(nil)
    
    // Valid job spec
    validJob := &api.Job{
        ID:   stringPtr("test"),
        Name: stringPtr("test"),
        Type: stringPtr("batch"),
    }
    err := runner.ValidateJobSpec(validJob)
    assert.NoError(t, err)
    
    // Invalid job spec (missing ID)
    invalidJob := &api.Job{
        Name: stringPtr("test"),
        Type: stringPtr("batch"),
    }
    err = runner.ValidateJobSpec(invalidJob)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "ID is required")
}
```

### Acceptance Test 2: Monitor job execution to completion

**UT2.4: Monitor job status with polling**
```go
func TestMonitorJob_PollsStatusWithTimeout(t *testing.T) {
    mockClient := &MockNomadClient{}
    runner := NewJobRunner(mockClient)
    
    // Mock progression: pending -> running -> complete
    mockClient.On("Info", "test-job", mock.Anything).Return(&api.Job{
        Status: stringPtr("pending"),
    }, nil, nil).Once()
    mockClient.On("Info", "test-job", mock.Anything).Return(&api.Job{
        Status: stringPtr("running"),
    }, nil, nil).Once()
    mockClient.On("Info", "test-job", mock.Anything).Return(&api.Job{
        Status: stringPtr("complete"),
    }, nil, nil).Once()
    
    finalStatus, err := runner.MonitorJob("test-job", 30*time.Second)
    assert.NoError(t, err)
    assert.Equal(t, "complete", finalStatus)
}
```

**UT2.5: Get current job status**
```go
func TestGetJobStatus_RetrievesCurrentState(t *testing.T) {
    mockClient := &MockNomadClient{}
    runner := NewJobRunner(mockClient)
    
    mockClient.On("Info", "test-job", mock.Anything).Return(&api.Job{
        Status: stringPtr("running"),
    }, nil, nil)
    
    status, err := runner.GetJobStatus("test-job")
    assert.NoError(t, err)
    assert.Equal(t, "running", status)
}
```

**UT2.6: Wait for specific job status**
```go
func TestWaitForStatus_BlocksUntilDesiredStatus(t *testing.T) {
    mockClient := &MockNomadClient{}
    runner := NewJobRunner(mockClient)
    
    // Return "running" first, then "complete"
    mockClient.On("Info", "test-job", mock.Anything).Return(&api.Job{
        Status: stringPtr("running"),
    }, nil, nil).Once()
    mockClient.On("Info", "test-job", mock.Anything).Return(&api.Job{
        Status: stringPtr("complete"),
    }, nil, nil).Once()
    
    err := runner.WaitForStatus("test-job", "complete", 10*time.Second)
    assert.NoError(t, err)
}
```

## Story 3: Job Output Verification

### Acceptance Test 1: Verify successful job logs

**UT3.1: Create OutputVerifier with API client**
```go
func TestNewOutputVerifier_CreatesWithClient(t *testing.T) {
    client := &api.Client{} // mock client
    verifier := NewOutputVerifier(client)
    assert.NotNil(t, verifier)
    assert.Equal(t, client, verifier.client)
}
```

**UT3.2: Retrieve job logs via API**
```go
func TestGetJobLogs_RetrievesLogsViaAPI(t *testing.T) {
    mockClient := &MockNomadClient{}
    verifier := NewOutputVerifier(mockClient)
    
    expectedLogs := "Hello from JAR\n"
    mockClient.On("Logs", "alloc-123", true, "stdout", mock.Anything, mock.Anything).Return(
        ioutil.NopCloser(strings.NewReader(expectedLogs)), nil)
    
    logs, err := verifier.GetJobLogs("test-job")
    assert.NoError(t, err)
    assert.Equal(t, expectedLogs, logs)
}
```

**UT3.3: Verify log content matches patterns**
```go
func TestVerifyLogContent_MatchesExpectedPatterns(t *testing.T) {
    verifier := NewOutputVerifier(nil)
    
    logs := "Hello from JAR\nExecution completed successfully\n"
    
    // Should match expected content
    err := verifier.VerifyLogContent(logs, "Hello from JAR")
    assert.NoError(t, err)
    
    // Should fail for unexpected content
    err = verifier.VerifyLogContent(logs, "Error occurred")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected content not found")
}
```

**UT3.4: Extract job exit code**
```go
func TestGetExitCode_ExtractsJobExitCode(t *testing.T) {
    mockClient := &MockNomadClient{}
    verifier := NewOutputVerifier(mockClient)
    
    // Mock allocation with exit code
    mockClient.On("Allocations", "test-job", false, mock.Anything).Return([]*api.AllocationListStub{
        {ID: "alloc-123"},
    }, nil, nil)
    mockClient.On("Info", "alloc-123", mock.Anything).Return(&api.Allocation{
        TaskStates: map[string]*api.TaskState{
            "task-1": {
                Events: []*api.TaskEvent{
                    {Type: "Terminated", ExitCode: 0},
                },
            },
        },
    }, nil, nil)
    
    exitCode, err := verifier.GetExitCode("test-job")
    assert.NoError(t, err)
    assert.Equal(t, 0, exitCode)
}
```

## Story 4: Test Isolation and Cleanup

### Acceptance Test 1: Independent test execution

**UT4.1: Create TestCleaner manager**
```go
func TestNewTestCleaner_CreatesCleanupManager(t *testing.T) {
    cleaner := NewTestCleaner()
    assert.NotNil(t, cleaner)
    assert.Empty(t, cleaner.cleanupFunctions)
}
```

**UT4.2: Register cleanup functions**
```go
func TestRegisterCleanup_AddsCleanupFunction(t *testing.T) {
    cleaner := NewTestCleaner()
    called := false
    
    cleaner.RegisterCleanup(func() error {
        called = true
        return nil
    })
    
    assert.Len(t, cleaner.cleanupFunctions, 1)
    
    cleaner.ExecuteCleanup()
    assert.True(t, called)
}
```

**UT4.3: Execute all cleanup operations**
```go
func TestExecuteCleanup_RunsAllCleanupOperations(t *testing.T) {
    cleaner := NewTestCleaner()
    count := 0
    
    cleaner.RegisterCleanup(func() error { count++; return nil })
    cleaner.RegisterCleanup(func() error { count++; return nil })
    cleaner.RegisterCleanup(func() error { count++; return nil })
    
    err := cleaner.ExecuteCleanup()
    assert.NoError(t, err)
    assert.Equal(t, 3, count)
}
```

## Story 5: Dynamic Configuration

### Acceptance Test 1: Generate configuration with free ports

**UT5.1: Allocate available ports**
```go
func TestAllocatePorts_FindsAvailablePorts(t *testing.T) {
    httpPort, rpcPort, serfPort, err := AllocatePorts()
    assert.NoError(t, err)
    assert.NotEqual(t, 0, httpPort)
    assert.NotEqual(t, 0, rpcPort)
    assert.NotEqual(t, 0, serfPort)
    assert.True(t, isPortAvailable(httpPort))
    assert.True(t, isPortAvailable(rpcPort))
    assert.True(t, isPortAvailable(serfPort))
}
```

**UT5.2: Generate configuration from template**
```go
func TestGenerateConfigFromTemplate_CreatesValidConfig(t *testing.T) {
    config := &ServerConfig{
        HTTPPort: 4646,
        RPCPort:  4647,
        SerfPort: 4648,
        PluginDir: "/tmp/plugins",
    }
    
    configPath, err := GenerateConfigFromTemplate(config)
    assert.NoError(t, err)
    assert.FileExists(t, configPath)
    
    content, err := ioutil.ReadFile(configPath)
    assert.NoError(t, err)
    assert.Contains(t, string(content), "4646") // HTTP port
    assert.Contains(t, string(content), "/tmp/plugins") // Plugin dir
}
```

**UT5.3: Validate HCL configuration syntax**
```go
func TestValidateConfiguration_ChecksHCLSyntax(t *testing.T) {
    // Valid HCL
    validConfig := `
server {
  enabled = true
}
client {
  enabled = true
}
`
    err := ValidateConfiguration(validConfig)
    assert.NoError(t, err)
    
    // Invalid HCL
    invalidConfig := `
server {
  enabled = true
  invalid_syntax
}
`
    err = ValidateConfiguration(invalidConfig)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "syntax")
}
```

## Implementation Order (TDD Cycles)

### Phase 1: Foundation (Story 1)
1. UT1.1: Create NomadServer instance
2. UT1.2: Generate dynamic configuration  
3. UT1.3: Start Nomad process
4. UT1.4: Wait for server readiness
5. UT1.5: Validate API connectivity
6. UT1.6: Stop Nomad process gracefully
7. UT1.7: Clean up temporary files
8. UT1.8: Release allocated ports

### Phase 2: Core Functionality (Story 2)
9. UT2.1: Create JobRunner with API client
10. UT2.2: Submit job via Nomad API
11. UT2.3: Validate job specification format
12. UT2.4: Monitor job status with polling
13. UT2.5: Get current job status
14. UT2.6: Wait for specific job status

### Phase 3: Verification (Story 3)
15. UT3.1: Create OutputVerifier with API client
16. UT3.2: Retrieve job logs via API
17. UT3.3: Verify log content matches patterns
18. UT3.4: Extract job exit code

### Phase 4: Reliability (Story 4)
19. UT4.1: Create TestCleaner manager
20. UT4.2: Register cleanup functions
21. UT4.3: Execute all cleanup operations

### Phase 5: Configuration (Story 5)
22. UT5.1: Allocate available ports
23. UT5.2: Generate configuration from template
24. UT5.3: Validate HCL configuration syntax

## Test Utilities and Mocks

### Mock Nomad API Client
```go
type MockNomadClient struct {
    mock.Mock
}

func (m *MockNomadClient) Register(job *api.Job, q *api.WriteOptions) (*api.JobRegisterResponse, *api.WriteMeta, error) {
    args := m.Called(job, q)
    return args.Get(0).(*api.JobRegisterResponse), args.Get(1).(*api.WriteMeta), args.Error(2)
}

func (m *MockNomadClient) Info(jobID string, q *api.QueryOptions) (*api.Job, *api.QueryMeta, error) {
    args := m.Called(jobID, q)
    return args.Get(0).(*api.Job), args.Get(1).(*api.QueryMeta), args.Error(2)
}
```

### Test Helper Functions
```go
func isPortAvailable(port int) bool {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return false
    }
    ln.Close()
    return true
}

func isProcessRunning(pid int) bool {
    process, err := os.FindProcess(pid)
    if err != nil {
        return false
    }
    err = process.Signal(syscall.Signal(0))
    return err == nil
}

func stringPtr(s string) *string {
    return &s
}
```