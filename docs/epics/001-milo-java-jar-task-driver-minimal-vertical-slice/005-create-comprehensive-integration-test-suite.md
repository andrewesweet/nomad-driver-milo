# User Story: Create Comprehensive Integration Test Suite

**Story ID:** 005  
**Epic:** [Milo Java JAR Task Driver - Minimal Vertical Slice](README.md)  
**Labels:** `spike`, `testing`, `integration`  
**Priority:** Medium

## User Story

As a developer, I want automated tests that verify all scenarios work, so I can confidently validate the spike.

## Test Coverage Required

- [ ] All 5 core scenarios automated
- [ ] Test artifacts created
- [ ] CI/CD pipeline integration
- [ ] Test documentation complete

## Definition of Done

- [ ] All 6 Gherkin scenarios automated
- [ ] Tests run in CI/CD pipeline
- [ ] Test artifacts created and documented
- [ ] Integration test documentation complete

## Test Scenarios

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

## Test Artifacts Specification

### hello-world.jar

```java
public class HelloWorld {
    public static void main(String[] args) {
        System.out.println("Hello from Java!");
        System.out.println("Milo driver test complete");
        System.exit(0);
    }
}
```

**Build Command**: `javac HelloWorld.java && jar cf hello-world.jar HelloWorld.class`

### long-running.jar

```java
public class LongRunning {
    public static void main(String[] args) throws InterruptedException {
        System.out.println("Starting application...");
        System.out.flush();
        
        while (true) {
            Thread.sleep(2000);
            System.out.println("Processing...");
            System.out.flush();
        }
    }
}
```

**Build Command**: `javac LongRunning.java && jar cf long-running.jar LongRunning.class`

### exit-code-test.jar

```java
public class ExitCodeTest {
    public static void main(String[] args) {
        System.out.println("Application encountered an error");
        System.exit(42);
    }
}
```

**Build Command**: `javac ExitCodeTest.java && jar cf exit-code-test.jar ExitCodeTest.class`

## Test Implementation Structure

```
tests/
├── integration/
│   ├── basic_execution_test.go
│   ├── log_streaming_test.go
│   ├── validation_test.go
│   ├── java_detection_test.go
│   └── exit_code_test.go
├── fixtures/
│   ├── hello-world.jar
│   ├── long-running.jar
│   ├── exit-code-test.jar
│   └── jobs/
│       ├── basic.nomad
│       ├── streaming.nomad
│       └── invalid.nomad
└── helpers/
    ├── nomad.go
    ├── assertions.go
    └── artifacts.go
```

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Java
        uses: actions/setup-java@v3
        with:
          java-version: '17'
          
      - name: Install crun
        run: |
          sudo apt-get update
          sudo apt-get install -y crun
          
      - name: Build Test Artifacts
        run: make test-artifacts
        
      - name: Run Integration Tests
        run: |
          make build
          go test -v ./tests/integration/...
```

## Test Utilities

### Helper Functions

```go
// Helper to run Nomad job and wait for completion
func RunJobAndWait(t *testing.T, jobFile string, expectedStatus string) {
    // Implementation
}

// Helper to verify log output
func AssertLogs(t *testing.T, jobID, taskName, expected string) {
    // Implementation
}

// Helper to check no container was created
func AssertNoContainer(t *testing.T, taskID string) {
    // Implementation
}
```

### Test Environment Setup

1. **Nomad Dev Server**
   - Start with plugin directory
   - Configure for testing
   - Clean state between tests

2. **Java Environment**
   - Mock Java installation paths
   - Test both present/absent cases
   - Version detection tests

3. **Container Runtime**
   - Verify crun available
   - Mock for unit tests
   - Real integration for e2e tests