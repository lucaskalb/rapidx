# State Machine Testing

RapidX now supports state machine testing using property-based testing. This functionality allows you to test systems with stateful behavior in a systematic and comprehensive way.

## Basic Concepts

### State Machine

A state machine is a computational model that describes how a system changes state in response to events (commands). In the context of testing, this allows us to:

- Test complex state transitions
- Validate system invariants
- Find bugs in specific operation sequences
- Ensure the system behaves correctly in all possible scenarios

### Main Components

#### StateMachine[S, C]

Defines a state machine with:
- `InitialState`: Initial state of the system
- `Commands`: List of available commands

```go
type StateMachine[S, C any] struct {
    InitialState S
    Commands     []Command[S, C]
}
```

#### Command[S, C]

Defines an individual command with:
- `Name`: Descriptive name for the command
- `Generator`: Generator that creates command instances
- `Execute`: Function that executes the command and returns the new state
- `Precondition`: Function that determines if the command can be executed
- `Postcondition`: Function that validates if the execution was correct

```go
type Command[S, C any] struct {
    Name         string
    Generator    gen.Generator[C]
    Execute      func(S, C) (S, error)
    Precondition func(S, C) bool
    Postcondition func(S, C, S) bool
}
```

## How to Use

### 1. Define the State

First, define the type that represents your system's state:

```go
type BankAccount struct {
    Balance int
    Closed  bool
}
```

### 2. Define Commands

Define the type that represents commands:

```go
type BankCommand struct {
    Type   string // "deposit", "withdraw", "close"
    Amount int
}
```

### 3. Implement Commands

Create commands with their respective implementations:

```go
depositCmd := Command[BankAccount, BankCommand]{
    Name: "deposit",
    Generator: gen.Map(gen.IntRange(1, 1000), func(amount int) BankCommand {
        return BankCommand{Type: "deposit", Amount: amount}
    }),
    Execute: func(state BankAccount, cmd BankCommand) (BankAccount, error) {
        if state.Closed {
            return state, errors.New("account is closed")
        }
        return BankAccount{
            Balance: state.Balance + cmd.Amount,
            Closed: state.Closed,
        }, nil
    },
    Precondition: func(state BankAccount, cmd BankCommand) bool {
        return !state.Closed
    },
    Postcondition: func(from BankAccount, cmd BankCommand, to BankAccount) bool {
        return to.Balance == from.Balance + cmd.Amount
    },
}
```

### 4. Create the State Machine

Combine the initial state with commands:

```go
sm := StateMachine[BankAccount, BankCommand]{
    InitialState: BankAccount{Balance: 0, Closed: false},
    Commands: []Command[BankAccount, BankCommand]{
        depositCmd,
        withdrawCmd,
        closeCmd,
    },
}
```

### 5. Run Tests

Use the `TestStateMachine` function to run tests:

```go
func TestBankAccount(t *testing.T) {
    cfg := prop.Config{
        Seed:        12345,
        Examples:    100,
        MaxShrink:   50,
        ShrinkStrat: "bfs",
        Parallelism: 1,
    }

    prop.TestStateMachine(t, sm, cfg)
}
```

## Complete Examples

### Bank Account

```go
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
                Postcondition: func(from BankAccount, cmd BankCommand, to BankAccount) bool {
                    return to.Balance == from.Balance + cmd.Amount
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
                Postcondition: func(from BankAccount, cmd BankCommand, to BankAccount) bool {
                    return to.Balance == from.Balance - cmd.Amount
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
                Postcondition: func(from BankAccount, cmd BankCommand, to BankAccount) bool {
                    return to.Closed && to.Balance == from.Balance
                },
            },
        },
    }

    prop.TestStateMachine(t, sm, prop.Default())
}
```

### Cache System

