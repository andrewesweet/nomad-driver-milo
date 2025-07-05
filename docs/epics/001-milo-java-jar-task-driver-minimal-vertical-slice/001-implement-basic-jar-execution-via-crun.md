# User Story: Implement Basic JAR Execution via crun

**Story ID:** 001  
**Epic:** [Milo Java JAR Task Driver - Minimal Vertical Slice](README.md)  
**Labels:** `spike`, `core-functionality`, `crun`  
**Priority:** High

## User Story

As a Java application developer, I want to run my JAR file using the Milo task driver, so that I can execute my application without managing containers.

## Acceptance Criteria

- [ ] Task driver validates artifact filename ends with `.jar`
- [ ] Task driver locates single Java runtime on host
- [ ] Task driver creates container using crun
- [ ] JAR executes using: `java -jar <artifact>`
- [ ] Task status reflects container state correctly

## Technical Requirements

- [ ] Implements Nomad `drivers.DriverPlugin` interface
- [ ] Integrates with crun binary
- [ ] Mounts host Java runtime read-only
- [ ] Handles container lifecycle (start, stop, status)

## Definition of Done

- [ ] Test JAR executes successfully
- [ ] Container cleanup works properly
- [ ] Integration test passes

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

## Implementation Notes

### Key Components to Implement

1. **Driver Plugin Structure**
   - Main driver struct implementing `drivers.DriverPlugin`
   - Configuration schema for driver options
   - Task configuration validation

2. **Container Management**
   - crun integration layer
   - Container spec generation
   - Lifecycle management (create, start, stop, delete)

3. **Java Runtime Detection**
   - Common path scanning (/usr/lib/jvm, /opt/java, etc.)
   - JAVA_HOME environment variable check
   - Validation of java executable

4. **Artifact Handling**
   - Integration with Nomad's artifact downloader
   - JAR file validation
   - Path resolution within container

### Error Handling

- Invalid artifact extension → Clear error message
- Missing Java runtime → Suggest installation steps
- Container creation failure → Include crun error details
- JAR execution failure → Capture and report exit code