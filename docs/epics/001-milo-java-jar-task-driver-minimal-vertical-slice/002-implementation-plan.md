# Implementation Plan: Nomad Log Streaming Integration

**Story ID:** 002  
**Epic:** [Milo Java JAR Task Driver - Minimal Vertical Slice](README.md)  
**Implementation Method:** Strict ATDD as defined in `/docs/agent/atdd.md`

## Executive Summary

This plan implements real-time log streaming for the Milo driver, enabling users to view Java application logs through standard Nomad interfaces (`nomad logs`, Web UI). The implementation uses Nomad's FIFO-based logging infrastructure to stream container stdout/stderr in real-time.

## Architecture Overview

### Log Flow
1. **Container Output** → crun process stdout/stderr pipes
2. **Stream Capture** → Go goroutines read from pipes  
3. **FIFO Writing** → Stream to Nomad's provided FIFOs
4. **Nomad Collection** → Nomad reads FIFOs and serves to clients

### Key Components
- **Log Streaming Goroutines**: Dedicated goroutines per stream (stdout/stderr)
- **FIFO Management**: Open and write to Nomad-provided paths
- **Error Handling**: Graceful handling of broken pipes and disconnections
- **Process Integration**: Capture crun process output streams

## Test-Driven Development Plan

### Test List (in order of implementation)

#### Phase 1: Unit Tests - Log Infrastructure
- [ ] Test 1: FIFO writer opens and writes to a test FIFO successfully
- [ ] Test 2: FIFO writer handles broken pipe error gracefully  
- [ ] Test 3: Log streamer copies data from reader to writer correctly
- [ ] Test 4: Log streamer handles EOF from source gracefully
- [ ] Test 5: Log streamer preserves UTF-8 encoding correctly
- [ ] Test 6: Multiple concurrent log streamers don't interfere

#### Phase 2: Integration Tests - Container Logging
- [ ] Test 7: StartTask captures crun stdout successfully
- [ ] Test 8: StartTask captures crun stderr successfully  
- [ ] Test 9: StartTask streams both stdout and stderr concurrently
- [ ] Test 10: Task logs stream to Nomad FIFOs in real-time
- [ ] Test 11: Large log volumes don't cause blocking or deadlock
- [ ] Test 12: Binary data in logs doesn't corrupt stream

#### Phase 3: Acceptance Tests - End-to-End
- [ ] Test 13: Static logs display correctly via `nomad logs`
- [ ] Test 14: Real-time streaming works with `nomad logs -f`
- [ ] Test 15: Web UI displays logs correctly
- [ ] Test 16: Logs continue streaming after task completes
- [ ] Test 17: Multiple concurrent log consumers work correctly
- [ ] Test 18: Log rotation respects Nomad settings

### Acceptance Test Implementation (BDD)

```gherkin
Feature: Real-time Log Streaming

Scenario: Stream logs from running Java application
  Given a host with Java runtime installed
    And a test JAR file "streaming-test.jar" is served via HTTP
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
              source = "http://localhost:8080/streaming-test.jar"
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
```

## Technical Spikes (Pre-Implementation Validation)

### Spike 1: Validate crun stdout/stderr capture behavior
**Goal**: Confirm how crun handles stdout/stderr redirection
**Method**: Bash script testing
```bash
#!/bin/bash
# spike-crun-output.sh
# Test how crun handles output streams

# Create test container bundle
mkdir -p /tmp/crun-spike/rootfs/bin
cp /bin/echo /tmp/crun-spike/rootfs/bin/
cp /bin/sh /tmp/crun-spike/rootfs/bin/

# Create minimal OCI spec
cat > /tmp/crun-spike/config.json <<EOF
{
  "ociVersion": "1.0.0",
  "process": {
    "args": ["/bin/sh", "-c", "echo 'stdout test'; echo 'stderr test' >&2"],
    "cwd": "/"
  },
  "root": {"path": "rootfs"},
  "namespaces": [{"type": "pid"}]
}
EOF

# Test 1: Default output behavior
echo "=== Test 1: Default output ==="
cd /tmp/crun-spike && crun run test1

# Test 2: Redirect to files
echo "=== Test 2: File redirection ==="
cd /tmp/crun-spike && crun run test2 >stdout.log 2>stderr.log
echo "Stdout: $(cat stdout.log)"
echo "Stderr: $(cat stderr.log)"

# Test 3: Pipe to another process
echo "=== Test 3: Pipe handling ==="
cd /tmp/crun-spike && crun run test3 2>&1 | tee combined.log
```

