# RapidX

RapidX is a property-based testing library for Go that allows you to test properties of your code by generating random test cases and automatically shrinking counterexamples when failures are found.

## Features

- **Property-based testing** with automatic test case generation
- **Intelligent shrinking** to find minimal counterexamples
- **Rich generator library** for common data types
- **Custom generators** with user-defined shrinking logic
- **Parallel execution** for faster test runs
- **Command-line configuration** via flags
- **Domain-specific generators** (e.g., CPF validation)

## Quick Start

```go
package main

import (
    "testing"
    "github.com/lucaskalb/rapidx/prop"
    "github.com/lucaskalb/rapidx/gen"
)

func TestAdditionIdentity(t *testing.T) {
    prop.ForAll(t, prop.Default(), gen.Int())(func(t *testing.T, x int) {
        if x+0 != x {
            t.Errorf("addition identity failed for %d", x)
        }
    })
}
```

## Installation

```bash
go get github.com/lucaskalb/rapidx
```

## Documentation

- [Complete Documentation](DOCUMENTATION.md)
- [Package Documentation](prop_docs.txt) - Property-based testing framework
- [Generator Documentation](gen_docs.txt) - Data generators
- [Quick Utilities Documentation](quick_docs.txt) - Testing utilities

## Examples

See the `examples/` directory for comprehensive usage examples including:
- Basic property testing
- Custom generators
- CPF validation testing
- String and integer property tests

## License

This project is licensed under the MIT License.