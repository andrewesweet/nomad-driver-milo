# BDD Testing Gap Implementation Plan

## Overview
Based on analysis and expert review, this plan addresses the disconnect between Gherkin scenarios and actual test implementation by creating true BDD tests that validate exact acceptance criteria using live system integration.

---

## Phase 1: Core Test Infrastructure (Priority: High)

### 1. `features/bdd_test_context.go` - New File
Create robust test context with improved state management:

```go
package features

import (
    "github.com/andrewesweet/nomad-driver-milo/e2e/live"
    "github.com/hashicorp/nomad/api"
)

type BDDTestContext struct {
    t               *testing.T
    nomadServer     *live.LiveNomadServer
    nomadClient     *api.Client
    tempFiles       map[string]string      // filename -> path
    nomadJobFiles   map[string]string      // filename -> content
    currentJobID    string
    currentAllocID  string
    jobAllocations  map[string]*api.Allocation
    expectedOutput  string
    testJavaPaths   []string              // For controlled Java detection
    origJavaPaths   []string              // Backup of original paths
}

func (ctx *BDDTestContext) setup() error {
    // Initialize Nomad server from existing e2e infrastructure
    ctx.nomadServer = live.NewLiveNomadServer(ctx.t)
    if err := ctx.nomadServer.Start(); err != nil {
        return fmt.Errorf("failed to start Nomad server: %v", err)
    }
    
    ctx.nomadClient = ctx.nomadServer.GetClient()
    return nil
}

func (ctx *BDDTestContext) cleanup() {
    // Cleanup all resources
    if ctx.nomadServer != nil {
        ctx.nomadServer.Stop()
    }
    // Clean temp files, restore Java paths, etc.
}
```

### 2. `features/features_test.go` - Refactored
Implement improved godog integration following best practices:

```go
func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"features"},
            TestingT: t,
            Tags:     "~@skip",
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}

func InitializeScenario(sc *godog.ScenarioContext) {
    testCtx := &BDDTestContext{}

    sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
        testCtx.t = sc.T()
        return ctx, testCtx.setup()
    })

    sc.After(func(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
        testCtx.cleanup()
        return ctx, nil
    })

    // Register step definitions as methods
    sc.Step(`^a host with Java runtime installed at "([^"]*)"$`, testCtx.aHostWithJavaRuntimeInstalledAt)
    sc.Step(`^a host with no Java runtime installed$`, testCtx.aHostWithNoJavaRuntimeInstalled)
    sc.Step(`^a test JAR file exists at "([^"]*)"$`, testCtx.aTestJARFileExistsAt)
    sc.Step(`^the user executes: "([^"]*)"$`, testCtx.theUserExecutes)
    sc.Step(`^the job status should show "([^"]*)"$`, testCtx.theJobStatusShouldShow)
    sc.Step(`^running "([^"]*)" should output exactly:$`, testCtx.runningShouldOutputExactly)
    sc.Step(`^the container OCI spec should include Linux namespaces$`, testCtx.theContainerOCISpecShouldIncludeLinuxNamespaces)
    // ... register all other steps
}
```

---

## Phase 2: Robust Step Implementations

### 3. `features/step_definitions.go` - Complete Rewrite
Consolidate step definitions with safe, controlled implementations:

```go
package features

// GIVEN steps - Environment setup
func (ctx *BDDTestContext) aHostWithJavaRuntimeInstalledAt(path string) error {
    // Verify Java exists at specified path
    javaExec := filepath.Join(path, "bin", "java")
    if _, err := os.Stat(javaExec); err != nil {
        return fmt.Errorf("Java not found at %s: %v", javaExec, err)
    }
    
    // Set controlled Java paths for the driver to use
    ctx.testJavaPaths = []string{path}
    return nil
}

func (ctx *BDDTestContext) aHostWithNoJavaRuntimeInstalled() error {
    // Use empty search paths instead of modifying global PATH
    ctx.testJavaPaths = []string{"/nonexistent/path"}
    return nil
}

func (ctx *BDDTestContext) aTestJARFileExistsAt(path string) error {
    // Copy test JAR to specified location
    testJar := "test-artifacts/hello-world.jar"
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }
    return copyFile(testJar, path)
}

