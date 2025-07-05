# Epic: Milo Java JAR Task Driver - Minimal Vertical Slice

**Epic ID:** 001  
**Type:** Technical Spike  
**Labels:** `spike`, `technical-debt`, `nomad`, `java`  
**Timeline:** 2-3 days maximum

## Goal

Prove end-to-end feasibility: User submits simple Nomad job → JAR executes via crun → user sees output in Nomad logs.

## Success Criteria

- [ ] User can submit job with `driver = "milo"` and artifact block
- [ ] JAR executes via crun container runtime
- [ ] User sees JAR output in Nomad logs
- [ ] All Gherkin scenarios pass

## Definition of Done

- [ ] End-to-end demo works: job submission → JAR execution → log output
- [ ] All acceptance criteria pass with test JAR files
- [ ] Error scenarios fail gracefully with clear messages
- [ ] Integration with standard Nomad workflows (CLI, UI, API)

## User Stories

1. [Core JAR Execution](001-implement-basic-jar-execution-via-crun.md) - Implement basic JAR execution via crun
2. [Log Streaming Integration](002-implement-nomad-log-streaming-integration.md) - Implement Nomad log streaming integration
3. [Artifact Validation](003-implement-artifact-validation-for-jar-files.md) - Implement artifact validation for JAR files
4. [Java Runtime Detection](004-implement-java-runtime-detection-and-mounting.md) - Implement Java runtime detection and mounting
5. [Integration Testing](005-create-comprehensive-integration-test-suite.md) - Create comprehensive integration test suite

## Technical Overview

The Milo driver will execute Java JAR files using container runtime (crun) while mounting the host's Java runtime. This approach provides:

- Process isolation via containers
- Reuse of existing Java installations
- Standard Nomad integration patterns
- Clear error handling and user feedback

## Architecture Decisions

- **Container Runtime**: crun (lightweight OCI runtime)
- **Java Detection**: Scan common Java installation paths
- **Logging**: Direct stdout/stderr capture from container
- **Artifact Support**: Leverage Nomad's built-in artifact stanza

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Multiple Java versions on host | Use first found or allow configuration |
| Missing Java runtime | Fail fast with clear error message |
| Container overhead | Use lightweight crun runtime |
| Log streaming complexity | Follow Nomad driver best practices |