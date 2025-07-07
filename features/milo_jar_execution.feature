Feature: Milo Java JAR Task Driver
  As a Java application developer
  I want to run my JAR files using the Milo task driver
  So that I can execute my applications without managing containers

  Scenario: Invalid artifact extension
    Given a host with Java runtime installed at "/usr/lib/jvm/java-17"
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
    When the user executes: "nomad job run invalid-test.nomad"
    And waits for task completion
    Then the job status should show "dead (failed)"
    And the task exit code should be non-zero
    And running "nomad logs invalid-test java-app" should contain:
      """
      Error: Artifact must be a .jar file, got: my-script.py
      """
    And the task events should include "Task failed to start"
    And no crun container should have been created

  Scenario: Missing artifact file
    Given a host with Java runtime installed at "/usr/lib/jvm/java-17"
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
    When the user executes: "nomad job run missing-test.nomad"
    And waits for task completion
    Then the job status should show "dead (failed)"
    And running "nomad logs missing-test java-app" should contain:
      """
      Error: Failed to download artifact: file not found
      """
    And no crun container should have been created

  Scenario: Missing Java runtime
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
    When the user executes: "nomad job run no-java-test.nomad"
    And waits for task completion
    Then the job status should show "dead (failed)"
    And running "nomad logs no-java-test java-app" should contain:
      """
      Error: No Java runtime found on host. Please install Java to use Milo driver.
      """
    And no crun container should have been created

  Scenario: Successful JAR execution
    Given a host with Java runtime installed at "/usr/lib/jvm/java-17"
    And a test JAR file exists at "/tmp/hello-world.jar"
    And the JAR when executed prints exactly:
      """
      Hello from Java!
      Milo driver test complete
      """
    And the JAR exits with code 0
    And a Nomad job file "success-test.nomad" contains:
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
    When the user executes: "nomad job run success-test.nomad"
    And waits for task completion
    Then the job status should show "dead (success)"
    And the task exit code should be 0
    And running "nomad logs hello-world-test java-app" should output exactly:
      """
      Hello from Java!
      Milo driver test complete
      """
    And the task events should include "Terminated: Exit Code: 0"
    
  Scenario: Container spec validation for crun
    Given a host with Java runtime installed at "/usr/lib/jvm/java-21-openjdk-amd64"
    And a test JAR file exists at "/tmp/hello-world.jar"
    And a Nomad job file "container-test.nomad" contains:
      """
      job "container-test" {
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
    When the user executes: "nomad job run container-test.nomad"
    Then the container OCI spec should include Linux namespaces
    And the container should start without crun configuration errors