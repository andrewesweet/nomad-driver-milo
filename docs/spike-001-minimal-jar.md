# GitHub Issues for Milo Java JAR Driver Spike

## Epic Issue

**Title:** `[SPIKE] Milo Java JAR Task Driver - Minimal Vertical Slice`

**Labels:** `spike`, `technical-debt`, `nomad`, `java`

**Description:**
```markdown
## Goal
Prove end-to-end feasibility: User submits simple Nomad job → JAR executes via crun → user sees output in Nomad logs.

## Success Criteria
- [ ] User can submit job with `driver = "milo"` and artifact block
- [ ] JAR executes via crun container runtime
- [ ] User sees JAR output in Nomad logs
- [ ] All Gherkin scenarios pass

## Timeline
Target: 2-3 days maximum

## Definition of Done
- [ ] End-to-end demo works: job submission → JAR execution → log output
- [ ] All acceptance criteria pass with test JAR files
- [ ] Error scenarios fail gracefully with clear messages
- [ ] Integration with standard Nomad workflows (CLI, UI, API)

## Testing Approach
- **Unit Tests**: Using Go's built-in `testing` package
- **Acceptance Tests**: Using Godog for BDD/Gherkin scenarios
- **Methodology**: Strict TDD/ATDD - no production code without failing test

## Related Issues
- #[issue-number] - Core JAR execution functionality
- #[issue-number] - Artifact validation
- #[issue-number] - Error handling
- #[issue-number] - Integration testing
```

---

## Individual Feature Issues

### Issue 1: Core JAR Execution

**Title:** `[SPIKE] Implement basic JAR execution via crun`

**Labels:** `spike`, `core-functionality`, `crun`

**Description:**
```markdown
## User Story
As a Java application developer, I want to run my JAR file using the Milo task driver, so that I can execute my application without managing containers.

## Acceptance Criteria
- [ ] Task driver validates artifact filename ends with `.jar`
- [ ] Task driver locates single Java runtime on host
- [ ] Task driver creates container using crun
- [ ] JAR executes using: `java -jar <artifact>`
- [ ] Task status reflects container state correctly

## Gherkin Scenario
```gherkin
Given a host with Java runtime installed at "/usr/lib/jvm/java-17"
  And a test JAR file exists at "/tmp/hello-world.jar"
  And the JAR when executed prints exactly:
    """
    Hello from Java!
    Milo driver test complete
    """
  And the JAR exits with code 0
  And a Nomad job file "test-job.nomad" contains:
    """
    job "hello-world-test" {
      type = "batch"
      group "app" {
        task "java-app" {
          driver = "milo"
          artifact {
            source = "file:///tmp/hello-world.jar"
          }
        }
      }
    }
    """

When the user executes: `nomad job run test-job.nomad`
  And waits for task completion

Then the job status should show "dead (success)"
  And the task exit code should be 0
  And running `nomad logs hello-world-test java-app` should output exactly:
    """
    Hello from Java!
    Milo driver test complete
    """
  And the task events should include "Task completed successfully"
```

## Technical Requirements
- [ ] Implements Nomad `drivers.DriverPlugin` interface
- [ ] Integrates with crun binary
- [ ] Mounts host Java runtime read-only
- [ ] Handles container lifecycle (start, stop, status)

## Definition of Done
- [ ] Test JAR executes successfully
- [ ] Container cleanup works properly
- [ ] Integration test passes
```

---

### Issue 2: Log Streaming Integration

**Title:** `[SPIKE] Implement Nomad log streaming integration`

**Labels:** `spike`, `logging`, `nomad-integration`

**Description:**
```markdown
## User Story
As a platform user, I want to see my Java application logs through standard Nomad interfaces, so I can debug and monitor normally.

## Acceptance Criteria
- [ ] JAR stdout/stderr streams to Nomad task logs
- [ ] `nomad logs <job> <task>` shows application output
- [ ] Real-time log streaming works with `-f` flag
- [ ] Nomad web UI displays logs correctly

## Gherkin Scenarios

### Real-time Streaming Scenario
```gherkin
Given a host with Java runtime installed
  And a test JAR file exists at "/tmp/long-running.jar"
  And the JAR when executed:
    - Prints "Starting application..." immediately
    - Prints "Processing..." every 2 seconds
    - Runs until terminated
  And a Nomad job file "streaming-test.nomad" contains:
    """
    job "streaming-test" {
      type = "service"
      group "app" {
        task "java-app" {
          driver = "milo"
          artifact {
            source = "file:///tmp/long-running.jar"
          }
        }
      }
    }
    """

When the user executes: `nomad job run streaming-test.nomad`
  And waits 5 seconds
  And executes: `nomad logs -f streaming-test java-app`

