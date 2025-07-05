# Epic 003: ATDD Implementation Plan

## Methodology

Following strict Acceptance Test Driven Development (ATDD) as defined in `docs/agent/atdd.md`:
- **RED**: Write failing acceptance test first
- **GREEN**: Implement minimal code to pass test  
- **REFACTOR**: Improve design while maintaining passing tests
- **5-minute rule**: If stuck >5 minutes, step back and simplify

## Implementation Order (Simple → Complex)

### 1. Live Nomad Server Management (Simplest)
**Why first**: Foundation for all other tests, isolated component

**Acceptance Test**: "Start and stop live Nomad server with plugin"
- Given a clean test environment
- When I start a Nomad test server with Milo plugin
- Then the server should be running and accessible
- And the Milo plugin should be loaded and available
- And I can stop the server cleanly

**Unit Tests Needed**:
- `TestNomadTestServer_Start()` - Server startup
- `TestNomadTestServer_Stop()` - Clean shutdown  
- `TestNomadTestServer_PluginLoad()` - Plugin availability
- `TestNomadTestServer_PortAllocation()` - Dynamic port assignment

### 2. Job Submission Framework (Simple)
**Why second**: Builds on server foundation, no execution complexity yet

**Acceptance Test**: "Submit job and track submission status"
- Given a running Nomad test server
- When I submit a valid job specification
- Then the job should be accepted by Nomad
- And I should receive a valid job ID
- And the job should appear in the job list

**Unit Tests Needed**:
- `TestJobRunner_SubmitJob()` - Job submission via API
- `TestJobRunner_ValidateJobSpec()` - Spec validation
- `TestJobRunner_GetJobStatus()` - Status retrieval
- `TestJobRunner_JobIDGeneration()` - Unique ID handling

### 3. Completion Monitoring (Medium) 
**Why third**: Adds state tracking, timeout handling complexity

**Acceptance Test**: "Wait for job completion with timeout"
- Given a submitted job
- When I wait for job completion
- Then the job should transition to completed state
- And I should be notified when state changes
- And I should timeout if job doesn't complete in time

**Unit Tests Needed**:
- `TestJobMonitor_WaitForCompletion()` - State polling
- `TestJobMonitor_TimeoutHandling()` - Timeout behavior
- `TestJobMonitor_StateTransitions()` - State change detection
- `TestJobMonitor_AllocationTracking()` - Allocation monitoring

### 4. Log Retrieval (Medium)
**Why fourth**: Requires completed jobs, adds I/O complexity  

**Acceptance Test**: "Retrieve actual task logs from allocation"
- Given a completed job allocation
- When I request task logs
- Then I should receive the actual console output
- And logs should contain expected content
- And I should handle cases where logs are not available

**Unit Tests Needed**:
- `TestLogRetriever_GetTaskLogs()` - Log fetching
- `TestLogRetriever_LogFormatting()` - Output formatting
- `TestLogRetriever_ErrorHandling()` - Missing log handling
- `TestLogRetriever_LogFiltering()` - Stdout/stderr separation

### 5. Success Path Integration (Complex)
**Why fifth**: Combines all components, tests actual JAR execution

**Acceptance Test**: "Execute JAR successfully and verify output"
- Given a test JAR file and job specification
- When I submit the job and wait for completion
- Then the JAR should execute successfully
- And the task logs should contain expected output
- And the job should complete with exit code 0

**Unit Tests Needed**:
- `TestE2EIntegration_SuccessfulExecution()` - Full pipeline
- `TestE2EIntegration_ArtifactHandling()` - JAR download/mount
- `TestE2EIntegration_ContainerExecution()` - crun integration
- `TestE2EIntegration_OutputVerification()` - Log validation

### 6. Failure Scenarios (Most Complex)
**Why last**: Requires understanding all failure modes, error propagation

**Acceptance Test**: "Handle container execution failures gracefully"
- Given a job that will fail due to container issues
- When I submit the job and wait for completion
- Then the job should fail with appropriate error messages
- And the task events should indicate the failure reason
- And no resources should be left orphaned

**Unit Tests Needed**:
- `TestE2EFailures_ContainerFailures()` - Container execution errors
- `TestE2EFailures_JavaMissing()` - Missing runtime detection
- `TestE2EFailures_InvalidJAR()` - Artifact validation errors
- `TestE2EFailures_ResourceCleanup()` - Cleanup verification

## Test Files Structure

```
e2e/live/
├── live_nomad_server_test.go     # Story 1: Server management
├── job_submission_test.go        # Story 2: Job submission  
├── completion_monitoring_test.go # Story 3: Completion tracking
├── log_retrieval_test.go         # Story 4: Log fetching
├── success_integration_test.go   # Story 5: Success scenarios
└── failure_scenarios_test.go     # Story 6: Failure handling
```

## Definition of Done

For each story:
1. ✅ All acceptance tests pass (RED → GREEN)
2. ✅ All unit tests pass with >90% coverage
3. ✅ Code follows existing patterns and conventions
4. ✅ Integration with build system (`make test-live-e2e`)
5. ✅ Documentation updated with test usage
6. ✅ Tests run reliably in CI environment