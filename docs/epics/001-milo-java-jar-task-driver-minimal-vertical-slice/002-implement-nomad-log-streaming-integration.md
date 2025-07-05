# User Story: Implement Nomad Log Streaming Integration

**Story ID:** 002  
**Epic:** [Milo Java JAR Task Driver - Minimal Vertical Slice](README.md)  
**Labels:** `spike`, `logging`, `nomad-integration`  
**Priority:** High

## User Story

As a platform user, I want to see my Java application logs through standard Nomad interfaces, so I can debug and monitor normally.

## Acceptance Criteria

- [ ] JAR stdout/stderr streams to Nomad task logs
- [ ] `nomad logs <job> <task>` shows application output
- [ ] Real-time log streaming works with `-f` flag
- [ ] Nomad web UI displays logs correctly

## Technical Requirements

- [ ] Container stdout/stderr captured properly
- [ ] Integration with Nomad's logging subsystem
- [ ] Real-time streaming support
- [ ] Log buffering handled correctly

## Definition of Done

- [ ] Static logs display correctly
- [ ] Streaming logs work in real-time
- [ ] Web UI integration functional

## Gherkin Scenario: Real-time Streaming

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

## Implementation Notes

### Log Capture Architecture

1. **Container Output Streams**
   - Configure crun to redirect stdout/stderr
   - Use named pipes or file descriptors
   - Handle both line-buffered and unbuffered output

2. **Nomad Logger Integration**
   - Implement `LogEventFn` callback
   - Use Nomad's `fifo` package for streaming
   - Proper cleanup on task termination

3. **Buffering Strategy**
   - Line-based buffering for stdout
   - Immediate forwarding for stderr
   - Handle partial lines correctly

### Key Integration Points

```go
// Example logger setup
logger := taskConfig.Logger
stdout := logger.StdoutFifo()
stderr := logger.StderrFifo()

// Configure container with log pipes
containerConfig := &crun.Config{
    Stdout: stdout,
    Stderr: stderr,
}
```

### Error Scenarios

- **Broken Pipe**: Handle gracefully when log consumer disconnects
- **Buffer Overflow**: Implement circular buffer with size limits
- **Character Encoding**: Ensure UTF-8 handling
- **Log Rotation**: Respect Nomad's log rotation settings

### Testing Requirements

1. **Unit Tests**
   - Log line parsing
   - Buffer management
   - Error handling

2. **Integration Tests**
   - Real-time streaming verification
   - Multiple concurrent log consumers
   - Large log volume handling
   - Binary output handling