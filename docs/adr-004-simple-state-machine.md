# ADR-004: Simple State Machine Testing

## Status
Accepted

## Context

Property-based testing is excellent for testing individual functions and data transformations, but many real-world systems involve stateful behavior that changes over time. Current property-based testing libraries often lack built-in support for testing state machines, forcing developers to implement complex state management manually.

### Key Requirements
- Test stateful systems with property-based testing
- Provide simple, intuitive API for state machine testing
- Support common state machine patterns
- Maintain compatibility with existing RapidX generators
- Enable shrinking of state machine sequences
- Support both deterministic and non-deterministic state machines

### Problem Statement
Without state machine support:
- Developers must manually implement state management in tests
- Complex state transitions are difficult to test comprehensively
- State machine bugs are hard to reproduce and debug
- No standard way to shrink failing state sequences
- Testing concurrent state machines is particularly challenging

### Use Cases
- **Database Operations**: Testing CRUD operations with constraints
- **Game Logic**: Testing game state transitions and rules
- **Protocol Implementation**: Testing network protocol state machines
- **Business Logic**: Testing workflow and approval processes
- **Cache Systems**: Testing cache invalidation and updates
- **Finite State Machines**: Testing any system with discrete states

## Decision

**We will implement a simple state machine testing framework that allows developers to define state machines and test their behavior using property-based testing.**

### Proposed API Design

#### 1. State Machine Definition
```go
type StateMachine[S, C any] struct {
    InitialState S
    Commands     []Command[S, C]
}

type Command[S, C any] struct {
    Name        string
    Generator   gen.Generator[C]
    Execute     func(S, C) (S, error)
    Precondition func(S, C) bool
    Postcondition func(S, C, S) bool
}
```

#### 2. State Machine Testing
```go
func TestStateMachine[S, C any](t *testing.T, sm StateMachine[S, C], cfg prop.Config)
```

#### 3. Example Usage
```go
type BankAccount struct {
    Balance int
    Closed  bool
}

type BankCommand struct {
    Type   string // "deposit", "withdraw", "close"
    Amount int
}

func TestBankAccount(t *testing.T) {
    sm := StateMachine[BankAccount, BankCommand]{
        InitialState: BankAccount{Balance: 0, Closed: false},
        Commands: []Command[BankAccount, BankCommand]{
            {
                Name: "deposit",
                Generator: gen.Map(gen.IntRange(1, 1000), func(amount int) BankCommand {
                    return BankCommand{Type: "deposit", Amount: amount}
                }),
                Execute: func(state BankAccount, cmd BankCommand) (BankAccount, error) {
                    if state.Closed {
                        return state, errors.New("account is closed")
                    }
                    return BankAccount{Balance: state.Balance + cmd.Amount, Closed: state.Closed}, nil
                },
                Precondition: func(state BankAccount, cmd BankCommand) bool {
                    return !state.Closed
                },
            },
            {
                Name: "withdraw",
                Generator: gen.Map(gen.IntRange(1, 1000), func(amount int) BankCommand {
                    return BankCommand{Type: "withdraw", Amount: amount}
                }),
                Execute: func(state BankAccount, cmd BankCommand) (BankAccount, error) {
                    if state.Closed {
                        return state, errors.New("account is closed")
                    }
                    if state.Balance < cmd.Amount {
                        return state, errors.New("insufficient funds")
                    }
                    return BankAccount{Balance: state.Balance - cmd.Amount, Closed: state.Closed}, nil
                },
                Precondition: func(state BankAccount, cmd BankCommand) bool {
                    return !state.Closed && state.Balance >= cmd.Amount
                },
            },
            {
                Name: "close",
                Generator: gen.Const(BankCommand{Type: "close", Amount: 0}),
                Execute: func(state BankAccount, cmd BankCommand) (BankAccount, error) {
                    return BankAccount{Balance: state.Balance, Closed: true}, nil
                },
                Precondition: func(state BankAccount, cmd BankCommand) bool {
                    return !state.Closed
                },
            },
        },
    }
    
    prop.TestStateMachine(t, sm, prop.Default())
}
```

### Implementation Details

#### 1. Command Sequence Generation
- Generate sequences of commands using existing RapidX generators
- Support weighted command selection for realistic test scenarios
- Allow custom sequence length control

#### 2. State Machine Execution
- Execute commands in sequence, maintaining state
- Skip commands that don't meet preconditions
- Track execution history for debugging

#### 3. Shrinking Support
- Shrink command sequences by removing commands
- Shrink individual commands using their generators
- Maintain state machine invariants during shrinking

#### 4. Property Validation
- Validate state invariants after each command
- Check postconditions for each command execution
- Support custom property functions

## Consequences

### Positive
- **Comprehensive Testing**: Test complex stateful behavior systematically
- **Bug Discovery**: Find edge cases in state transitions
- **Reproducible Failures**: Shrinking provides minimal failing sequences
- **Intuitive API**: Simple, declarative state machine definition
- **Integration**: Works seamlessly with existing RapidX generators
- **Flexibility**: Support various state machine patterns

### Negative
- **Implementation Complexity**: Significant development effort required
- **API Surface**: Adds complexity to the library
- **Performance**: State machine execution may be slower than simple property tests
- **Learning Curve**: Developers need to understand state machine concepts
- **Debugging**: State machine failures may be harder to understand

