# Epic 003: Live Integration E2E Tests

## Overview

Implement comprehensive end-to-end tests that execute real job specifications against live Nomad servers to catch integration issues that mock tests miss.

## Problem Statement

Current e2e tests use mock data and don't actually execute jobs through Nomad, failing to catch critical integration issues like Java library dependencies, container execution failures, and real-world deployment problems.

## Success Criteria

- E2E tests start real Nomad servers with Milo plugin loaded
- Tests submit actual job specifications and wait for completion
- Tests verify real log output from executed tasks
- Tests catch container execution failures before manual testing
- Tests validate both success and failure scenarios

## Implementation Plan

### Phase 1: Live Nomad Server Integration
1. **Enhanced NomadTestServer** - Start real Nomad processes with plugin loading
2. **Job Submission Framework** - Submit real job specs via Nomad API
3. **Completion Monitoring** - Wait for job state transitions with timeouts
4. **Log Retrieval** - Fetch actual task logs from Nomad allocations

### Phase 2: Test Scenarios
1. **Success Path** - Valid JAR executes successfully with expected output
2. **Java Missing** - Graceful failure when Java runtime unavailable
3. **Invalid JAR** - Proper error handling for corrupted/missing artifacts
4. **Container Failures** - Detect and report container execution issues

### Phase 3: Test Infrastructure
1. **Test Isolation** - Each test gets clean Nomad server instance
2. **Resource Management** - Proper cleanup of servers, jobs, allocations
3. **Parallel Execution** - Tests run independently without conflicts
4. **CI Integration** - Tests work in automated build environments