```go
type Cache struct {
    Data   map[string]string
    Size   int
    MaxSize int
}

type CacheCommand struct {
    Type  string
    Key   string
    Value string
}

func TestCache(t *testing.T) {
    sm := StateMachine[Cache, CacheCommand]{
        InitialState: Cache{Data: make(map[string]string), Size: 0, MaxSize: 100},
        Commands: []Command[Cache, CacheCommand]{
            {
                Name: "set",
                Generator: gen.Bind(
                    gen.StringAlphaNum(gen.Size{Min: 1, Max: 10}),
                    func(key string) gen.Generator[CacheCommand] {
                        return gen.Map(
                            gen.StringAlphaNum(gen.Size{Min: 1, Max: 20}),
                            func(value string) CacheCommand {
                                return CacheCommand{Type: "set", Key: key, Value: value}
                            },
                        )
                    },
                ),
                Execute: func(state Cache, cmd CacheCommand) (Cache, error) {
                    newState := state
                    if newState.Data == nil {
                        newState.Data = make(map[string]string)
                    }
                    newData := make(map[string]string)
                    for k, v := range state.Data {
                        newData[k] = v
                    }

                    _, exists := newData[cmd.Key]
                    newData[cmd.Key] = cmd.Value

                    newState.Data = newData
                    if !exists {
                        newState.Size++
                    }
                    return newState, nil
                },
                Precondition: func(state Cache, cmd CacheCommand) bool {
                    return state.Size < state.MaxSize || state.Data[cmd.Key] != ""
                },
                Postcondition: func(from Cache, cmd CacheCommand, to Cache) bool {
                    return to.Data[cmd.Key] == cmd.Value
                },
            },
            // ... other commands
        },
    }

    prop.TestStateMachine(t, sm, prop.Default())
}
```

## Advanced Features

### Shrinking

The shrinking system works automatically:

1. **Sequence Shrinking**: Removes commands from the sequence to find the smallest sequence that reproduces the error
2. **Command Shrinking**: Uses individual generator shrinkers to reduce command parameters
3. **Strategies**: Supports BFS (breadth-first) and DFS (depth-first) for shrinking

### Preconditions and Postconditions

- **Preconditions**: Commands that don't meet preconditions are automatically skipped
- **Postconditions**: If a postcondition fails, the test fails with detailed information
- **Validation**: The system automatically validates all postconditions after each execution

### Configuration

Use the `Config` structure to customize behavior:

```go
cfg := prop.Config{
    Seed:              12345,        // Seed for reproducibility
    Examples:          100,          // Number of sequences to test
    MaxShrink:         50,           // Maximum shrinking steps
    ShrinkStrat:       "bfs",        // Shrinking strategy
    StopOnFirstFailure: true,        // Stop on first error
    Parallelism:       1,            // Number of parallel workers
}
```

## Best Practices

### 1. Command Design

- **Atomic Commands**: Each command should represent an atomic operation
- **Clear Preconditions**: Define preconditions that reflect business rules
- **Precise Postconditions**: Validate exactly what should happen after execution

### 2. Data Generation

- **Realistic Generators**: Use generators that produce realistic data for your domain
- **Appropriate Ranges**: Define value ranges that make sense for the context
- **Combinators**: Use `gen.Bind`, `gen.Map` and other combinators to create complex generators

### 3. Error Handling

- **Expected Errors**: Return errors for conditions that should fail
- **Validation**: Use postconditions to validate that unexpected errors don't occur
- **Logging**: The system automatically logs execution history

### 4. Performance

- **Number of Examples**: Adjust the number of examples based on system complexity
- **Parallelism**: Use parallelism to speed up tests on systems with many examples
- **Shrinking**: Configure shrinking limits based on sequence complexity

## Common Use Cases

### 1. Database Systems

Test CRUD operations with constraints and transactions:

```go
// Test database operations
// - Insertion with unique keys
// - Updates with constraints
// - Deletion with references
// - Transactions with rollback
```

### 2. Cache Systems

Test cache policies and invalidation:

```go
// Test cache systems
// - LRU/LFU policies
// - TTL invalidation
// - Dependency invalidation
// - Size limits
```

### 3. Protocol State Machines

Test protocol implementations:

```go
// Test network protocols
// - Connection states
// - Handshakes
// - Timeouts and reconnections
// - Message sequences
```

### 4. Workflow Systems

Test complex workflows:

```go
// Test business workflows
// - Multi-step approvals
// - Branching conditions
// - Operation rollbacks
// - Notifications and alerts
```

## Troubleshooting

### Common Issues

1. **Commands Always Skipped**: Check if preconditions are too restrictive
2. **Postconditions Failing**: Make sure postconditions reflect actual behavior
3. **Slow Shrinking**: Reduce number of examples or shrinking limit
4. **Non-deterministic Tests**: Use fixed seeds for reproducibility

### Debugging

The system provides detailed information about failures:

- **Execution History**: Complete sequence of executed commands
- **Intermediate States**: States before and after each command
- **Skipped Commands**: List of commands that were skipped and why
- **Minimal Sequence**: Minimal sequence that reproduces the error after shrinking

## Integration with RapidX

The state machine system is fully integrated with RapidX:

- **Generators**: Uses all existing RapidX generators
- **Shrinking**: Integrated with RapidX's shrinking system
- **Configuration**: Uses the same configuration structure
- **Flags**: Supports all RapidX command-line flags

For more information about generators and other RapidX features, see the main project documentation.