Then the log output should show:
    """
    Starting application...
    Processing...
    Processing...
    """
  And new "Processing..." lines should appear every 2 seconds
  And the task status should show "running"

When the user executes: `nomad job stop streaming-test`

Then the task should terminate within 5 seconds
  And the final task status should show "dead (success)"
```

## Technical Requirements
- [ ] Container stdout/stderr captured properly
- [ ] Integration with Nomad's logging subsystem
- [ ] Real-time streaming support
- [ ] Log buffering handled correctly

## Definition of Done
- [ ] Static logs display correctly
- [ ] Streaming logs work in real-time
- [ ] Web UI integration functional
```

---

### Issue 3: Artifact Validation

**Title:** `[SPIKE] Implement artifact validation for JAR files`

**Labels:** `spike`, `validation`, `error-handling`

**Description:**
```markdown
## User Story
As a platform operator, I want invalid artifacts to fail gracefully with clear errors, so users understand what went wrong.

## Acceptance Criteria
- [ ] Non-JAR files fail with clear error message
- [ ] Missing files fail with clear error message
- [ ] Validation occurs before container creation
- [ ] Error messages include specific guidance

## Gherkin Scenarios

## Gherkin Scenarios

### Invalid Extension Scenario
```gherkin
Given a host with Java runtime installed
  And a Python script exists at "/tmp/my-script.py"
  And a Nomad job file "invalid-test.nomad" contains:
    """
    job "invalid-test" {
      type = "batch"
      group "app" {
        task "java-app" {
          driver = "milo"
          artifact {
            source = "file:///tmp/my-script.py"
          }
        }
      }
    }
    """

When the user executes: `nomad job run invalid-test.nomad`
  And waits for task completion

Then the job status should show "dead (failed)"
  And the task exit code should be non-zero
  And running `nomad logs invalid-test java-app` should contain:
    """
    Error: Artifact must be a .jar file, got: my-script.py
    """
  And the task events should include "Task failed to start"
  And no crun container should have been created
```

### Missing File Scenario
```gherkin
Given a host with Java runtime installed
  And no file exists at "/tmp/missing.jar"
  And a Nomad job file "missing-test.nomad" contains:
    """
    job "missing-test" {
      type = "batch"
      group "app" {
        task "java-app" {
          driver = "milo"
          artifact {
            source = "file:///tmp/missing.jar"
          }
        }
      }
    }
    """

When the user executes: `nomad job run missing-test.nomad`
  And waits for task completion

Then the job status should show "dead (failed)"
  And running `nomad logs missing-test java-app` should contain:
    """
    Error: Failed to download artifact: file not found
    """
  And no crun container should have been created
```

## Definition of Done
- [ ] File extension validation works
- [ ] Missing file detection works
- [ ] Error messages are user-friendly
- [ ] No resource leaks on validation failures
```

---

### Issue 4: Java Runtime Detection

**Title:** `[SPIKE] Implement Java runtime detection and mounting`

**Labels:** `spike`, `java`, `runtime-detection`

**Description:**
```markdown
## User Story
As a Java application, I want the correct Java runtime available in my container, so I can execute properly.

## Acceptance Criteria
- [ ] Task driver locates Java runtime on host
- [ ] Java runtime mounted read-only into container
- [ ] JAVA_HOME environment variable set correctly
- [ ] Missing Java fails with clear error

## Gherkin Scenarios

### Missing Java Runtime Scenario
```gherkin
Given a host with no Java runtime installed
  And a test JAR file exists at "/tmp/hello-world.jar"
  And a Nomad job file "no-java-test.nomad" contains:
    """
    job "no-java-test" {
      type = "batch"
      group "app" {
        task "java-app" {
          driver = "milo"
          artifact {
            source = "file:///tmp/hello-world.jar"
          }
        }
      }
    }
    """

When the user executes: `nomad job run no-java-test.nomad`
  And waits for task completion

Then the job status should show "dead (failed)"
  And running `nomad logs no-java-test java-app` should contain:
    """
    Error: No Java runtime found on host. Please install Java to use Milo driver.
    """
  And no crun container should have been created
```

### Exit Code Propagation Scenario
```gherkin
Given a host with Java runtime installed
  And a test JAR file exists at "/tmp/exit-code-test.jar"
  And the JAR when executed:
    - Prints "Application encountered an error"
    - Exits with code 42
  And a Nomad job file "exit-code-test.nomad" contains:
    """
    job "exit-code-test" {
      type = "batch"
      group "app" {
        task "java-app" {
          driver = "milo"
          artifact {
            source = "file:///tmp/exit-code-test.jar"
          }
        }
      }
    }
    """