### Neutral
- **Memory Usage**: Additional memory for state tracking and history
- **Documentation**: Requires comprehensive documentation and examples
- **Testing**: Need extensive testing of the state machine framework itself

## Alternatives Considered

### 1. External Library Integration
**Approach**: Integrate with existing state machine libraries.

**Rejected because**:
- Limited control over shrinking behavior
- Potential API conflicts
- Additional dependencies
- Less integration with RapidX ecosystem

### 2. Manual State Management
**Approach**: Provide utilities for manual state management in tests.

**Rejected because**:
- High complexity for developers
- Inconsistent patterns across codebases
- Difficult to shrink state sequences
- No standard approach

### 3. Model-Based Testing
**Approach**: Implement full model-based testing framework.

**Rejected because**:
- Excessive complexity for most use cases
- Steep learning curve
- Overkill for simple state machines
- Difficult to implement shrinking

### 4. State Machine DSL
**Approach**: Create a domain-specific language for state machines.

**Rejected because**:
- Complex parser and compiler
- Limited Go integration
- Difficult to debug
- Over-engineering for the problem

## Implementation Plan

### Phase 1: Core Framework
- Implement basic `StateMachine` and `Command` types
- Create command sequence generation
- Implement state machine execution engine
- Add basic property validation

### Phase 2: Shrinking Support
- Implement command sequence shrinking
- Add command-level shrinking
- Ensure state machine invariants during shrinking
- Optimize shrinking performance

### Phase 3: Advanced Features
- Support for concurrent state machines
- Advanced command selection strategies
- State machine visualization tools
- Performance optimizations

### Phase 4: Documentation and Examples
- Comprehensive documentation
- Real-world examples
- Best practices guide
- Migration guide from manual approaches

## Open Questions

### 1. Shrinking Strategy
- How should we shrink command sequences?
- Should we prioritize removing commands or modifying them?
- How do we handle state-dependent shrinking?

### 2. Error Handling
- How should we handle command execution errors?
- Should errors be part of the shrinking process?
- How do we report state machine failures?

### 3. Performance
- What are the performance implications of state tracking?
- How do we optimize for large state machines?
- Should we support parallel state machine execution?

### 4. API Design
- Is the proposed API intuitive enough?
- Should we support more advanced state machine features?
- How do we handle state machine composition?

## Examples

### Simple Counter State Machine
```go
type Counter struct {
    Value int
}

type CounterCommand struct {
    Type  string // "increment", "decrement", "reset"
    Delta int
}

func TestCounter(t *testing.T) {
    sm := StateMachine[Counter, CounterCommand]{
        InitialState: Counter{Value: 0},
        Commands: []Command[Counter, CounterCommand]{
            {
                Name: "increment",
                Generator: gen.Map(gen.IntRange(1, 10), func(delta int) CounterCommand {
                    return CounterCommand{Type: "increment", Delta: delta}
                }),
                Execute: func(state Counter, cmd CounterCommand) (Counter, error) {
                    return Counter{Value: state.Value + cmd.Delta}, nil
                },
            },
            // ... other commands
        },
    }
    
    prop.TestStateMachine(t, sm, prop.Default())
}
```

### Cache State Machine
```go
type Cache struct {
    Data map[string]string
    Size int
    MaxSize int
}

type CacheCommand struct {
    Type string // "get", "set", "delete", "clear"
    Key  string
    Value string
}

func TestCache(t *testing.T) {
    sm := StateMachine[Cache, CacheCommand]{
        InitialState: Cache{Data: make(map[string]string), Size: 0, MaxSize: 100},
        Commands: []Command[Cache, CacheCommand]{
            {
                Name: "set",
                Generator: gen.Map2(
                    gen.StringAlphaNum(gen.Size{Min: 1, Max: 10}),
                    gen.StringAlphaNum(gen.Size{Min: 1, Max: 20}),
                    func(key, value string) CacheCommand {
                        return CacheCommand{Type: "set", Key: key, Value: value}
                    },
                ),
                Execute: func(state Cache, cmd CacheCommand) (Cache, error) {
                    newState := state
                    if _, exists := newState.Data[cmd.Key]; !exists {
                        newState.Size++
                    }
                    newState.Data[cmd.Key] = cmd.Value
                    return newState, nil
                },
                Precondition: func(state Cache, cmd CacheCommand) bool {
                    return state.Size < state.MaxSize || state.Data[cmd.Key] != ""
                },
            },
            // ... other commands
        },
    }
    
    prop.TestStateMachine(t, sm, prop.Default())
}
```

## References

- [QuickCheck: State Machine Testing](https://www.cse.chalmers.se/~rjmh/QuickCheck/manual.html#state)
- [Hypothesis: Stateful Testing](https://hypothesis.readthedocs.io/en/latest/stateful.html)
- [Model-Based Testing](https://en.wikipedia.org/wiki/Model-based_testing)
- [Finite State Machines](https://en.wikipedia.org/wiki/Finite-state_machine)

## Related ADRs

- ADR-001: Serial Shrinking in Parallel Test Execution
- ADR-002: Shrinking Strategy Selection (BFS vs DFS)
- ADR-003: Replay Functionality via Command Line