# Epic 002: Acceptance Tests

## Overview

This document defines acceptance tests for each user story using Gherkin syntax. These tests follow the ATDD methodology from `docs/agent/atdd.md` and will be implemented using godog.

## Story 1: Nomad Server Lifecycle Management

### Feature: Nomad Server Lifecycle
```gherkin
Feature: Nomad Server Lifecycle Management
  As a test developer
  I want to programmatically start and stop Nomad servers
  So that I can run tests against a real Nomad instance with clean state

  Scenario: Start Nomad server successfully
    Given I have a test environment with required binaries
    When I start a Nomad server with dynamic configuration
    Then the server should become ready within 10 seconds
    And the server should accept API connections
    And the server should have the Milo plugin loaded

  Scenario: Stop Nomad server cleanly
    Given I have a running Nomad server
    When I stop the Nomad server
    Then all Nomad processes should be terminated
    And all temporary files should be cleaned up
    And allocated ports should be freed

  Scenario: Multiple servers with different ports
    Given I need to run two tests simultaneously
    When I start two Nomad servers
    Then each server should get different HTTP ports
    And each server should get different RPC ports
    And both servers should be accessible independently
```

## Story 2: Job Submission and Monitoring

### Feature: Job Submission and Monitoring
```gherkin
Feature: Job Submission and Monitoring
  As a test developer
  I want to submit Nomad jobs and monitor their execution
  So that I can verify job lifecycle behavior programmatically

  Scenario: Submit job successfully via API
    Given I have a running Nomad server
    And I have a valid job specification with a JAR artifact
    When I submit the job via the Nomad API client
    Then the job should be registered successfully
    And the job should receive a valid job ID
    And the job status should be "pending"

  Scenario: Monitor job execution to completion
    Given I have submitted a job to Nomad
    When I monitor the job status with polling
    Then the job should transition to "running" status
    And the job should complete within 30 seconds
    And the final status should be "complete"

  Scenario: Handle job submission failure
    Given I have a running Nomad server
    And I have an invalid job specification
    When I attempt to submit the job
    Then the API should return an error
    And the error message should be descriptive
    And no job should be created in Nomad
```

## Story 3: Job Output Verification

### Feature: Job Output Verification
```gherkin
Feature: Job Output Verification
  As a test developer
  I want to verify job outputs, logs, and exit codes
  So that I can validate that jobs produce expected results

  Scenario: Verify successful job logs
    Given I have a completed job that prints "Hello from JAR"
    When I retrieve the job logs via API
    Then the logs should contain "Hello from JAR"
    And the exit code should be 0
    And the job status should be "complete"

  Scenario: Verify job failure logs
    Given I have a job that fails with exit code 1
    When I retrieve the job logs via API
    Then the logs should contain error information
    And the exit code should be 1
    And the job status should be "failed"

  Scenario: Verify log streaming during execution
    Given I have a long-running job that prints periodically
    When I stream the logs during job execution
    Then I should receive log messages in real-time
    And log messages should arrive in chronological order
    And streaming should stop when job completes
```

## Story 4: Test Isolation and Cleanup

### Feature: Test Isolation and Cleanup
```gherkin
Feature: Test Isolation and Cleanup
  As a test developer
  I want each test to run in isolation with automatic cleanup
  So that tests don't interfere with each other or leave artifacts

  Scenario: Independent test execution
    Given I have multiple test cases
    When I run test case A followed by test case B
    Then test case B should have a clean Nomad server
    And test case B should not see jobs from test case A
    And each test should use different port numbers

  Scenario: Cleanup after test failure
    Given I have a test that causes a panic
    When the test fails unexpectedly
    Then the Nomad server process should still be terminated
    And temporary configuration files should be removed
    And no stray processes should remain running

  Scenario: Resource cleanup verification
    Given I have completed a test that created multiple jobs
    When the test finishes
    Then all job allocations should be stopped
    And all temporary directories should be removed
    And system resources should be freed
```

## Story 5: Dynamic Configuration

### Feature: Dynamic Configuration
```gherkin
Feature: Dynamic Configuration
  As a test developer
  I want tests to use dynamic configuration with free ports
  So that tests can run reliably in CI environments

  Scenario: Generate configuration with free ports
    Given I need to start a Nomad server for testing
    When I generate the server configuration
    Then the HTTP port should be dynamically allocated
    And the RPC port should be dynamically allocated
    And the Serf port should be dynamically allocated
    And all ports should be available and not in use

  Scenario: Template-based configuration generation
    Given I have a Nomad agent configuration template
    When I generate configuration with dynamic values
    Then the configuration should include correct plugin paths
    And the configuration should include dynamic ports
    And the configuration should be valid HCL syntax

  Scenario: Handle port allocation conflicts
    Given some ports are already in use on the system
    When I attempt to allocate ports for testing
    Then the framework should find alternative free ports
    And the framework should not fail due to port conflicts
    And allocated ports should be properly released after tests
```

## Story 6: Build System Integration

### Feature: Build System Integration
```gherkin
Feature: Build System Integration
  As a developer
  I want e2e tests integrated with the existing build system
  So that I can run comprehensive tests with a single command

  Scenario: Run e2e tests via make target
    Given I have the Milo driver source code
    When I execute the e2e test suite via make target
    Then the plugin should be built automatically
    And all e2e tests should execute
    And test results should be reported clearly
    And the exit code should reflect test success/failure

  Scenario: Separate e2e tests from unit tests
    Given I want to run only fast tests during development
    When I execute only unit tests via make target
    Then e2e tests should not execute
    And only unit tests should run
    And the build should complete quickly

  Scenario: Run all test types together
    Given I want comprehensive test coverage
    When I execute all test types via make target
    Then unit tests should run first
    And acceptance tests should run second
    And e2e tests should run last
    And all test results should be aggregated
```