When the user executes: `nomad job run exit-code-test.nomad`
  And waits for task completion

Then the job status should show "dead (failed)"
  And the task exit code should be 42
  And running `nomad logs exit-code-test java-app` should contain:
    """
    Application encountered an error
    """
```

## Technical Requirements
- [ ] Java installation discovery mechanism
- [ ] Read-only mount configuration
- [ ] Environment variable setup
- [ ] Error handling for missing Java

## Definition of Done
- [ ] Java detection works reliably
- [ ] Container has functional Java environment
- [ ] Missing Java handled gracefully
```

---

### Issue 5: Integration Testing

**Title:** `[SPIKE] Create comprehensive integration test suite`

**Labels:** `spike`, `testing`, `integration`

**Description:**
```markdown
## User Story
As a developer, I want automated tests that verify all scenarios work, so I can confidently validate the spike.

## Test Coverage Required

### Test Scenario 1: Successful JAR Execution
```gherkin
Given a host with Java runtime installed at "/usr/lib/jvm/java-17"
  And a test JAR file exists at "/tmp/hello-world.jar"
  And the JAR when executed prints exactly:
    """
    Hello from Java!
    Milo driver test complete
    """
  And the JAR exits with code 0

When the user executes: `nomad job run test-job.nomad`
  And waits for task completion

Then the job status should show "dead (success)"
  And the task exit code should be 0
  And running `nomad logs hello-world-test java-app` should output exactly:
    """
    Hello from Java!
    Milo driver test complete
    """
```

### Test Scenario 2: Real-time Log Streaming
```gherkin
Given a test JAR file exists at "/tmp/long-running.jar"
  And the JAR when executed:
    - Prints "Starting application..." immediately
    - Prints "Processing..." every 2 seconds
    - Runs until terminated

When the user executes: `nomad job run streaming-test.nomad`
  And waits 5 seconds
  And executes: `nomad logs -f streaming-test java-app`

Then the log output should show:
    """
    Starting application...
    Processing...
    Processing...
    """
  And new "Processing..." lines should appear every 2 seconds
```

### Test Scenario 3: Invalid File Extension
```gherkin
Given a Python script exists at "/tmp/my-script.py"

When the user executes: `nomad job run invalid-test.nomad`

Then the job status should show "dead (failed)"
  And running `nomad logs invalid-test java-app` should contain:
    """
    Error: Artifact must be a .jar file, got: my-script.py
    """
```

### Test Scenario 4: Missing Java Runtime
```gherkin
Given a host with no Java runtime installed

When the user executes: `nomad job run no-java-test.nomad`

Then running `nomad logs no-java-test java-app` should contain:
    """
    Error: No Java runtime found on host. Please install Java to use Milo driver.
    """
```

### Test Scenario 5: Exit Code Propagation
```gherkin
Given a test JAR file exists at "/tmp/exit-code-test.jar"
  And the JAR when executed:
    - Prints "Application encountered an error"
    - Exits with code 42

When the user executes: `nomad job run exit-code-test.nomad`

Then the task exit code should be 42
  And the job status should show "dead (failed)"
```

## Test Artifacts Needed

### hello-world.jar Specification
- **Behavior**: Prints exactly two lines to stdout
- **Line 1**: "Hello from Java!"
- **Line 2**: "Milo driver test complete"
- **Exit Code**: 0
- **Duration**: Approximately 1 second to execute

### long-running.jar Specification
- **Behavior**: 
  - Prints "Starting application..." immediately
  - Prints "Processing..." every 2 seconds
  - Runs indefinitely until terminated
- **Exit Code**: 0 on SIGTERM

### exit-code-test.jar Specification
- **Behavior**: 
  - Prints "Application encountered an error"
  - Exits with code 42
- **Duration**: Immediate exit after printing

## Definition of Done
- [ ] All 6 Gherkin scenarios automated
- [ ] Tests run in CI/CD pipeline
- [ ] Test artifacts created and documented
- [ ] Integration test documentation complete
```

---

## Issue Template for Copy/Paste

When creating these in GitHub, use this structure:

```markdown
**Title:** [SPIKE] [Brief description]

**Labels:** spike, [relevant-tech-labels]

**Assignee:** [your-username]

**Project:** Milo Task Driver

**Milestone:** Technical Spike - Phase 1

**Description:** [Copy relevant section from above]
```

---

## ATDD Implementation Plan

### Overview
This spike will be implemented using strict Acceptance Test-Driven Development (ATDD) methodology. Each feature will be built incrementally, starting with acceptance tests (using Godog) and breaking them down into unit tests (using Go test).

### Implementation Order (Simplest to Most Complex)

#### 1. Invalid Artifact Extension Validation
**Why first**: Simplest validation logic, no container runtime or Java detection needed