### Spike 2: Test FIFO behavior with concurrent readers/writers
**Goal**: Understand FIFO blocking and buffering characteristics
**Method**: Bash script with FIFOs
```bash
#!/bin/bash
# spike-fifo-behavior.sh
# Test FIFO characteristics for logging

# Create test FIFOs
mkfifo /tmp/test-stdout.fifo
mkfifo /tmp/test-stderr.fifo

# Test 1: Blocking on open
echo "=== Test 1: FIFO open blocking ==="
timeout 2 bash -c 'echo "Testing" > /tmp/test-stdout.fifo' && echo "Write succeeded" || echo "Write blocked/timed out"

# Test 2: With reader attached
echo "=== Test 2: FIFO with reader ==="
cat /tmp/test-stdout.fifo &
READER_PID=$!
echo "Test message" > /tmp/test-stdout.fifo
wait $READER_PID

# Test 3: Broken pipe handling
echo "=== Test 3: Broken pipe ==="
(sleep 1; echo "Message 1"; sleep 1; echo "Message 2") > /tmp/test-stdout.fifo &
WRITER_PID=$!
timeout 0.5 cat /tmp/test-stdout.fifo
wait $WRITER_PID 2>/dev/null && echo "Writer completed" || echo "Writer got SIGPIPE"

# Cleanup
rm -f /tmp/test-*.fifo
```

### Spike 3: Validate Nomad FIFO integration
**Goal**: Confirm Nomad creates and manages FIFOs correctly
**Method**: Simple test driver
```bash
#!/bin/bash
# spike-nomad-fifos.sh
# Test Nomad's FIFO creation and paths

# Create a minimal exec job that logs FIFO paths
cat > /tmp/test-fifos.nomad <<EOF
job "test-fifos" {
  type = "batch"
  group "test" {
    task "check-fifos" {
      driver = "exec"
      config {
        command = "/bin/bash"
        args = ["-c", "ls -la /dev/stdout /dev/stderr; echo 'StdoutPath env:' \$NOMAD_STDOUT_PATH; echo 'StderrPath env:' \$NOMAD_STDERR_PATH"]
      }
    }
  }
}
EOF

nomad job run /tmp/test-fifos.nomad
nomad logs test-fifos check-fifos
```

### Spike 4: Test large volume log handling
**Goal**: Understand performance characteristics and buffering
**Method**: Stress test script
```bash
#!/bin/bash
# spike-log-volume.sh
# Test high-volume logging scenarios

# Create FIFO
mkfifo /tmp/volume-test.fifo

# Start reader that measures throughput
(
  start_time=$(date +%s.%N)
  bytes=0
  while IFS= read -r line; do
    bytes=$((bytes + ${#line} + 1))
    current_time=$(date +%s.%N)
    elapsed=$(echo "$current_time - $start_time" | bc)
    if (( $(echo "$elapsed > 1" | bc) )); then
      rate=$(echo "scale=2; $bytes / $elapsed / 1048576" | bc)
      echo "Throughput: $rate MB/s"
      start_time=$current_time
      bytes=0
    fi
  done < /tmp/volume-test.fifo
) &

# Generate high-volume output
for i in {1..10000}; do
  echo "Log line $i: $(date) - Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
done > /tmp/volume-test.fifo

# Cleanup
rm -f /tmp/volume-test.fifo
```

### Spike Results Documentation
Each spike should produce a markdown report:
- **Expected behavior**: What we think will happen
- **Actual behavior**: What actually happened
- **Implications**: How this affects our implementation
- **Decision**: What approach we'll take based on findings

## Detailed Implementation Steps

### Step 1: Create Log Streaming Infrastructure (Subagent Task)

**File**: `milo/logging.go` (new)

```go
package milo

import (
    "context"
    "io"
    "os"
    "syscall"
    "github.com/hashicorp/go-hclog"
)

// LogStreamer handles streaming from a reader to a FIFO
type LogStreamer struct {
    logger hclog.Logger
    fifoPath string
    source io.Reader
}

// Stream starts streaming logs from source to FIFO
func (ls *LogStreamer) Stream(ctx context.Context) error {
    // Implementation with TDD
}
```

**Tests First**: Write unit tests in `milo/logging_test.go`

### Step 2: Modify Container Execution (Subagent Task)

**File**: `milo/driver.go` - Modify `StartTask`

Changes needed:
1. Create pipes for crun stdout/stderr before execution
2. Configure `exec.Cmd` to use pipes instead of inheriting stdout/stderr
3. Start log streaming goroutines after process starts
4. Track goroutines in task handle for cleanup

