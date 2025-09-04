# ADR-003: Replay Functionality via Command Line

## Status
Accepted

## Context

Property-based testing generates random test cases, which means that test failures are often non-deterministic and difficult to reproduce. When a property-based test fails, developers need a reliable way to reproduce the exact same failure for debugging purposes.

### Key Requirements
- Reproduce exact test failures for debugging
- Provide deterministic test execution
- Enable easy debugging workflow
- Maintain compatibility with Go's testing framework
- Support both individual test and full test suite execution

### Problem Statement
Without replay functionality:
- Developers cannot reliably reproduce failing test cases
- Debugging becomes difficult and time-consuming
- CI/CD failures are hard to investigate
- Test flakiness cannot be properly diagnosed

## Decision

**We will implement replay functionality that allows developers to reproduce exact test failures using command-line flags, specifically the seed value that generated the failing test case.**

### Implementation Details

#### 1. Seed Capture and Reporting
When a property-based test fails, RapidX captures and reports:
- The random seed used for test generation
- The number of examples run before failure
- The number of shrinking steps performed
- The minimal counterexample found
- A complete replay command

#### 2. Replay Command Generation
The system generates a complete `go test` command that can be used to reproduce the failure:

```go
full := fmt.Sprintf("^%s$/%s(/|$)", t.Name(), name)
t.Fatalf("[rapidx] property failed; seed=%d; examples_run=%d; shrunk_steps=%d\n"+
    "counterexample (min): %#v\nreplay: go test -run '%s' -rapidx.seed=%d",
    seed, i+1, steps, min, full, seed)
```

#### 3. Command-Line Integration
The replay functionality integrates with Go's standard testing flags:
- `-run` flag to specify which test to run
- `-rapidx.seed` flag to set the random seed
- Standard Go test flags for additional control

### Example Usage

#### Test Failure Output
```
[rapidx] property failed; seed=12345; examples_run=42; shrunk_steps=15
counterexample (min): [1, 2, 3]
replay: go test -run '^TestMyProperty$/ex#l2(/|$)' -rapidx.seed=12345
```

#### Replay Commands
```bash
# Reproduce the exact failure
go test -run '^TestMyProperty$/ex#l2(/|$)' -rapidx.seed=12345

# Run with verbose output for debugging
go test -run '^TestMyProperty$/ex#l2(/|$)' -rapidx.seed=12345 -v

# Run with additional Go test flags
go test -run '^TestMyProperty$/ex#l2(/|$)' -rapidx.seed=12345 -count=1 -failfast
```

## Consequences

### Positive
- **Deterministic Reproduction**: Exact same test case is generated every time
- **Easy Debugging**: Simple copy-paste workflow for reproducing failures
- **CI/CD Integration**: Failed builds can be reproduced locally
- **Developer Experience**: Clear, actionable error messages
- **Standard Tooling**: Uses familiar Go testing commands
- **Flexibility**: Works with all Go test flags and options

### Negative
- **Seed Dependency**: Replay only works if the same seed produces the same failure
- **Generator Changes**: Changes to generators may break replay functionality
- **Long Commands**: Generated replay commands can be quite long
- **Manual Process**: Requires manual copy-paste of commands

### Neutral
- **Storage**: Seeds are just integers, minimal storage overhead
- **Performance**: No performance impact on normal test execution
- **Compatibility**: Works with existing Go testing infrastructure

## Alternatives Considered

### 1. Test Case Serialization
**Approach**: Serialize and store the exact failing test case.

**Rejected because**:
- Complex serialization for all data types
- Storage overhead for test cases
- Difficult to handle complex nested structures
- Version compatibility issues

### 2. Test Case Logging
**Approach**: Log all generated test cases to files.

**Rejected because**:
- Massive log files for large test suites
- Performance impact on test execution
- Difficult to find specific failing cases
- Storage and cleanup concerns

### 3. Interactive Debugging
**Approach**: Provide an interactive debugging mode.

**Rejected because**:
- Complex implementation
- Not suitable for CI/CD environments
- Requires additional tooling
- Doesn't integrate well with standard Go workflow

### 4. Test Case Replay Files
**Approach**: Generate separate replay files for each failure.

**Rejected because**:
- File management complexity
- Version control pollution
- Difficult to clean up old replay files
- Additional tooling required

## Implementation Notes

### Seed Management
- Seeds are captured at the beginning of test execution
- Seeds are passed to all generators to ensure reproducibility
- Seed 0 triggers random seed generation based on current time

### Error Message Format
The error message follows a consistent format:
```
[rapidx] property failed; seed=<seed>; examples_run=<count>; shrunk_steps=<steps>
counterexample (min): <value>
replay: go test -run '<pattern>' -rapidx.seed=<seed>
```

### Integration with Go Testing
- Uses standard `go test` command structure
- Compatible with all Go testing flags (`-v`, `-count`, `-failfast`, etc.)
- Works with test patterns and package selection
- Integrates with IDE test runners

### Edge Cases
- **Empty test names**: Handled with proper escaping
- **Special characters**: Properly escaped in test patterns
- **Multiple failures**: Each failure gets its own replay command
- **Parallel execution**: Each worker reports its own seed and replay command

## Examples

### Basic Replay
```bash
# Original test run (fails)
go test ./...

# Output:
# [rapidx] property failed; seed=12345; examples_run=42; shrunk_steps=15
# counterexample (min): [1, 2, 3]
# replay: go test -run 'TestMyProperty' -rapidx.seed=12345

# Reproduce the failure
go test -run 'TestMyProperty' -rapidx.seed=12345
```

### Advanced Replay
```bash
# Reproduce with verbose output
go test -run 'TestMyProperty' -rapidx.seed=12345 -v

# Reproduce with specific package
go test -run 'TestMyProperty' -rapidx.seed=12345 ./mypackage

# Reproduce with additional flags
go test -run 'TestMyProperty' -rapidx.seed=12345 -count=1 -failfast -race
```

### CI/CD Integration
```bash
# In CI, capture the replay command from logs
REPLAY_CMD=$(grep "replay:" test_output.log | cut -d' ' -f2-)

# Execute the replay command
$REPLAY_CMD
```

## Future Considerations

### Potential Enhancements
- **Test Case Export**: Export failing test cases to files for external analysis
- **Replay History**: Maintain a history of recent failures and their replay commands
- **IDE Integration**: Direct integration with IDE debuggers
- **Automated Replay**: Automatic replay of failures in CI/CD pipelines

### Compatibility Considerations
- **Generator Changes**: Document when generator changes break replay compatibility
- **Version Migration**: Provide migration tools for replay commands across versions
- **Cross-Platform**: Ensure replay commands work across different operating systems

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [QuickCheck: Replay](https://www.cse.chalmers.se/~rjmh/QuickCheck/manual.html#replay)
- [Hypothesis: Reproducing Failures](https://hypothesis.readthedocs.io/en/latest/reproducing.html)

## Related ADRs

- ADR-001: Serial Shrinking in Parallel Test Execution
- ADR-002: Shrinking StrategTestMyPropertyray Selection (BFS vs DFS)