**Acceptance Test**: "Invalid Extension Scenario" (lines 206-236)

**Unit Tests Required**:
1. `TestTaskConfig_ExtractArtifactSource` - Parse artifact source from task config
2. `TestValidateArtifactExtension_RejectsNonJar` - Validate .jar extension requirement
3. `TestFormatInvalidExtensionError` - Generate user-friendly error message

**Implementation Steps**:
- RED: Write failing acceptance test in Godog
- For each unit test:
  - RED: Write failing unit test
  - GREEN: Write minimal code to pass
  - REFACTOR: Clean up code structure
- Verify acceptance test passes

#### 2. Missing Artifact File Validation
**Why second**: Builds on artifact validation, still no container needed

**Acceptance Test**: "Missing File Scenario" (lines 238-266)

**Unit Tests Required**:
1. `TestCheckArtifactExists` - Verify file existence after download
2. `TestHandleMissingArtifactError` - Error handling for missing files
3. `TestFormatMissingFileError` - Generate clear error message

#### 3. Missing Java Runtime Detection
**Why third**: Introduces Java detection without container complexity

**Acceptance Test**: "Missing Java Runtime Scenario" (lines 296-324)

**Unit Tests Required**:
1. `TestScanJavaInstallationPaths` - Check /usr/lib/jvm/*, /opt/java
2. `TestCheckJavaHomeEnvironment` - Read JAVA_HOME variable
3. `TestValidateJavaExecutable` - Verify java binary exists and is executable
4. `TestFormatMissingJavaError` - Generate helpful error message

#### 4. Basic JAR Execution
**Why fourth**: Core functionality with container runtime

**Acceptance Test**: "Successful JAR Execution" (lines 60-94)

**Unit Tests Required**:
1. `TestGenerateCrunCommand` - Build crun execution command
2. `TestCreateContainerSpec` - Generate OCI container specification
3. `TestMountJavaRuntime` - Configure Java runtime mount
4. `TestMountTaskDirectory` - Configure task directory mount
5. `TestCaptureContainerOutput` - Redirect stdout/stderr
6. `TestWaitForContainerExit` - Monitor container completion
7. `TestCleanupContainer` - Remove container after exit

#### 5. Exit Code Propagation
**Why fifth**: Extends basic execution with error handling

**Acceptance Test**: "Exit Code Propagation Scenario" (lines 327-357)

**Unit Tests Required**:
1. `TestCaptureNonZeroExitCode` - Handle container exit codes
2. `TestUpdateTaskStateOnFailure` - Update Nomad task state
3. `TestGenerateFailureEvent` - Create appropriate Nomad events

#### 6. Real-time Log Streaming
**Why last**: Most complex, requires concurrent I/O handling

**Acceptance Test**: "Real-time Streaming Scenario" (lines 129-169)

**Unit Tests Required**:
1. `TestInitializeLogBuffers` - Set up concurrent-safe buffers
2. `TestStreamContainerLogs` - Real-time log capture
3. `TestHandleLogStreamTermination` - Clean shutdown of streams
4. `TestIntegrateWithNomadLogger` - Connect to Nomad's log subsystem

### TDD Cycle Discipline

For each test cycle:
1. **RED Phase** (max 2 minutes)
   - Write failing test
   - Run test to confirm failure
   - Commit: "Test: Add failing test for [feature]"

2. **GREEN Phase** (max 2 minutes)
   - Write minimal code to pass
   - Run all tests
   - Commit: "Behavioral: Implement [feature]"

3. **REFACTOR Phase** (max 1 minute)
   - Improve code structure
   - Run all tests
   - Commit: "Structural: Refactor [component]"

### File Structure

```
nomad-driver-milo/
├── milo/
│   ├── driver.go          # Main driver implementation
│   ├── driver_test.go     # Driver unit tests
│   ├── config.go          # Configuration structures
│   ├── config_test.go     # Configuration tests
│   ├── artifact.go        # Artifact validation
│   ├── artifact_test.go   # Artifact tests
│   ├── java.go            # Java runtime detection
│   ├── java_test.go       # Java detection tests
│   ├── crun.go            # Container runtime integration
│   └── crun_test.go       # Container tests
├── features/
│   ├── milo_jar_execution.feature  # Acceptance tests
│   └── step_definitions_test.go    # Godog step implementations
└── test-artifacts/
    ├── hello-world.jar    # Test JAR files
    ├── long-running.jar
    └── exit-code-test.jar
```

### Success Metrics

- 100% test coverage for new code
- All acceptance tests passing
- No production code without failing test first
- Clear separation of behavioral and structural changes
- Average cycle time under 5 minutes