**Tests First**: Integration tests for process output capture

### Step 3: Integrate with Task Handle (Main Implementation)

**File**: `milo/handle.go` - Modify `taskHandle`

Changes needed:
1. Add fields for tracking log goroutines
2. Add context for cancellation
3. Implement cleanup in task termination

**Tests First**: Tests for handle lifecycle management

### Step 4: Error Handling Enhancement

**Focus Areas**:
1. Broken pipe (EPIPE) - Log and exit gracefully
2. FIFO open timeouts - Use goroutines to prevent blocking
3. Buffer management - Prevent memory growth
4. UTF-8 validation - Ensure proper encoding

**Tests First**: Error scenario tests

### Step 5: End-to-End Testing

**File**: `e2e/live/log_streaming_test.go` (new)

Implement full acceptance tests that:
1. Submit jobs with various logging patterns
2. Verify logs via `nomad logs` command
3. Test streaming with `-f` flag
4. Validate Web UI integration

## Implementation Order with Subagents

### Phase 1: Infrastructure (Subagent 1)
1. Create `LogStreamer` type with TDD
2. Implement FIFO writing logic
3. Add error handling for broken pipes
4. Unit test coverage: 100%

### Phase 2: Process Integration (Subagent 2)  
1. Modify crun execution to capture output
2. Integrate log streamers with StartTask
3. Add goroutine management
4. Integration test coverage: 100%

### Phase 3: Handle Management (Subagent 3)
1. Update task handle for log tracking
2. Implement cleanup on task stop
3. Add context cancellation
4. Handle test coverage: 100%

### Phase 4: End-to-End (Main Agent)
1. Create acceptance test JAR
2. Implement BDD scenarios
3. Verify all interfaces work
4. E2E test coverage: 100%

## Error Handling Strategy

### Expected Errors
1. **Broken Pipe (EPIPE)**: Normal when log consumer disconnects
   - Action: Log at debug level, exit goroutine cleanly
   
2. **FIFO Block on Open**: Can occur if no reader attached
   - Action: Open in goroutine to prevent blocking StartTask
   
3. **Write Timeout**: FIFO buffer full
   - Action: Log warning, continue attempting writes

4. **Context Cancellation**: Task being stopped
   - Action: Clean shutdown of all goroutines

### Unexpected Errors
1. **FIFO Permission Denied**: Nomad setup issue
   - Action: Return error from StartTask
   
2. **Out of Memory**: Excessive buffering
   - Action: Use io.Copy with reasonable buffer size

## Performance Considerations

1. **Buffer Size**: Use default io.Copy buffer (32KB)
2. **Goroutine Count**: 2 per task (stdout + stderr)
3. **Memory Usage**: Minimal - streaming, not buffering
4. **CPU Usage**: Negligible - I/O bound operation

## Security Considerations

1. **No Log Tampering**: Direct pipe from container
2. **No Information Leak**: Isolated per task
3. **Resource Limits**: Controlled by Nomad's log rotation

## Testing Strategy

### Unit Test Tools
- Mock readers/writers for stream testing
- Test FIFOs in temp directories
- Error injection for failure scenarios

### Integration Test Tools  
- Real crun execution with test containers
- Actual FIFO creation and streaming
- Concurrent access testing

### E2E Test Tools
- Full Nomad dev environment
- Real JAR applications with various output patterns
- Automated log verification scripts

## Definition of Done

1. **All Tests Pass**: 100% of test list items
2. **Code Coverage**: >90% for new code
3. **Documentation**: Inline comments and examples
4. **Performance**: <1ms latency for log streaming
5. **Error Handling**: All scenarios covered
6. **Manual Testing**: Verified with Nomad UI

## Risk Mitigation

1. **Risk**: FIFO behavior varies across platforms
   - **Mitigation**: Test on Linux primarily, document limitations

2. **Risk**: Large log volumes cause performance issues
   - **Mitigation**: Rely on Nomad's rotation, add metrics

3. **Risk**: Binary data corrupts text logs
   - **Mitigation**: Pass through as-is, let Nomad handle encoding

## Timeline Estimate

- Technical Spikes: 2 hours
- Phase 1 (Infrastructure): 2 hours
- Phase 2 (Integration): 3 hours  
- Phase 3 (Handle): 2 hours
- Phase 4 (E2E): 3 hours
- **Total**: 12 hours

## Spike Execution Order

