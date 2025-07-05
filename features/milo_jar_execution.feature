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