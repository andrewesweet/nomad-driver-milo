# Implementation Plan: Enhanced Artifact Validation

**Story ID**: 003  
**Story**: Implement Artifact Validation for JAR Files
**Date**: 2025-01-07

## Executive Summary

Enhance the existing artifact validation to provide better error messages, handle edge cases, and validate JAR file integrity beyond just checking the extension.

## Architecture Overview

### Current State
- Basic extension validation (`.jar` suffix check)
- Simple file existence check
- Basic error messages

### Target State
- Case-insensitive extension validation
- JAR integrity validation (ZIP structure)
- Comprehensive error messages with guidance
- Edge case handling (multiple JARs, symlinks, etc.)
- Validation before container creation

## Test-Driven Development Plan

### Test List (in order of implementation)

#### Phase 1: Enhanced Extension Validation
- [ ] Test 1: Case-insensitive extension check (.jar, .JAR, .Jar)
- [ ] Test 2: Detailed error message for non-JAR files
- [ ] Test 3: Handle files with no extension

#### Phase 2: JAR Integrity Validation
- [ ] Test 4: Valid JAR passes ZIP structure check
- [ ] Test 5: Corrupt JAR fails with specific error
- [ ] Test 6: HTML error page disguised as JAR fails
- [ ] Test 7: Empty file fails with specific error

#### Phase 3: Multiple JAR Handling
- [ ] Test 8: Single JAR found successfully
- [ ] Test 9: Multiple JARs trigger specific error
- [ ] Test 10: No JARs found triggers specific error

#### Phase 4: Edge Cases
- [ ] Test 11: Symlink to JAR validates correctly
- [ ] Test 12: Nested JARs in subdirectories handled
- [ ] Test 13: Special characters in filenames work

#### Phase 5: Integration Tests
- [ ] Test 14: Validation prevents container creation on failure
- [ ] Test 15: Error messages appear in Nomad logs
- [ ] Test 16: BDD scenarios pass

## Detailed Implementation Steps

### Step 1: Create Enhanced Validation Module (Subagent Task)

**File**: `milo/artifact_validation.go` (new)

```go
package milo

import (
    "archive/zip"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// ArtifactValidator provides comprehensive JAR validation
type ArtifactValidator struct {
    taskDir string
}

// ValidationError provides structured error information
type ValidationError struct {
    What       string
    Expected   string
    Got        string
    Suggestion string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf(
        "Error: %s\nExpected: %s\nGot: %s\nSuggestion: %s",
        e.What, e.Expected, e.Got, e.Suggestion,
    )
}
```

**Tests First**: Write comprehensive unit tests in `milo/artifact_validation_test.go`

### Step 2: Implement Case-Insensitive Extension Check (TDD)

1. Write failing test for case variations
2. Implement using `strings.ToLower()`
3. Ensure all variations pass

### Step 3: Implement JAR Integrity Validation (TDD)

1. Write tests for valid/invalid JARs
2. Implement ZIP structure validation
3. Add specific error messages for each failure type

### Step 4: Handle Multiple JAR Scenarios (TDD)

1. Write tests for 0, 1, and multiple JAR cases
2. Implement search logic with clear rules
3. Provide actionable error messages

### Step 5: Update Driver Integration (Main Task)

**File**: `milo/driver.go` - Update StartTask

1. Replace simple validation with comprehensive validator
2. Ensure validation happens before any container setup
3. Return structured errors to Nomad

### Step 6: End-to-End Testing

**File**: `e2e/live/artifact_validation_test.go`

Implement BDD scenarios from user story:
1. Invalid extension scenario
2. Missing file scenario
3. Corrupt JAR scenario
4. Multiple JAR scenario

## Implementation Order with Subagents

### Phase 1: Core Validation Logic (Subagent 1)
- Create ValidationError type
- Implement case-insensitive extension check
- Add enhanced error messages
- Unit test coverage: 100%

### Phase 2: JAR Integrity Checking (Subagent 2)
- Implement ZIP structure validation
- Add corruption detection
- Handle edge cases
- Integration test coverage: 100%

### Phase 3: Driver Integration (Subagent 3)
- Update StartTask validation flow
- Ensure early failure before container creation
- Add logging for validation failures
- Manual testing with real Nomad

### Phase 4: End-to-End Tests (Main Agent)
- Implement BDD scenarios
- Verify error messages in Nomad logs
- Test all edge cases
- Documentation updates

## Error Handling Strategy

### Validation Errors (Expected)
1. **Invalid Extension**: Clear message about JAR requirement
2. **Missing File**: Indicate download may have failed
3. **Multiple JARs**: Ask user to be specific
4. **Corrupt JAR**: Suggest re-download

### System Errors (Unexpected)
1. **Permission Denied**: Escalate with admin suggestion
2. **Disk Full**: Clear system error
3. **Read Errors**: Include OS error details

## Performance Considerations

1. **Extension Check**: Negligible (<1ms)
2. **ZIP Validation**: Fast for normal JARs (<10ms)
3. **File Search**: Use optimized directory walk
4. **Early Exit**: Fail fast on first error

## Security Considerations

1. **Symlinks**: Validate target, prevent directory traversal
2. **File Permissions**: Check readability before validation
3. **Resource Limits**: Prevent DoS with huge files
4. **Path Sanitization**: Clean all user-provided paths

## Testing Strategy

### Unit Tests
- Mock filesystem for edge cases
- Test error message formatting
- Validate all code paths

### Integration Tests
- Real JAR files of various types
- Actual filesystem operations
- Error propagation to Nomad

### E2E Tests
- Full Nomad job submission
- Log verification
- UI error display

## Definition of Done

1. **All Tests Pass**: 100% of test list items
2. **Code Coverage**: >95% for validation code
3. **Error Messages**: User-friendly and actionable
4. **Performance**: <50ms for validation
5. **Security**: No path traversal vulnerabilities
6. **Documentation**: Updated with new validation behavior

## Risk Mitigation

1. **Risk**: Breaking existing jobs
   - **Mitigation**: Maintain backward compatibility

2. **Risk**: Performance impact
   - **Mitigation**: Benchmark and optimize

3. **Risk**: False positives
   - **Mitigation**: Comprehensive test suite

## Timeline Estimate

- Phase 1 (Core Logic): 2 hours
- Phase 2 (JAR Validation): 3 hours
- Phase 3 (Integration): 2 hours
- Phase 4 (E2E): 2 hours
- **Total**: 9 hours

## Code Review Focus Areas

1. Error message clarity and helpfulness
2. Edge case handling completeness
3. Performance of file operations
4. Security of path handling
5. Test coverage and quality

## ATDD Test Sequence

Following strict TDD, implement tests in this order:

1. **Simplest**: Extension validation (just string check)
2. **Next**: File existence (filesystem interaction)
3. **Next**: Single JAR finding (directory walk)
4. **Next**: ZIP structure (archive library)
5. **Complex**: Multiple JARs (business logic)
6. **Complex**: Integration (full flow)
7. **Final**: E2E scenarios (real Nomad)