// WHEN steps - Actions
func (ctx *BDDTestContext) theUserExecutes(command string) error {
    parts := strings.Fields(command)
    if len(parts) != 4 || parts[0] != "nomad" || parts[1] != "job" || parts[2] != "run" {
        return fmt.Errorf("unexpected command format: %s", command)
    }
    
    jobFile := parts[3]
    jobSpec := ctx.nomadJobFiles[jobFile]
    
    // Parse job ID from HCL
    jobID := extractJobIDFromHCL(jobSpec)
    ctx.currentJobID = jobID
    
    // Submit job with controlled Java paths
    job, err := jobspec.ParseHCL(jobSpec)
    if err != nil {
        return fmt.Errorf("failed to parse job HCL: %v", err)
    }
    
    // Inject test Java paths via task meta
    for _, taskGroup := range job.TaskGroups {
        for _, task := range taskGroup.Tasks {
            if task.Driver == "milo" {
                if task.Meta == nil {
                    task.Meta = make(map[string]string)
                }
                task.Meta["test_java_paths"] = strings.Join(ctx.testJavaPaths, ":")
                task.Meta["test_mode"] = "true" // Enable OCI spec logging
            }
        }
    }
    
    _, _, err = ctx.nomadClient.Jobs().Register(job, nil)
    return err
}

// THEN steps - Validations
func (ctx *BDDTestContext) theJobStatusShouldShow(expectedStatus string) error {
    // Wait for job completion
    if err := ctx.waitForJobCompletion(30 * time.Second); err != nil {
        return err
    }
    
    // Fetch job summary
    summary, _, err := ctx.nomadClient.Jobs().Summary(ctx.currentJobID, nil)
    if err != nil {
        return err
    }
    
    // Validate status based on expected format
    if strings.Contains(expectedStatus, "success") {
        for _, taskGroup := range summary.Summary {
            if taskGroup.Complete == 0 || taskGroup.Failed > 0 {
                return fmt.Errorf("job did not complete successfully")
            }
        }
    } else if strings.Contains(expectedStatus, "failed") {
        hasFailures := false
        for _, taskGroup := range summary.Summary {
            if taskGroup.Failed > 0 {
                hasFailures = true
                break
            }
        }
        if !hasFailures {
            return fmt.Errorf("job was expected to fail but didn't")
        }
    }
    
    return nil
}

func (ctx *BDDTestContext) runningShouldOutputExactly(command string, expected *godog.DocString) error {
    // Fetch logs from allocation
    logs, err := ctx.fetchJobLogs()
    if err != nil {
        return err
    }
    
    expectedOutput := strings.TrimSpace(expected.Content)
    actualOutput := strings.TrimSpace(logs)
    
    if actualOutput != expectedOutput {
        return fmt.Errorf("output mismatch:\nExpected:\n%s\nActual:\n%s", 
            expectedOutput, actualOutput)
    }
    
    return nil
}
```

---

## Phase 3: Safe OCI Spec Inspection

### 4. Enhanced Driver Logging - `milo/driver.go`
Add safe OCI spec logging without system modification:

```go
// In StartTask method, after container spec generation:
func (d *MiloDriverPlugin) StartTask(cfg *drivers.TaskConfig) (*drivers.TaskHandle, *drivers.DriverNetwork, error) {
    // ... existing code ...
    
    // Generate container spec
    spec, err := CreateContainerSpec(cfg, javaPath, artifactPath)
    if err != nil {
        return nil, nil, err
    }
    
    // Log OCI spec for BDD tests (test mode only)
    if cfg.TaskMeta["test_mode"] == "true" {
        specBytes, _ := json.MarshalIndent(spec, "", "  ")
        specPath := filepath.Join(cfg.TaskDir().LogDir, "oci-spec.json")
        ioutil.WriteFile(specPath, specBytes, 0644)
        d.logger.Debug("OCI spec logged for testing", "path", specPath)
    }
    
    // ... rest of existing code ...
}
```

### 5. `features/oci_validation.go` - New File
Implement safe container spec validation:

```go
package features

