# Architecture Review Reference Guide: What to Look For and How to Fix It

## Core Architectural Principles

### The Fundamental Laws of Architecture
1. **Architecture is about intent, not frameworks** - The system should scream its purpose, not its implementation choices
2. **Dependencies point toward stability** - Volatile components depend on stable components, never the reverse
3. **Policy and details must be separated** - Business rules should know nothing about databases, UI, or frameworks
4. **Boundaries are drawn at axes of change** - Components that change for different reasons must be separated
5. **Testability is architectural** - If the system is hard to test, the architecture is wrong

## The Dependency Rule

The most important rule in clean architecture: **Source code dependencies must point only inward, toward higher-level policies.**

```
┌─────────────────────────────────────┐
│   Frameworks & Drivers (Blue)       │
│  ┌───────────────────────────────┐  │
│  │  Interface Adapters (Green)   │  │
│  │  ┌─────────────────────────┐  │  │
│  │  │  Application Business   │  │  │
│  │  │     Rules (Red)         │  │  │
│  │  │  ┌─────────────────┐    │  │  │
│  │  │  │ Enterprise Bus. │    │  │  │
│  │  │  │  Rules (Yellow) │    │  │  │
│  │  │  └─────────────────┘    │  │  │
│  │  └─────────────────────────┘  │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

Dependencies flow: Blue → Green → Red → Yellow (never the reverse)

## What to Look For: Architectural Smells

### 1. Layer Violations

**What to Look For:**
- Database code in the UI layer
- Business rules that know about web frameworks
- Use cases that directly manipulate UI elements
- SQL queries in domain entities
- Framework annotations in business logic
- HTTP concepts in the domain layer

**How to Fix:**
- **Introduce Interface Adapters**: Create translation layers between circles
- **Dependency Inversion**: Make inner layers define interfaces that outer layers implement
- **Extract Gateway Interfaces**: Define data access contracts in the business layer
- **Create Presenters**: Separate UI preparation from business logic

**Example Fix:**
```java
// Bad - Business rule knows about database
public class Order {
    public void save() {
        Database.execute("INSERT INTO orders...");
    }
}

// Good - Business rule defines interface
public class Order {
    private OrderRepository repository;
    
    public void save() {
        repository.save(this);
    }
}

// In infrastructure layer
public class SqlOrderRepository implements OrderRepository {
    public void save(Order order) {
        // SQL implementation details
    }
}
```

### 2. Component Coupling Problems

**What to Look For:**
- Cyclic dependencies between components
- Components that change together frequently
- "Junk drawer" components with no cohesion
- Components dependent on unstable components
- Abstract components depending on concrete ones

**How to Fix:**
- **Apply Acyclic Dependencies Principle**: Break cycles with dependency inversion
- **Common Closure Principle**: Group things that change together
- **Stable Dependencies Principle**: Depend in the direction of stability
- **Create Component Facades**: Reduce coupling surface area

**Metrics to Check:**
- **Instability (I)**: I = Fan-out / (Fan-in + Fan-out)
- **Abstractness (A)**: A = Abstract Classes / Total Classes
- **Distance from Main Sequence**: D = |A + I - 1|

### 3. Missing Architectural Boundaries

**What to Look For:**
- Business rules mixed with delivery mechanisms
- No clear separation between use cases
- Database schema driving domain model design
- UI frameworks dictating application flow
- External services called directly from business logic

**How to Fix:**
- **Draw Boundaries at Natural Seams**: Identify axes of change
- **Create Plugin Architecture**: Make volatile components plug into stable ones
- **Use Humble Objects**: Separate testable logic from hard-to-test boundaries
- **Implement Screaming Architecture**: Make intent visible in structure

**Example Fix:**
```python
# Bad - No boundary
class UserController:
    def register(self, request):
        user = User(request.form['email'])
        db.session.add(user)
        send_email(user.email, "Welcome!")
        return render_template('success.html')

# Good - Clear boundaries
class UserController:
    def __init__(self, register_user_use_case):
        self.use_case = register_user_use_case
    
    def register(self, request):
        result = self.use_case.execute(
            email=request.form['email']
        )
        return self.present(result)

class RegisterUserUseCase:
    def __init__(self, user_repo, notifier):
        self.user_repo = user_repo
        self.notifier = notifier
    
    def execute(self, email):
        user = User(email)
        self.user_repo.save(user)
        self.notifier.welcome(user)
        return user
