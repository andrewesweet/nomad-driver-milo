# Epic 004: Code Quality & Refactoring

## Epic Description
Refactor and improve code quality based on Gemini's comprehensive code review feedback, focusing on simplifying architecture, removing legacy code, and improving maintainability.

## Success Criteria
- [ ] Executor abstraction removed in favor of direct process execution
- [ ] All greeting configuration remnants removed
- [ ] StartTask function refactored into smaller, focused functions
- [ ] Hardcoded paths parameterized in container logic
- [ ] All tests pass after refactoring
- [ ] No breaking changes to external API

## User Stories

### Story 1: Simplify Container Execution
**As a** developer maintaining the codebase  
**I want** to remove the unnecessary executor abstraction  
**So that** the container execution flow is simpler and more explicit  

**Acceptance Criteria:**
- Replace `hashicorp/nomad/drivers/shared/executor` usage with direct `os/exec.CommandContext`
- Maintain all existing error handling and process management
- Reduce dependencies where appropriate
- All existing tests continue to pass

### Story 2: Remove Legacy Configuration
**As a** user of the milo driver  
**I want** obsolete greeting configuration removed  
**So that** I'm not confused by unused configuration options  

**Acceptance Criteria:**
- Remove `greeting` field from `taskConfigSpec`
- Remove `Greeting` field from `TaskConfig` struct
- Update all related tests and documentation
- Ensure no references to greeting functionality remain

### Story 3: Refactor StartTask Function
**As a** developer working on the driver  
**I want** the StartTask function broken into smaller, focused functions  
**So that** the code is easier to understand, test, and maintain  

**Acceptance Criteria:**
- Extract `prepareTask` function for validation and environment setup
- Extract `launchTask` function for container execution
- Maintain all existing functionality and error handling
- Each function has clear, single responsibility
- Unit tests can be written for individual functions

### Story 4: Parameterize Container Paths
**As a** developer extending container functionality  
**I want** hardcoded paths in OCI spec generation to be parameterized  
**So that** the container logic is more flexible and maintainable  

**Acceptance Criteria:**
- Java mount destination path configurable
- JAR path inside container configurable  
- JAVA_HOME environment variable derived from parameters
- All existing functionality maintained
- Configuration remains backward compatible

## Dependencies
- None (can be implemented independently)

## Estimated Effort
- Story 1: 2-3 hours
- Story 2: 1-2 hours  
- Story 3: 3-4 hours
- Story 4: 2-3 hours
- **Total: 8-12 hours**

## Definition of Done
- All user stories completed with passing acceptance tests
- Code review passed (can use Gemini for validation)
- No regressions in existing functionality
- Documentation updated to reflect changes