func (ctx *BDDTestContext) theContainerOCISpecShouldIncludeLinuxNamespaces() error {
    // Read the logged OCI spec
    allocs, _, err := ctx.nomadClient.Jobs().Allocations(ctx.currentJobID, false, nil)
    if err != nil || len(allocs) == 0 {
        return fmt.Errorf("no allocations found")
    }
    
    // Get allocation log directory
    allocInfo, _, err := ctx.nomadClient.Allocations().Info(allocs[0].ID, nil)
    if err != nil {
        return err
    }
    
    // Read OCI spec from log directory
    specPath := filepath.Join(ctx.nomadServer.GetDataDir(), "alloc", allocInfo.ID, "java-app", "logs", "oci-spec.json")
    specData, err := ioutil.ReadFile(specPath)
    if err != nil {
        return fmt.Errorf("OCI spec not found: %v", err)
    }
    
    var spec specs.Spec
    if err := json.Unmarshal(specData, &spec); err != nil {
        return err
    }
    
    // Validate required namespaces
    requiredNamespaces := []string{"pid", "ipc", "uts", "mount"}
    for _, reqNS := range requiredNamespaces {
        found := false
        for _, namespace := range spec.Linux.Namespaces {
            if string(namespace.Type) == reqNS {
                found = true
                break
            }
        }
        if !found {
            return fmt.Errorf("missing required namespace: %s", reqNS)
        }
    }
    
    return nil
}
```

---

## Phase 4: Enhanced Driver Integration

### 6. `milo/java.go` - Update for Controlled Testing
Modify Java detection to support test-controlled paths:

```go
// Add test mode support to FindJavaOnHost
func FindJavaOnHost(logger hclog.Logger, testPaths []string) (string, error) {
    // Use test paths if provided
    searchPaths := defaultJavaPaths
    if len(testPaths) > 0 {
        searchPaths = testPaths
    }
    
    // ... existing implementation ...
}

// Update driver to check for test paths
func (d *MiloDriverPlugin) StartTask(cfg *drivers.TaskConfig) (*drivers.TaskHandle, *drivers.DriverNetwork, error) {
    // ... existing code ...
    
    // Check for test Java paths in meta
    var testJavaPaths []string
    if testPathsStr := cfg.TaskMeta["test_java_paths"]; testPathsStr != "" {
        testJavaPaths = strings.Split(testPathsStr, ":")
    }
    
    javaPath, err := FindJavaOnHost(d.logger, testJavaPaths)
    if err != nil {
        return nil, nil, fmt.Errorf("Error: No Java runtime found on host. Please install Java to use Milo driver.")
    }
    
    // ... rest of existing code ...
}
```

---

## Phase 5: Error Message Alignment

### 7. Update Error Messages
Align error messages with Gherkin expectations:

**`milo/artifact.go`:**
```go
if !strings.HasSuffix(filename, ".jar") {
    return fmt.Errorf("Error: Artifact must be a .jar file, got: %s", filename)
}
```

**`milo/driver.go`:**
```go
if javaPath == "" {
    return fmt.Errorf("Error: No Java runtime found on host. Please install Java to use Milo driver.")
}
```

---

## Phase 6: Build & CI Integration

### 8. `Makefile` - Enhanced Test Targets
```makefile
.PHONY: test-bdd
test-bdd: build
	@echo "==> Running BDD acceptance tests..."
	go test -v ./features -tags=bdd,live_e2e -timeout 30m

.PHONY: test-bdd-focus
test-bdd-focus: build
	@echo "==> Running focused BDD test..."
	go test -v ./features -tags=bdd,live_e2e -timeout 30m -focus="$(SCENARIO)"

.PHONY: test-all
test-all: test test-acceptance test-bdd
	@echo "==> All tests completed"
```

---

## Implementation Timeline

**Week 1**: Core infrastructure (files 1-2)
**Week 2**: Step implementations and validations (files 3-5)  
**Week 3**: Driver enhancements and OCI validation (files 4-6)
**Week 4**: Integration testing and documentation

## Key Improvements from Review

1. **Safe OCI Inspection**: Uses driver logging instead of risky crun wrapper
2. **Controlled Java Testing**: Uses driver meta instead of global PATH manipulation  
3. **Cleaner godog Integration**: Method-based step definitions with proper context management
4. **Consolidated Files**: Reduced fragmentation while maintaining clarity
5. **Robust Error Handling**: Better state management and cleanup

## Success Criteria

✅ All Gherkin scenarios pass with real system integration  
✅ No system-level modifications required  
✅ Tests are reliable and safe for parallel execution  
✅ Exact error messages and outputs validated  
✅ Container specifications properly inspected  
✅ CI/CD pipeline ready