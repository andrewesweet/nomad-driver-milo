# Code Review Reference Guide: What to Look For and How to Fix It

**Scope**: This guide focuses on code quality within a single module or small group of related modules. For system-wide architectural concerns, see the Architecture Review Reference Guide.

## Core Principles of Good Code

### The Fundamental Rules
1. **Code should clearly express intent** - A reader should understand WHAT the code does without studying HOW
2. **Each element should have one reason to change** - Classes, methods, and modules should have single, well-defined responsibilities
3. **Dependencies should flow in one direction** - High-level policies should not depend on low-level details
4. **Code without tests is broken code** - Untested code is legacy code by definition
5. **Duplication is the root of all evil** - Every piece of knowledge should have a single, authoritative representation

## What to Look For: Code Smells and Issues

### 1. Naming Problems

**What to Look For:**
- Names that require mental translation (`d`, `temp`, `data`, `obj`)
- Abbreviations and acronyms (`calcTotAmnt`, `procReq`)
- Generic names that don't express purpose (`manager`, `processor`, `helper`)
- Inconsistent naming across the codebase
- Names that lie about what the code actually does

**How to Fix:**
- **Rename Method/Variable/Class**: Choose names that reveal intent
- Names should be:
  - Searchable (avoid single letters except for loop counters)
  - Pronounceable (avoid `xsqrtn`)
  - Descriptive of purpose, not implementation
  - Consistent with domain language

**Example Fix:**
```python
# Bad
def calc(x, y):
    return x * 0.1 + y

# Good  
def calculate_total_with_tax(subtotal, tax_amount):
    return subtotal * TAX_RATE + tax_amount
```

### 2. Function/Method Issues

**What to Look For:**
- Functions longer than 20 lines
- Functions doing more than one thing
- Functions with more than 3 parameters
- Functions with flag arguments
- Functions with side effects hidden behind innocent names
- Functions at multiple levels of abstraction

**How to Fix:**
- **Extract Method**: Pull cohesive code into named functions
- **Replace Temp with Query**: Convert variables to methods
- **Introduce Parameter Object**: Group related parameters
- **Remove Flag Argument**: Create separate methods
- **Separate Query from Modifier**: Split functions that both return values and change state

**Example Fix:**
```java
// Bad
public void processOrder(Order order, boolean sendEmail, boolean updateInventory) {
    // validation logic
    // calculate totals
    // save to database
    // send email if flag
    // update inventory if flag
}

// Good
public void processOrder(Order order) {
    validateOrder(order);
    calculateOrderTotals(order);
    saveOrder(order);
}

public void sendOrderConfirmation(Order order) { ... }
public void updateInventoryForOrder(Order order) { ... }
```

### 3. Class Problems

**What to Look For:**
- Classes with 200+ lines
- Classes with more than one reason to change
- Classes that expose their internals
- Classes with too many instance variables (>7)
- Classes where methods use only some instance variables
- God classes that control everything

**How to Fix:**
- **Extract Class**: Split responsibilities into cohesive units
- **Extract Superclass/Subclass**: Create hierarchies for varying behavior
- **Replace Data Class with Object**: Add behavior to data holders
- **Hide Delegate**: Encapsulate relationships
- **Replace Type Code with Subclasses**: Use polymorphism over conditionals

**Example Fix:**
```python
# Bad
class Employee:
    def __init__(self):
        self.name = ""
        self.salary = 0
        self.street = ""
        self.city = ""
        self.state = ""
        self.manager_name = ""
        self.department = ""
        
    def get_full_address(self):
        return f"{self.street}, {self.city}, {self.state}"
        
    def calculate_pay(self):
        # Complex pay calculation
        pass

# Good
class Employee:
    def __init__(self, name, salary, address, department):
        self.name = name
        self.salary = salary
        self.address = address
        self.department = department

class Address:
    def __init__(self, street, city, state):
        self.street = street
        self.city = city
        self.state = state
        
    def get_full_address(self):
        return f"{self.street}, {self.city}, {self.state}"
```

### 4. Duplication

**What to Look For:**
- Exact code duplication
- Structural duplication (same pattern, different names)
- Algorithmic duplication
- Repeated switch/case statements
- Parallel inheritance hierarchies

**How to Fix:**
- **Extract Method**: For duplicate code blocks
- **Pull Up Method**: Move to common parent
- **Form Template Method**: Extract varying parts
- **Replace Conditional with Polymorphism**: For repeated switches
- **Extract Superclass**: For duplicate functionality

### 5. Dependencies and Coupling

**What to Look For:**
- Classes that know too much about other classes
- Long message chains (a.getB().getC().getD())
- Feature envy (method more interested in other class)
- Inappropriate intimacy between classes
- Divergent change (class changes for multiple reasons)
- Shotgun surgery (one change affects many classes)

