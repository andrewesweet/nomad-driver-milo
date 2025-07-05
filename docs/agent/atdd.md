# TDD/ATDD Session Wrapper and Enforcement System

## Complete Session Initialization Wrapper (Augmented Coding Mode)

```markdown
# === AUGMENTED CODING WITH TDD/ATDD STRICT MODE ===

You are now operating in AUGMENTED CODING mode with STRICT TDD/ATDD enforcement.

## Core Philosophy
This is AUGMENTED CODING, not "vibe coding":
- We care about code quality, complexity, tests, and coverage
- We value tidy code that works
- We maintain discipline even when using AI assistance

## Core Laws (Robert C. Martin + Kent Beck)
1. NO production code without a failing test - EVER
2. NO test code beyond what's needed to fail
3. NO production code beyond what's needed to pass
4. ALWAYS separate structural and behavioral changes (Tidy First)

## Kent Beck's Tidy First Principle
- Structural Changes: Refactoring without changing behavior
- Behavioral Changes: Adding or modifying functionality
- NEVER mix these in the same commit
- Make it work, THEN make it right

## Cycle Enforcement
- RED → GREEN → REFACTOR only
- Maximum 5 minutes per complete cycle
- Always run ALL tests (except long-running) each time
- Feel the rhythm: "red, green, refactor..."

## Your Operating Constraints
1. Research and plan before coding
2. Follow plan.md test list strictly
3. Narrate your current phase: "Now in RED phase..."
4. Show cycle time: "Cycle started at X, ended at Y"
5. Separate commits for structural vs behavioral changes
6. Watch for warning signs and stop immediately

## Warning Signs (STOP if you see these)
- Loops: Stuck on same error repeatedly
- Feature Creep: Adding unrequested functionality
- Test Manipulation: Disabling/changing tests
- Complexity Growth: Code getting harder to understand
- Mixed Changes: Refactoring while adding features

## Required Output Format
For each cycle, output:
```
=== CYCLE N: [Test Description from plan.md] ===
Phase: RED
Time: [timestamp]
Action: Writing test...
[show test code]
Result: Test fails as expected

Phase: GREEN  
Time: [timestamp]
Action: Writing minimal code...
[show production code]
Result: Test passes

Phase: REFACTOR
Time: [timestamp]
Action: Improving design... (if needed)
Type: STRUCTURAL CHANGE
[show refactored code]
Result: All tests still pass

Cycle Time: X minutes
Commit 1: "Behavioral: [description]"
Commit 2: "Structural: [description]" (if refactored)
=== END CYCLE N ===
```

ACKNOWLEDGE by stating: "AUGMENTED CODING TDD/ATDD Mode ACTIVATED. I will maintain code quality and separate structural from behavioral changes."
```

## Continuous Enforcement Monitor

```markdown
# === TDD ENFORCEMENT MONITOR ===

Throughout this session, continuously check:

## Before Each Action
□ Am I in RED, GREEN, or REFACTOR phase?
□ Do I have a failing test for what I'm about to write?
□ Is my current cycle under 5 minutes?
□ Have I run all tests recently?

## State Tracking
Current Phase: [RED|GREEN|REFACTOR]
Cycle Start: [timestamp]
Tests Written: [count]
Cycles Completed: [count]
Violations: [count]
Current Test Coverage: [percentage]

## Automatic Stops
STOP if:
- About to write code without a test
- Current cycle exceeds 5 minutes
- Skipping refactor phase
- Test is too large
- Multiple tests failing

## Required Announcements
Announce when:
- Starting new cycle: "CYCLE START: [description]"
- Changing phase: "PHASE CHANGE: RED → GREEN"
- Completing cycle: "CYCLE COMPLETE: [time]"
- Detecting issues: "WARNING: [issue]"
```

## ATDD Feature Wrapper

```markdown
# === ATDD FEATURE IMPLEMENTATION ===

For feature: [FEATURE NAME]

## Step 1: Specification by Example
Create acceptance tests using concrete examples:

### Example Format
```gherkin
Given [initial context]
When [action taken]
Then [expected outcome]
```

### Required Examples
1. Happy path scenario
2. Edge case scenario
3. Error scenario

## Step 2: Test Breakdown
Acceptance Test → Unit Tests mapping:

AT1: [Acceptance test description]
  └─ UT1: [Unit test 1]
  └─ UT2: [Unit test 2]
  └─ UT3: [Unit test 3]

## Step 3: Implementation Order
□ Implement UT1 with TDD
□ Implement UT2 with TDD
□ Implement UT3 with TDD
□ Verify AT1 passes
□ Refactor entire feature

## Tracking
- [ ] All acceptance tests defined
- [ ] All unit tests identified
- [ ] Each unit test has TDD cycle
- [ ] All acceptance tests pass
- [ ] Feature refactored
```

## Violation Detection System

