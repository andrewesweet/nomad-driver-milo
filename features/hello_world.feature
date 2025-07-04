Feature: Hello World Driver
  As a developer
  I want to test the hello world driver
  So that I can ensure it works correctly

  Scenario: Plugin initialization
    Given I have a hello world driver plugin
    When I create a new plugin instance
    Then the plugin should be properly initialized

  Scenario: Plugin info retrieval
    Given I have a hello world driver plugin
    When I request plugin information
    Then I should receive valid plugin information
    And the plugin name should be "hello-world-example"
    And the plugin version should be "v0.1.0"

  Scenario: Configuration schema
    Given I have a hello world driver plugin
    When I request the configuration schema
    Then I should receive a valid configuration schema
    And the schema should contain "shell" configuration

  Scenario: Task configuration schema
    Given I have a hello world driver plugin
    When I request the task configuration schema
    Then I should receive a valid task configuration schema
    And the schema should contain "greeting" configuration