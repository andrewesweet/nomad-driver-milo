# User Story: Implement Java Runtime Detection and Mounting

**Story ID:** 004  
**Epic:** [Milo Java JAR Task Driver - Minimal Vertical Slice](README.md)  
**Labels:** `spike`, `java`, `runtime-detection`  
**Priority:** High

## User Story

As a Java application, I want the correct Java runtime available in my container, so I can execute properly.

## Acceptance Criteria

- [ ] Task driver locates Java runtime on host
- [ ] Java runtime mounted read-only into container
- [ ] JAVA_HOME environment variable set correctly
- [ ] Missing Java fails with clear error

## Technical Requirements

- [ ] Java installation discovery mechanism
- [ ] Read-only mount configuration
- [ ] Environment variable setup
- [ ] Error handling for missing Java

## Definition of Done

- [ ] Java detection works reliably
- [ ] Container has functional Java environment
- [ ] Missing Java handled gracefully

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

## Implementation Notes

### Java Detection Strategy

1. **Search Locations** (in order)
   ```go
   var javaSearchPaths = []string{
       "/usr/lib/jvm/default-java",
       "/usr/lib/jvm/java-17-openjdk",
       "/usr/lib/jvm/java-11-openjdk",
       "/usr/local/openjdk17",
       "/usr/local/openjdk11",
       "/opt/java/openjdk",
       // Check JAVA_HOME env var
       os.Getenv("JAVA_HOME"),
   }
   ```

2. **Validation Steps**
   - Check directory exists
   - Verify `bin/java` executable present
   - Test execution with `java -version`
   - Cache result for performance

3. **Container Mount Configuration**
   ```json
   {
     "mounts": [
       {
         "source": "/usr/lib/jvm/java-17-openjdk",
         "destination": "/opt/java",
         "type": "bind",
         "options": ["rbind", "ro"]
       }
     ],
     "env": [
       "JAVA_HOME=/opt/java",
       "PATH=/opt/java/bin:$PATH"
     ]
   }
   ```

### Error Handling

1. **No Java Found**
   ```
   Error: No Java runtime found on host. Please install Java to use Milo driver.
   
   Searched locations:
   - /usr/lib/jvm/default-java (not found)
   - /usr/lib/jvm/java-17-openjdk (not found)
   - /usr/local/openjdk17 (not found)
   - JAVA_HOME environment variable (not set)
   
   To fix:
   1. Install Java: sudo apt install openjdk-17-jdk
   2. Or set JAVA_HOME to existing installation
   ```

2. **Invalid Java Installation**
   ```
   Error: Java installation at /opt/java appears invalid
   
   Missing executable: /opt/java/bin/java
   
   Please verify Java installation is complete.
   ```

### Performance Considerations

1. **Caching**
   - Cache Java location after first detection
   - Invalidate on driver restart
   - Store in driver state

2. **Lazy Detection**
   - Don't search for Java until first task
   - Fail fast if no Java when needed

3. **Mount Optimization**
   - Use read-only mounts
   - Bind mount only necessary directories
   - Avoid mounting unnecessary files

### Security Considerations

1. **Read-only Mounts**
   - Prevent container from modifying host Java
   - Use `ro` mount option

2. **Path Restrictions**
   - Only mount Java directory, not parent paths
   - Validate mount paths are under expected locations

3. **Environment Isolation**
   - Clear sensitive environment variables
   - Only pass through necessary Java settings