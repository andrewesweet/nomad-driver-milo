# Epic 005: Test Framework Completion

## Epic Description
Complete the test framework implementation to provide real end-to-end validation of the milo driver functionality, addressing Gemini's feedback about mock implementations and missing live test scenarios.

## Success Criteria
- [ ] Godog acceptance tests execute against live Nomad servers
- [ ] All scenarios from epic-003 implementation plan are covered
- [ ] Tests detect real integration failures (like container execution issues)
- [ ] Test framework provides reliable feedback on driver functionality
- [ ] CI/CD pipeline includes comprehensive test execution

## User Stories

### Story 1: Implement Real Acceptance Test Execution
**As a** developer using acceptance tests for validation  
**I want** Godog step definitions to execute real Nomad commands  
**So that** acceptance tests provide genuine validation of driver behavior  

**Acceptance Criteria:**
- Replace mock implementations in `features/step_definitions_test.go`
- Step definitions drive the live E2E framework (`e2e/live/`)
- Tests execute real `nomad run`, `nomad status`, `nomad logs` commands
- Tests validate actual Nomad API responses and job states
- Error scenarios trigger real failure conditions

### Story 2: Complete Live E2E Test Scenarios
**As a** developer ensuring driver reliability  
**I want** comprehensive E2E test scenarios implemented  
**So that** I can validate all critical driver functionality  

**Acceptance Criteria:**
- Implement job submission and success validation
- Implement log retrieval and content verification
- Implement failure handling and error reporting tests
- Implement resource cleanup and isolation tests
- All scenarios from `docs/epic-003-atdd-implementation-plan.md` covered

### Story 3: Integrate Container Execution Validation
**As a** developer preventing integration failures  
**I want** E2E tests to validate actual container execution  
**So that** I can detect JAR execution issues before manual testing  

**Acceptance Criteria:**
- Tests download real JAR artifacts
- Tests execute JARs in actual crun containers
- Tests validate JAR output in Nomad logs
- Tests detect and report container execution failures
- Tests verify Java runtime availability and functionality

### Story 4: Enhance Test Framework Reliability
**As a** developer running tests in different environments  
**I want** the test framework to be robust and reliable  
**So that** tests provide consistent results across different setups  

**Acceptance Criteria:**
- Tests handle dynamic port allocation correctly
- Tests provide proper cleanup on failure
- Tests have appropriate timeouts and retry logic
- Tests run reliably in CI/CD environments
- Tests provide clear failure diagnostics

## Dependencies
- Epic 003 (Live E2E framework) must be completed
- Epic 004 Story 2 (remove greeting config) should be completed first

## Estimated Effort
- Story 1: 4-6 hours
- Story 2: 6-8 hours
- Story 3: 4-6 hours
- Story 4: 3-4 hours
- **Total: 17-24 hours**

## Definition of Done
- All acceptance tests execute against live Nomad servers
- Live E2E tests cover all critical scenarios
- Tests reliably detect integration failures
- Test suite runs successfully in CI/CD pipeline
- Documentation updated with test execution instructions

## Test Scenarios to Implement

### Core Functionality
1. **JAR Download and Execution**
   - Submit job with artifact block
   - Validate JAR download and execution
   - Verify expected output in logs

2. **Error Handling**
   - Invalid JAR file handling
   - Missing Java runtime handling
   - Container execution failures

3. **Resource Management**
   - Task lifecycle (start, stop, destroy)
   - Resource cleanup on task completion
   - Multiple concurrent tasks

### Integration Validation
4. **Container Isolation**
   - Multiple tasks don't interfere
   - Proper namespace isolation
   - File system isolation

5. **Nomad Integration** 
   - Job status reporting
   - Log streaming
   - Task recovery after agent restart