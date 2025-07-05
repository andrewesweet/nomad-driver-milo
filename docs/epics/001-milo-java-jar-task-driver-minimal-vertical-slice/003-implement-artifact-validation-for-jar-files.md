# User Story: Implement Artifact Validation for JAR Files

**Story ID:** 003  
**Epic:** [Milo Java JAR Task Driver - Minimal Vertical Slice](README.md)  
**Labels:** `spike`, `validation`, `error-handling`  
**Priority:** Medium

## User Story

As a platform operator, I want invalid artifacts to fail gracefully with clear errors, so users understand what went wrong.

## Acceptance Criteria

- [ ] Non-JAR files fail with clear error message
- [ ] Missing files fail with clear error message
- [ ] Validation occurs before container creation
- [ ] Error messages include specific guidance

## Definition of Done

- [ ] File extension validation works
- [ ] Missing file detection works
- [ ] Error messages are user-friendly
- [ ] No resource leaks on validation failures

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

## Implementation Notes

### Validation Pipeline

1. **Pre-download Validation**
   - Check artifact URL/path format
   - Validate `.jar` extension early
   - Provide clear error for unsupported schemes

2. **Post-download Validation**
   - Verify file exists at expected location
   - Check file is readable
   - Optionally validate JAR file structure

3. **Error Message Standards**
   ```
   Error: <What went wrong>
   Expected: <What should be correct>
   Got: <What was actually provided>
   Suggestion: <How to fix it>
   ```

### Implementation Example

```go
func validateArtifact(artifact *structs.TaskArtifact) error {
    // Extension validation
    if !strings.HasSuffix(artifact.GetterSource, ".jar") {
        return fmt.Errorf(
            "Error: Artifact must be a .jar file\n" +
            "Expected: File ending with .jar\n" +
            "Got: %s\n" +
            "Suggestion: Ensure your artifact points to a Java JAR file",
            filepath.Base(artifact.GetterSource),
        )
    }
    
    // Further validations...
    return nil
}
```

### Edge Cases to Handle

1. **Case Sensitivity**
   - Accept .jar, .JAR, .Jar
   - Normalize for comparison

2. **Complex URLs**
   - Handle query parameters
   - Extract filename from URL path

3. **Archive Files**
   - Detect if artifact is .zip/.tar containing JARs
   - Provide specific error message

4. **Symbolic Links**
   - Follow symlinks safely
   - Prevent directory traversal

### User Experience

- **Clear Error Categories**
  - Invalid format
  - File not found
  - Permission denied
  - Network errors (for remote artifacts)

- **Actionable Messages**
  - Tell users exactly what to check
  - Suggest common fixes
  - Include relevant file paths

- **Fast Failure**
  - Validate before expensive operations
  - Don't create containers for invalid artifacts
  - Clean up any partial state