## Story 7: Comprehensive Test Scenarios

### Feature: Comprehensive Test Scenarios
```gherkin
Feature: Comprehensive Test Scenarios
  As a test developer
  I want a suite of test scenarios covering common use cases
  So that I can verify the system works in realistic situations

  Scenario: Simple JAR execution success
    Given I have a "Hello World" JAR file
    And I have a job specification that runs the JAR
    When I submit and execute the job
    Then the job should complete successfully
    And the output should contain "Hello World"
    And the exit code should be 0

  Scenario: JAR with command line arguments
    Given I have a JAR that processes command line arguments
    And I have a job specification with arguments "arg1 arg2"
    When I submit and execute the job
    Then the job should process the arguments correctly
    And the output should reflect the provided arguments

  Scenario: Resource-constrained job execution
    Given I have a job with memory limit 64MB
    And I have a JAR that stays within the limit
    When I submit and execute the job
    Then the job should complete successfully
    And memory usage should not exceed the limit

  Scenario: Job execution with missing Java runtime
    Given the system does not have Java installed
    When I submit a job that requires Java
    Then the job should fail with a clear error message
    And the error should indicate Java is not found
    And the job should not leave any stray processes

  Scenario: Multiple concurrent job execution
    Given I have three different JAR files
    When I submit all three jobs simultaneously
    Then all jobs should execute concurrently
    And each job should complete independently
    And no job should interfere with others
```

## Unit Test Mapping

Following ATDD methodology, each acceptance test maps to specific unit tests:

### Story 1: Nomad Server Lifecycle
**AT1: Start Nomad server successfully**
- UT1: `NewNomadServer()` creates server instance
- UT2: `GenerateConfig()` creates valid configuration
- UT3: `Start()` launches Nomad process successfully
- UT4: `WaitForReady()` polls server status until ready
- UT5: `IsReady()` validates server accepts connections

**AT2: Stop Nomad server cleanly**
- UT6: `Stop()` terminates Nomad process gracefully
- UT7: `Cleanup()` removes temporary files
- UT8: `ReleasePorts()` frees allocated ports

### Story 2: Job Submission and Monitoring
**AT1: Submit job successfully via API**
- UT1: `NewJobRunner()` creates job runner with API client
- UT2: `SubmitJob()` registers job via Nomad API
- UT3: `ValidateJobSpec()` verifies job specification format

**AT2: Monitor job execution to completion**
- UT4: `MonitorJob()` polls job status with timeout
- UT5: `GetJobStatus()` retrieves current job state
- UT6: `WaitForStatus()` blocks until desired status reached

### Story 3: Job Output Verification
**AT1: Verify successful job logs**
- UT1: `NewOutputVerifier()` creates verifier with API client
- UT2: `GetJobLogs()` retrieves logs via Nomad API
- UT3: `VerifyLogContent()` matches expected patterns
- UT4: `GetExitCode()` extracts job exit code

### Story 4: Test Isolation and Cleanup
**AT1: Independent test execution**
- UT1: `NewTestCleaner()` creates cleanup manager
- UT2: `RegisterCleanup()` adds cleanup functions
- UT3: `ExecuteCleanup()` runs all cleanup operations

### Story 5: Dynamic Configuration
**AT1: Generate configuration with free ports**
- UT1: `AllocatePorts()` finds available ports
- UT2: `GenerateConfigFromTemplate()` creates configuration
- UT3: `ValidateConfiguration()` checks HCL syntax

## Story 8: Driver Recovery Testing

### Feature: Driver Recovery Testing
```gherkin
Feature: Driver Recovery Testing
  As a test developer
  I want to test the driver recovery process when plugins crash
  So that I can verify Nomad's RecoverTask flow works correctly

  Scenario: Plugin crash during job execution
    Given I have a running job via the Milo driver
    When I simulate a plugin process crash by killing the plugin
    Then Nomad should detect the plugin failure
    And Nomad should attempt to recover the task
    And the RecoverTask method should be called
    And the job should be restored to its previous state

  Scenario: Plugin restart and task recovery
    Given I have a crashed plugin with a running task
    When the plugin process restarts
    And Nomad reconnects to the plugin
    Then the task state should be properly recovered
    And the task should continue execution or restart as appropriate
    And no resource leaks should remain

  Scenario: Multiple task recovery after plugin crash
    Given I have multiple running jobs via the Milo driver
    When the plugin process crashes unexpectedly
    Then all tasks should be marked for recovery
    And each task should be individually recovered
    And task state should be consistent after recovery
```

## Implementation Order

Following the principle of simplest to most complex:

1. **Story 1**: Nomad server lifecycle (foundation)
2. **Story 2**: Job submission and monitoring (core functionality)
3. **Story 3**: Output verification (validation capability)
4. **Story 4**: Test isolation and cleanup (reliability)
5. **Story 5**: Dynamic configuration (CI compatibility)
6. **Story 6**: Build system integration (developer experience)
7. **Story 7**: Comprehensive scenarios (coverage)
8. **Story 8**: Driver recovery testing (advanced reliability)