**How to Fix:**
- **Extract Interface**: Define contracts, not implementations
- **Dependency Injection**: Pass dependencies, don't create them
- **Hide Delegate**: Encapsulate chains
- **Move Method/Field**: Put code with its data
- **Introduce Parameter Object**: Reduce coupling

**Example Fix:**
```csharp
// Bad - Direct dependency
public class OrderService {
    public void ProcessOrder(Order order) {
        var emailer = new EmailService(); // Direct creation
        emailer.SendConfirmation(order);
    }
}

// Good - Dependency injection
public class OrderService {
    private readonly IEmailService emailService;
    
    public OrderService(IEmailService emailService) {
        this.emailService = emailService;
    }
    
    public void ProcessOrder(Order order) {
        emailService.SendConfirmation(order);
    }
}
```

### 6. Legacy Code (Code Without Tests)

**What to Look For:**
- Any code without test coverage
- Code that's hard to test due to dependencies
- Hidden dependencies (statics, singletons, global state)
- Database or file system calls in business logic
- UI mixed with business logic
- Time dependencies (DateTime.Now)

**How to Fix:**
- **Write Characterization Tests**: Capture current behavior
- **Extract Interface**: Create seams for testing
- **Parameterize Constructor**: Inject dependencies
- **Extract and Override Call**: Create testing points
- **Introduce Sensing Variable**: Make hidden behavior observable
- **Sprout Method/Class**: Add new tested code alongside old

**Example Fix:**
```java
// Bad - Hard to test
public class PaymentProcessor {
    public void processPayment(Payment payment) {
        Database db = Database.getInstance();
        PaymentGateway gateway = new PaymentGateway();
        
        db.save(payment);
        gateway.charge(payment);
        EmailSender.send("Payment processed");
    }
}

// Good - Testable with seams
public class PaymentProcessor {
    private final PaymentRepository repository;
    private final PaymentGateway gateway;
    private final NotificationService notifier;
    
    public PaymentProcessor(
        PaymentRepository repository,
        PaymentGateway gateway,
        NotificationService notifier) {
        this.repository = repository;
        this.gateway = gateway;
        this.notifier = notifier;
    }
    
    public void processPayment(Payment payment) {
        repository.save(payment);
        gateway.charge(payment);
        notifier.notifyPaymentProcessed(payment);
    }
}
```

### 7. Error Handling

**What to Look For:**
- Returning null instead of throwing exceptions
- Empty catch blocks
- Using exceptions for control flow
- Catching Exception (too broad)
- Error codes instead of exceptions
- Not cleaning up resources

**How to Fix:**
- **Replace Error Code with Exception**: Use type system
- **Introduce Special Case/Null Object**: Eliminate null checks
- **Extract Try/Catch Blocks**: Separate error handling
- Use try-with-resources or using statements

### 8. Comments

**What to Look For:**
- Comments explaining WHAT instead of WHY
- Commented-out code
- Redundant comments
- Comments used instead of good names
- Out-of-date comments
- TODO comments that never get done

**How to Fix:**
- **Extract Method with Descriptive Name**: Replace comment with code
- **Rename to Express Intent**: Make comment unnecessary
- Delete commented code (version control has it)
- Keep only comments that explain WHY or provide context

## Quick Reference: Refactoring Priority

### Must Fix Immediately
1. **No Tests** → Write characterization tests
2. **Security Vulnerabilities** → Fix with tested solution
3. **Functions > 50 lines** → Extract methods
4. **Circular Dependencies** → Introduce interfaces
5. **Exposed Internals** → Encapsulate

### Should Fix Soon
1. **Duplicate Code** → Extract common functionality
2. **Long Parameter Lists** → Introduce parameter objects
3. **Feature Envy** → Move method to proper class
4. **Mixed Abstraction Levels** → Extract to consistent level
5. **Poor Names** → Rename for clarity

### Consider Fixing
1. **Large Classes** → Extract if cohesion is low
2. **Comments** → Replace with better code
3. **Complex Conditionals** → Consider polymorphism
4. **Data Classes** → Add behavior if appropriate
5. **Speculative Generality** → Simplify if unused

## The Refactoring Process

1. **Ensure tests exist** - Never refactor without tests
2. **Make the change** - Small, incremental steps
3. **Run tests** - After every small change
4. **Commit** - When tests pass
5. **Repeat** - Until code expresses intent clearly

Remember: The goal is code that clearly expresses intent, is easy to change, and is proven correct through tests. Every refactoring should move toward these goals.

## When to Escalate to Architecture Review

Some issues discovered during code review indicate system-wide problems requiring architectural attention:

1. **Layer Violations** - Business logic knowing about databases or UI
2. **Widespread Duplication** - Same pattern across many modules
3. **Dependency Cycles** - Circular dependencies between packages
4. **Framework Bleeding** - Framework details throughout the codebase
5. **Deployment Coupling** - Cannot deploy module independently
6. **Cross-Cutting Concerns** - Security, logging inconsistencies across system

When these are found, document them for architectural review rather than attempting local fixes.