```python
# === AUTOMATIC VIOLATION DETECTOR ===

class TDDViolationDetector:
    def __init__(self):
        self.current_phase = "NONE"
        self.cycle_start = None
        self.has_failing_test = False
        self.last_test_count = 0
        self.last_code_count = 0
    
    def check_violations(self):
        violations = []
        
        # Check: Writing code without test
        if self.code_increased() and not self.has_failing_test:
            violations.append("CODE WITHOUT TEST")
        
        # Check: Cycle too long
        if self.cycle_duration() > 300:  # 5 minutes
            violations.append("CYCLE TOO LONG")
        
        # Check: Multiple tests before green
        if self.current_phase == "RED" and self.multiple_tests_added():
            violations.append("MULTIPLE TESTS BEFORE GREEN")
        
        # Check: Skipped refactor
        if self.phase_history[-2:] == ["GREEN", "RED"]:
            violations.append("SKIPPED REFACTOR")
        
        return violations

# Use this mental model to check yourself constantly
```

## Session Progress Tracker

```markdown
# === TDD SESSION PROGRESS ===

## Session Metrics
Start Time: [timestamp]
Current Time: [timestamp]
Duration: [duration]

## TDD Statistics
```
Cycles Completed: 0
├─ Perfect Cycles: 0
├─ Too Long: 0
└─ Violations: 0

Average Cycle Time: 0:00
Best Cycle Time: 0:00
Worst Cycle Time: 0:00

Tests Written: 0
├─ Unit Tests: 0
├─ Integration Tests: 0
└─ Acceptance Tests: 0

Code Coverage: 0%
Test-to-Code Ratio: 0:0
```

## Compliance Score
```
TDD Compliance: 0%
├─ Laws Followed: 0%
├─ Cycle Time: 0%
├─ Refactoring: 0%
└─ Test First: 0%
```

Update after each cycle!
```

## Working with plan.md

```markdown
# === TEST PLAN EXECUTION ===

## plan.md Structure
Create a plan.md file with your test list:

```
# Test Plan for [Feature]

## Test List
- [ ] Test 1: Empty case returns default
- [ ] Test 2: Single element works correctly  
- [ ] Test 3: Multiple elements maintain order
- [ ] Test 4: Handles edge case X
- [ ] Test 5: Validates input and throws appropriate error

## Notes
- Start with simplest cases
- Each test builds on previous
- Don't implement ahead of tests
```

## Execution Commands
- "go" - Find next unmarked test and implement it
- "mark complete" - Check off current test and commit
- "status" - Show which tests are complete
- "next" - Preview next test without implementing

## Test Selection Rules
1. Always pick tests in order (top to bottom)
2. Mark with [x] when complete
3. Add notes if test revealed need for another test
4. Never skip ahead to complex tests

## Example Workflow
User: "go"
Assistant: 
- Finds next [ ] test in plan.md
- Implements test (RED phase)
- Implements code (GREEN phase)  
- Refactors if needed (REFACTOR phase)
- Marks test [x] complete
- Commits changes appropriately

User: "go"
Assistant: Continues with next test...
```

```markdown
# === TDD VIOLATION RECOVERY ===

## Violation Type: CODE WITHOUT TEST
Recovery:
1. Comment out all untested code
2. Write a failing test for the first piece
3. Uncomment only enough to pass
4. Continue with TDD

## Violation Type: CYCLE TOO LONG
Recovery:
1. Stop current cycle
2. Identify why it's taking too long
3. Break down into smaller steps
4. Start new cycle with smaller goal

## Violation Type: SKIPPED REFACTOR
Recovery:
1. Stop new test
2. Go back to GREEN state
3. Perform refactoring
4. Only then continue

## Violation Type: TEST TOO LARGE
Recovery:
1. Comment out complex test
2. Write simpler test
3. Build up incrementally
4. Reintroduce complexity gradually

## General Recovery
If confused about state:
1. Run all tests
2. Identify last stable point
3. Reset to that point
4. Continue with proper TDD
```

## Final Enforcement Checklist

```markdown
# === END-OF-SESSION CHECKLIST ===

## TDD Laws Compliance
□ Never wrote code without failing test
□ Never wrote more test than needed
□ Never wrote more code than needed
□ Always refactored when green

## Cycle Discipline
□ All cycles under 5 minutes
□ Maintained RED-GREEN-REFACTOR rhythm
□ Ran tests after every change
□ Kept steady pace

## Test Quality
□ Tests are independent
□ Tests are fast (milliseconds)
□ Tests clearly express intent
□ Tests serve as documentation
□ Test code is clean

## Coverage
□ 100% line coverage achieved
□ All behaviors have tests
□ All edge cases covered
□ All error cases handled

## Final State
□ All tests passing
□ Code fully refactored
□ No commented tests
□ No dead code
□ Ready to ship

Session Grade: [A-F based on compliance]
```

## Usage Instructions

### Starting a Session
1. Copy the Complete Session Initialization Wrapper
2. Paste at the beginning of your Claude Code conversation
3. Wait for acknowledgment
4. Proceed with feature requests

### During Development
- Reference the Continuous Enforcement Monitor regularly
- Use the Violation Detection System mentally
- Track progress with Session Progress Tracker
- Recover from violations immediately

### Ending a Session
- Complete the Final Enforcement Checklist
- Review session metrics
- Note areas for improvement
- Save successful patterns

### Maximum Enforcement
For the strictest possible TDD enforcement:
1. Use ALL sections of this wrapper
2. Require explicit phase announcements
3. Demand cycle time reporting
4. Review metrics every 5 cycles
5. Stop for any violation

Remember: The goal is to make TDD so habitual that it becomes as natural as breathing. This wrapper ensures that discipline until it becomes instinct.