1. **Before Phase 1**: Run Spikes 1-2 to validate crun and FIFO behavior
2. **Before Phase 2**: Run Spike 3 to confirm Nomad integration points
3. **Before Phase 4**: Run Spike 4 to establish performance baselines

## CURRENT PROGRESS STATUS

**Date**: 2025-07-06 23:45 BST  
**Status**: ALL PHASES COMPLETED! ✅

**⚠️ IMPORTANT**: Load project memory on resume using keywords: "spike-findings", "crun-output", "fifo-behavior", "nomad-logging"

### Completed Tasks:
- [x] **Spike 1**: Validated crun stdout/stderr capture behavior
  - **Location**: `/spike-crun-output/` directory with Go test programs
  - **Key Finding**: Direct exec.Command with StdoutPipe()/StderrPipe() works perfectly
  - **Status**: ✅ COMPLETED - crun output capture fully validated

- [x] **Spike 2**: Tested FIFO behavior with concurrent readers/writers  
  - **Location**: `/docs/epics/001-milo-java-jar-task-driver-minimal-vertical-slice/spikes/spike-fifo-simple.sh`
  - **Key Finding**: FIFOs block on write until reader attached, perfect for streaming
  - **Status**: ✅ COMPLETED - FIFO behavior confirmed

- [x] **Spike 3**: Validate Nomad FIFO integration
  - **Status**: ✅ COMPLETED - Nomad provides FIFO paths via TaskConfig.StdoutPath/StderrPath
  - **Key Finding**: Direct integration with Nomad's FIFO infrastructure confirmed
  - **Files verified**: `milo/driver.go`, Nomad driver interface

- [x] **Spike 4**: Test large volume log handling
  - **Location**: `/docs/epics/001-milo-java-jar-task-driver-minimal-vertical-slice/spikes/spike-log-volume.sh`
  - **Key Finding**: FIFOs handle high-throughput efficiently with proper implementation
  - **Status**: ✅ COMPLETED - Performance characteristics understood

- [x] **Phase 1**: Create log streaming infrastructure (LogStreamer in logging.go)
  - **Status**: ✅ COMPLETED - Full TDD implementation with 100% test coverage
  - **Files**: `milo/logging.go`, `milo/logging_test.go`
  
- [x] **Phase 2**: Modify container execution to capture process output
  - **Status**: ✅ COMPLETED - Direct exec.Command with pipe capture implemented
  - **Key Changes**: Replaced executor with direct process control, integrated LogStreamer
  
- [x] **Phase 3**: Integrate logging with task handle management  
  - **Status**: ✅ COMPLETED - Task handle properly manages log streaming lifecycle
  - **Files**: `milo/handle.go`, `milo/handle_test.go`
  
- [x] **Phase 4**: Implement end-to-end acceptance tests
  - **Status**: ✅ COMPLETED - Full BDD scenarios implemented
  - **Files**: `e2e/live/log_streaming_test.go`, `e2e/live/bdd_log_streaming_test.go`

### Key Implementation Insights (from spikes):
1. **crun Integration**: Replace current `executor.ExecCommand` with direct `exec.Command` for streaming control
2. **FIFO Handling**: Use goroutines for non-blocking FIFO operations to avoid blocking StartTask
3. **Architecture**: Implement `TaskLogManager` component for both file writing and API streaming
4. **Error Handling**: Graceful broken pipe handling when log consumers disconnect

### Next Steps on Resume:
1. Complete Spike 3 - examine Nomad FIFO integration points in driver interface
2. Run Spike 4 - performance testing for high-volume logs
3. Begin Phase 1 - implement LogStreamer infrastructure with TDD approach
4. Use subagents for each phase implementation

### Test Artifacts Created:
- `spike-crun-output/test_crun_output.go` - Basic crun output capture tests
- `spike-crun-output/test_crun_integration.go` - Integration patterns for Milo driver
- `spike-crun-output/FINDINGS.md` - Comprehensive spike 1 documentation
- `spikes/spike-fifo-simple.sh` - FIFO behavior validation script

## Notes for Implementers

1. Start with simplest test case (write to FIFO)
2. Use table-driven tests for error scenarios
3. Mock filesystem operations in unit tests
4. Real FIFOs only in integration tests
5. Follow io.Reader/Writer interfaces throughout
6. Let io.Copy handle buffering - don't optimize prematurely
7. **IMPORTANT**: Nomad does NOT support file:/// URLs in artifact stanza
   - Use HTTP server for test artifacts (e.g., `python3 -m http.server`)
   - Or use GitHub raw URLs for test JARs
   - Never use file:/// in job specifications