```

### 4. Framework Coupling

**What to Look For:**
- Business logic inheriting from framework classes
- Framework annotations throughout domain code
- Inability to test without framework running
- Framework upgrade breaking business logic
- Core logic tied to specific framework lifecycle

**How to Fix:**
- **Keep Frameworks at Arm's Length**: Don't marry the framework
- **Use Adapters**: Translate between framework and your code
- **Main Partition**: Isolate framework configuration
- **Dependency Rule**: Frameworks are details, keep them in outer circle

### 5. Database-Centric Design

**What to Look For:**
- Entity classes mirroring database tables
- Business logic in stored procedures
- ORM driving domain model design
- Database transactions spanning use cases
- Performance concerns driving business rules

**How to Fix:**
- **Separate Entity from Data Model**: Don't let tables dictate objects
- **Repository Pattern**: Hide data access behind interfaces
- **Unit of Work**: Manage transactions at use case level
- **CQRS**: Separate read and write models when needed

### 6. Deployment Rigidity

**What to Look For:**
- Cannot deploy components independently
- Entire system must be rebuilt for small changes
- Test environment requires full production setup
- Cannot scale components individually
- Deployment coupled to specific infrastructure

**How to Fix:**
- **Component Independence**: Ensure independent deployability
- **Service Boundaries**: Draw boundaries that allow separate deployment
- **Configuration Isolation**: Externalize all configuration
- **Infrastructure Abstraction**: Don't depend on deployment details

### 7. Testability Issues

**What to Look For:**
- Need database to run unit tests
- Tests break when UI changes
- Cannot test business rules in isolation
- Test setup requires entire system
- Slow test suites due to infrastructure

**How to Fix:**
- **Design for Testability**: Make testing a primary concern
- **Test Doubles at Boundaries**: Use mocks/stubs at architectural boundaries
- **Humble Object Pattern**: Extract logic from hard-to-test components
- **Test API**: Design system with testing interface

## Component Design Principles

### Component Cohesion Principles
1. **REP (Reuse/Release Equivalence)**: Components are the unit of release
2. **CCP (Common Closure)**: Gather together things that change together
3. **CRP (Common Reuse)**: Don't force users to depend on things they don't use

### Component Coupling Principles
1. **ADP (Acyclic Dependencies)**: No cycles in the dependency graph
2. **SDP (Stable Dependencies)**: Depend in the direction of stability
3. **SAP (Stable Abstractions)**: Stable components should be abstract

## Architecture Evaluation Checklist

### System Intent
- [ ] Can you tell what the system does by its structure?
- [ ] Are use cases first-class citizens?
- [ ] Is the domain model independent of persistence?
- [ ] Could you swap frameworks without changing business logic?

### Dependency Management
- [ ] Do all dependencies point toward business rules?
- [ ] Are there any circular dependencies?
- [ ] Are stable components more abstract?
- [ ] Can you test business rules without infrastructure?

### Boundaries
- [ ] Are boundaries drawn at natural change points?
- [ ] Do components have single reasons to change?
- [ ] Is there clear separation between policy and mechanism?
- [ ] Are external systems behind interfaces?

### Evolution
- [ ] Can you add features without modifying existing code?
- [ ] Can components be deployed independently?
- [ ] Is the system open for extension, closed for modification?
- [ ] Can you defer implementation decisions?

## Red Flags Requiring Immediate Action

1. **Business Logic in the Database** - Extract immediately
2. **Circular Dependencies** - Break with dependency inversion
3. **Framework in the Core** - Push to outer layers
4. **No Tests for Business Rules** - Add before any changes
5. **UI Driving Architecture** - Invert dependencies
6. **Deployment Coupling** - Introduce boundaries

## The Architecture Improvement Process

1. **Map Current State** - Identify layers and dependencies
2. **Find Violations** - Look for dependency rule breaks
3. **Identify Seams** - Find natural boundary points
4. **Plan Transitions** - Design target architecture
5. **Refactor Incrementally** - Move toward target with tests
6. **Validate Improvements** - Ensure testability and flexibility

Remember: Good architecture maximizes the number of decisions NOT made. Every architectural decision should increase options, not decrease them.