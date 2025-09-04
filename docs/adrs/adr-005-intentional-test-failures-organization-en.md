# ADR-005: Organization of Tests with Intentional Failures

## Status

**ACCEPTED** - 2024-12-19

## Context

The rapidx project contains various types of tests, including tests that are designed to fail intentionally. These tests serve different purposes:

1. **Demonstration Tests**: Show how the framework works when properties fail
2. **Framework Tests**: Verify the correct behavior of the framework in failure scenarios
3. **Comparison Tests**: Demonstrate expected failures in comparison functions

Currently, these tests are scattered across different directories:
- `examples/` - contains demonstration tests mixed with functional examples
- `prop/prop_test.go` - contains framework tests with intentional failures
- `quick/quick_test.go` - contains comparison tests with expected failures

This organization creates confusion about which tests should pass and which are designed to fail, making maintenance and selective test execution difficult.

## Decision

Create a dedicated structure for tests with intentional failures, organizing them into specific subpackages:

```
testfailures/
├── demo/              # Demonstration tests
│   ├── shrinking_demo_test.go
│   ├── property_demo_test.go
│   └── comparison_demo_test.go
├── framework/         # Framework functionality tests
│   ├── failure_behavior_test.go
│   ├── shrinking_failure_test.go
│   └── parallel_failure_test.go
└── integration/       # Integration tests with failures
    └── end_to_end_failure_test.go
```

### Implementation Characteristics:

1. **Build Tags**: All tests with intentional failures use `//go:build demo` to allow selective execution
2. **Purpose Separation**: Each subpackage has a specific and well-documented purpose
3. **Clear Documentation**: Each file contains comments explaining the purpose of the tests
4. **Functionality Preservation**: Original tests are moved, not removed

## Consequences

### Positive:

- **Clarity**: Clear separation between functional and demonstration tests
- **Maintainability**: Easier to find and manage specific tests
- **CI/CD Friendly**: Possibility to run only functional tests by default
- **Documentation**: Each subpackage can have its own documentation
- **Selective Execution**: Use of build tags allows granular control

### Negative:

- **Reorganization**: Requires moving existing files
- **Build Tags**: Adds complexity to test execution
- **Structure**: Increases the depth of the directory structure

### Mitigated Risks:

- **Functionality Breakage**: Tests are moved, not removed
- **Confusion**: Clear documentation in each file explains the purpose
- **Execution**: Build tags allow selective execution without affecting functional tests

## Execution Commands

```bash
# Run only functional tests (default)
go test ./...

# Run demonstration tests
go test -tags demo ./testfailures/demo/...

# Run framework tests
go test -tags demo ./testfailures/framework/...

# Run all tests (including intentional failures)
go test -tags demo ./...
```

## Alternatives Considered

1. **Keep Current Structure**: Rejected for creating confusion about test purposes
2. **Use Suffixes**: Rejected for not solving the organization problem
3. **Single Directory**: Rejected for not allowing adequate categorization
4. **Build Tags Without Reorganization**: Rejected for not solving the clarity problem

## Implementation

The implementation was completed on 2024-12-19, including:

1. Creation of the `testfailures/` directory structure
2. Moving demonstration tests from `examples/` to `testfailures/demo/`
3. Moving framework tests from `prop/prop_test.go` to `testfailures/framework/`
4. Moving comparison tests from `quick/quick_test.go` to `testfailures/demo/`
5. Adding `//go:build demo` build tags to all moved tests
6. Updating documentation in each moved file

## References

- [ADR-001: Serial Shrinking Strategy](./adr-001-serial-shrinking.md)
- [ADR-002: Shrinking Strategies](./adr-002-shrinking-strategies.md)
- [ADR-003: Replay Command Line](./adr-003-replay-command-line.md)
- [ADR-004: Simple State Machine](./adr-004-simple-state-machine.md)