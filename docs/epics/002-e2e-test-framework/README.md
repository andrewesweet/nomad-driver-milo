# Epic 002: End-to-End Test Framework

## Overview

This epic implements a comprehensive end-to-end (e2e) testing framework for the Milo Java JAR task driver. The framework enables automated testing of complete workflows from Nomad job submission through container execution and cleanup.

## Business Value

**Problem**: Currently, we can only test individual components (artifact validation, Java detection, container execution) in isolation. We lack confidence that the complete system works end-to-end when integrated with a real Nomad server.

**Solution**: A robust e2e test framework that:
- Starts/stops Nomad servers programmatically
- Submits real job specifications and monitors execution
- Verifies job outputs, logs, and cleanup
- Runs in CI/CD pipelines with reliable isolation

**Benefits**:
- Catch integration issues before deployment
- Verify real-world workflows work as expected
- Enable confident refactoring with regression protection
- Provide executable documentation of system behavior

## Goals

### Primary Goals
1. **Automated Nomad Server Management**: Start/stop Nomad servers for testing
2. **Job Lifecycle Testing**: Submit, monitor, and verify job execution
3. **Output Verification**: Validate job logs, exit codes, and artifacts
4. **Test Isolation**: Each test runs with clean state and resources
5. **CI/CD Integration**: Framework runs reliably in automated environments

### Secondary Goals
1. **Performance Testing**: Measure job startup and execution times
2. **Resource Usage Validation**: Verify memory and CPU constraints
3. **Error Recovery Testing**: Test failure scenarios and cleanup
4. **Multiple Java Versions**: Test compatibility across Java versions

## Success Criteria

### Must Have
- [ ] Framework can start/stop Nomad server programmatically
- [ ] Tests can submit jobs via Nomad API and monitor status
- [ ] Framework verifies job outputs and logs correctly
- [ ] Tests run in isolation with automatic cleanup
- [ ] Integration with existing build system (`make test-e2e`)
- [ ] All tests pass in CI environment

### Should Have
- [ ] Dynamic port allocation prevents test conflicts
- [ ] Framework uses Nomad Go API client for reliability
- [ ] Polling patterns avoid flaky sleep-based waits
- [ ] Test execution completes within reasonable time (< 5 minutes)

### Could Have
- [ ] Testcontainers integration for perfect isolation
- [ ] Parallel test execution capability
- [ ] Performance benchmarking and regression detection
- [ ] Cross-platform compatibility (Linux, macOS, Windows)

## Architecture Principles

### Design Principles
1. **Reliability First**: Use proven patterns (Nomad API, polling, cleanup)
2. **Test Isolation**: Each test operates independently with clean state
3. **Fast Feedback**: Tests complete quickly with clear pass/fail results
4. **Maintainable**: Clear structure, good documentation, easy to extend

### Technical Constraints
1. **ATDD Methodology**: Follow strict acceptance test driven development
2. **Production Patterns**: Use same APIs and patterns as production code
3. **No Test Pollution**: Tests must not affect each other or system state
4. **Resource Limits**: Framework must not consume excessive resources

## Implementation Approach

This epic follows **Acceptance Test Driven Development (ATDD)** as defined in `docs/agent/atdd.md`:

1. **Specification by Example**: Define acceptance tests with concrete scenarios
2. **Test Breakdown**: Map each acceptance test to required unit tests
3. **TDD Implementation**: Implement each unit test with RED-GREEN-REFACTOR cycles
4. **Integration Verification**: Ensure acceptance tests pass after unit implementation

## User Stories

See `user-stories.md` for detailed breakdown of individually valuable features.

## Implementation Plan

See `implementation-plan.md` for technical architecture and development approach.

## Dependencies

### Required Dependencies
- Nomad binary available in PATH
- Java runtime for test scenarios
- crun container runtime
- Go testing framework and godog

### Integration Points
- Builds on Epic 001 (basic JAR execution)
- Integrates with existing test infrastructure
- Uses established Nomad plugin patterns

## Risks and Mitigations

### Technical Risks
1. **Flaky Tests**: Network timing, process lifecycle
   - *Mitigation*: Robust polling, proper cleanup, deterministic ports
2. **Resource Leaks**: Stray processes, temp files
   - *Mitigation*: `t.Cleanup()` patterns, comprehensive teardown
3. **CI Environment**: Missing dependencies, permissions
   - *Mitigation*: Docker-based testing, documented requirements

### Schedule Risks
1. **Complexity Underestimate**: E2E testing is inherently complex
   - *Mitigation*: Start with minimal viable framework, iterate
2. **Integration Issues**: Nomad API compatibility, version differences
   - *Mitigation*: Use official Nomad Go client, version pinning

## Timeline

**Phase 1**: Core Framework (Week 1)
- Nomad server lifecycle management
- Basic job submission and monitoring
- Simple output verification

**Phase 2**: Enhanced Testing (Week 2)
- Dynamic configuration and port allocation
- Comprehensive test scenarios
- CI/CD integration

**Phase 3**: Advanced Features (Week 3)
- Performance testing
- Error recovery scenarios
- Documentation and examples