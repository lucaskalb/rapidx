# RapidX - Property-Based Testing for Go

RapidX is a property-based testing library for Go that allows you to test properties of your code by generating random test cases and automatically shrinking counterexamples when failures are found.

## Packages

### prop - Property-Based Testing Framework

Package prop provides property-based testing functionality for Go. It allows you to test properties of your code by generating random test cases and automatically shrinking counterexamples when failures are found.

#### Functions

##### ForAll[T any](t *testing.T, cfg Config, g gen.Generator[T]) func(func(*testing.T, T))

ForAll creates a property-based test that generates test cases using the provided generator and runs them against the given test function. It returns a function that takes the test body as a parameter.

The test will generate cfg.Examples number of test cases, and if any fail, it will attempt to shrink the counterexample to find a minimal failing case.

Example usage:

```go
ForAll(t, prop.Default(), gen.Int())(func(t *testing.T, x int) {
    // Test property: x + 0 == x
    if x+0 != x {
        t.Errorf("addition identity failed for %d", x)
    }
})
```

#### Types

##### Config

Config holds the configuration for property-based testing.

```go
type Config struct {
    // Seed is the random seed used for test case generation.
    // If zero, a random seed will be generated based on the current time.
    Seed int64

    // Examples is the number of test cases to generate and run.
    Examples int

    // MaxShrink is the maximum number of shrinking steps to perform
    // when a counterexample is found.
    MaxShrink int

    // ShrinkStrat specifies the shrinking strategy to use.
    // Supported strategies: "bfs" (breadth-first), "dfs" (depth-first).
    ShrinkStrat string

    // StopOnFirstFailure determines whether to stop testing
    // after the first failing test case is found.
    StopOnFirstFailure bool

    // Parallelism specifies the number of parallel workers to use
    // for running test cases. Must be at least 1.
    Parallelism int
}
```

##### Default() Config

Default returns a Config with default values based on command-line flags. This is the recommended way to create a configuration for property-based testing.

---

### quick - Quick Testing Utilities

Package quick provides quick testing utilities for Go. It includes helper functions for common testing patterns, particularly for value comparison and assertion utilities.

#### Functions

##### Equal[T any](t *testing.T, got, want T)

Equal compares two values of the same type and fails the test if they are not equal. It uses go-cmp for deep comparison and provides detailed diff output when values differ. The function calls t.Helper() to mark itself as a test helper function.

Parameters:
- t: The testing.T instance for the current test
- got: The actual value obtained from the code under test
- want: The expected value

Example usage:

```go
quick.Equal(t, result, expected)
quick.Equal(t, []int{1, 2, 3}, []int{1, 2, 3})
quick.Equal(t, map[string]int{"a": 1}, map[string]int{"a": 1})
```

---

### gen - Generators for Property-Based Testing

Package gen provides generators for property-based testing in Go. It includes generators for various data types and utilities for creating custom generators with shrinking capabilities.

#### Constants

Common alphabet shortcuts (pure ASCII to avoid surprises):

```go
const (
    // AlphabetLower contains lowercase letters a-z.
    AlphabetLower = "abcdefghijklmnopqrstuvwxyz"

    // AlphabetUpper contains uppercase letters A-Z.
    AlphabetUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

    // AlphabetAlpha contains both lowercase and uppercase letters.
    AlphabetAlpha = AlphabetLower + AlphabetUpper

    // AlphabetDigits contains digits 0-9.
    AlphabetDigits = "0123456789"

    // AlphabetAlphaNum contains letters and digits.
    AlphabetAlphaNum = AlphabetAlpha + AlphabetDigits

    // AlphabetASCII contains all printable ASCII characters.
    AlphabetASCII = AlphabetAlphaNum + " !\"#$%&'()*+,-./:;<=>?@[\\]^_{|}~"
)
```

#### Functions

##### Basic Generators

- **Bool() Generator[bool]** - Generates boolean values uniformly. Shrink: prioritizes reducing to false (smaller counterexample by convention).

- **Int(size Size) Generator[int]** - Generates integers with automatic range based on Size. If sz.Max (or |sz.Min|) > 0: range := [-M, M], where M = max(|sz.Min|, |sz.Max|). Otherwise, uses default range [-100, 100].

- **IntRange(min, max int) Generator[int]** - Generates integers uniformly in the range [min, max] (inclusive). Ignores sz for the range (useful when you want explicit control).

- **Int64(size Size) Generator[int64]** - Generates 64-bit integers with automatic range based on Size. If no Size is provided, uses [-100, 100].

- **Int64Range(min, max int64) Generator[int64]** - Generates int64 uniformly in the range [min, max] (inclusive).

- **Uint(size Size) Generator[uint]** - Generates unsigned integers with automatic range based on Size. If no Size is provided, uses [0, 100].

- **UintRange(min, max uint) Generator[uint]** - Generates uint uniformly in the range [min, max].

- **Uint64(size Size) Generator[uint64]** - Generates unsigned 64-bit integers with automatic range based on Size. If nothing is provided, uses [0, 100].

- **Uint64Range(min, max uint64) Generator[uint64]** - Generates uint64 uniformly in the range [min, max] (inclusive).

##### Float Generators

- **Float32(size Size) Generator[float32]** - Generates float32 values with automatic range based on Size. Default: [-100, 100]. Does not include NaN/Inf.

- **Float32Range(min, max float32, includeNaN, includeInf bool) Generator[float32]** - Generates float32 in [min, max]; can optionally produce NaN/±Inf.

- **Float64(size Size) Generator[float64]** - Generates floats with automatic range based on Size. If no Size is provided, uses range [-100, 100]. Does not include NaN/Inf (focused on business numeric cases).

- **Float64Range(min, max float64, includeNaN, includeInf bool) Generator[float64]** - Generates floats uniformly in [min, max] (inclusive on finite bounds). Parameters includeNaN/includeInf allow injecting special cases.

##### String Generators

- **String(alphabet string, size Size) Generator[string]** - Generates strings using an alphabet (set of runes) and a Size. If size.Min/Max = 0, uses default: Min=0, Max=32. If alphabet is empty, uses AlphabetAlphaNum.

- **StringAlpha(size Size) Generator[string]** - Generates strings using only alphabetic characters.

- **StringAlphaNum(size Size) Generator[string]** - Generates strings using alphanumeric characters.

- **StringDigits(size Size) Generator[string]** - Generates strings using only digits.

- **StringASCII(size Size) Generator[string]** - Generates strings using all printable ASCII characters.

##### Collection Generators

- **ArrayOf[T any](elem Generator[T], n int) Generator[[]T]** - Generates a slice of **exact** length n, using the element generator. It is "array-like": great when you need to simulate [N]T. Shrink: cannot remove elements; only tries local shrink at each position, exploring multiple branches (BFS/DFS) and deduplicating candidates.

- **SliceOf[T any](elem Generator[T], size Size) Generator[[]T]** - Generates []T from an element generator. size.Min/Max control the length (default Min=0, Max=16). Shrink: (1) remove large blocks (half, quarter, ...) → remove indices, (2) remove isolated element (right→left), (3) try shrink on elements (propagating accept).

##### CPF Generators

- **CPF(masked bool) Generator[string]** - Generates valid CPF numbers; masked controls the format.

- **CPFAny() Generator[string]** - Generates CPF numbers with 50/50 chance of being masked or unmasked.

##### Combinator Functions

- **Const[T any](v T) Generator[T]** - Always returns the same value (without shrinking).

- **OneOf[T any](gs ...Generator[T]) Generator[T]** - Chooses uniformly from one of the generators.

- **Weighted[T any](weight func(T) float64, gs ...Generator[T]) Generator[T]** - Chooses a generator based on dynamic weights (by value). The strategy here captures which index was selected to be able to "shrink" reusing the shrinker of the chosen generator. Optionally, in shrinking it also tries to migrate to neighbors (other indices).

- **Map[A, B any](ga Generator[A], f func(A) B) Generator[B]** - Applies f: A -> B preserving shrinking (maps A's candidates).

- **Filter[T any](g Generator[T], pred func(T) bool, maxTries int) Generator[T]** - Keeps only values that satisfy pred. Implements "rebase" in shrink: when accepting, shrinks on top of the new minimum ensuring that the next candidates also satisfy the predicate.

- **Bind[A, B any](ga Generator[A], f func(A) Generator[B]) Generator[B]** - Bind (flatMap): the output generator depends on the value generated in A. Shrinking: first tries to shrink in B; when exhausted, shrinks in A and regenerates B.

- **From[T any](fn func(*rand.Rand, Size) (T, Shrinker[T])) Generator[T]** - Creates a Generator from a function that implements the Generator interface. This is a convenience function for creating custom generators.

##### Utility Functions

- **SetShrinkStrategy(s string)** - Sets the shrinking strategy for all generators. Valid strategies are "dfs" (depth-first search) and "bfs" (breadth-first search). Any other value defaults to "bfs".

- **ValidCPF(s string) bool** - Checks if a string is a valid CPF number.

- **MaskCPF(raw string) string** - Formats a raw CPF string with dots and dashes.

- **UnmaskCPF(s string) string** - Removes all non-digit characters from a CPF string.

#### Types

##### Generator[T any] interface

Generator is the public contract for all generators. It defines the interface that all generators must implement.

```go
type Generator[T any] interface {
    // Generate produces a value and a shrinker function for that value.
    // The random number generator and size constraints are provided as parameters.
    Generate(r *rand.Rand, sz Size) (value T, shrink Shrinker[T])
}
```

##### Shrinker[T any] func(accept bool) (next T, ok bool)

Shrinker proposes "smaller" candidates during the shrinking process. The accept parameter indicates whether the PREVIOUS candidate was accepted (i.e., it reproduced the failure). This allows the shrinker to "rebase" and generate new neighbors from the new minimum.

##### Size struct

Size controls the scale and limits of generators. It defines the minimum and maximum bounds for generated values.

```go
type Size struct {
    // Min is the minimum bound for generated values.
    Min int
    // Max is the maximum bound for generated values.
    Max int
}
```

##### GenFunc[T any] struct

GenFunc is a helper type for creating a Generator from a closure. It wraps a function that implements the Generator interface.

##### T[T any] = Generator[T]

T is an optional alias for Generator[T] for compatibility.

---

### examples - Example Usage

The examples package demonstrates how to use the rapidx property-based testing library. These examples show various testing patterns and how the shrinking mechanism helps find minimal counterexamples when properties fail.

#### Example Tests

##### CPF Examples

- **Test_CPF_AlwaysValid** - Demonstrates a property-based test for CPF generation. This test verifies that all generated CPF numbers are valid according to the CPF validation algorithm, and that the UnmaskCPF function is idempotent.

- **Test_CPF_MaskUnmaskRoundTrip** - Demonstrates testing the round-trip property of CPF masking and unmasking operations. This test verifies that unmasking a masked CPF and then masking it again produces the same result.

- **Test_CPF_Any_Valid** - Demonstrates testing CPFAny() generator which produces CPF numbers with random masking (50/50 chance of masked or unmasked). This test verifies that all generated CPF numbers are valid regardless of format.

- **Test_CPF_Invalid** - Demonstrates a property-based test that is designed to fail. This test expects all CPF numbers to start with '9', which is not true for valid CPF generation. This example shows how the shrinking mechanism will find a minimal counterexample when the property fails.

##### Integer Examples

- **Test_Slice_SomaNaoNegativa** - Demonstrates a property-based test with a custom generator that is designed to fail. This test verifies a false property: "the sum of a slice is always 0". The custom integer generator creates values in the range [-100, 100] with a simple shrinking strategy that approaches 0. This example shows how to create custom generators and how the shrinking mechanism will find a minimal counterexample when the property fails.

##### String Examples

- **Test_String_FalsaRegra** - Demonstrates a property-based test that is designed to fail. This test verifies a false property: "all generated strings are empty". This example shows how the shrinking mechanism will find a minimal counterexample when the property fails, helping developers understand why their assumptions are incorrect.

## Getting Started

1. Import the packages you need:
```go
import (
    "github.com/lucaskalb/rapidx/prop"
    "github.com/lucaskalb/rapidx/gen"
    "github.com/lucaskalb/rapidx/quick"
)
```

2. Write a property-based test:
```go
func TestAdditionIdentity(t *testing.T) {
    prop.ForAll(t, prop.Default(), gen.Int())(func(t *testing.T, x int) {
        if x+0 != x {
            t.Errorf("addition identity failed for %d", x)
        }
    })
}
```

3. Run your tests:
```bash
go test -v
```

## Command Line Flags

RapidX supports several command-line flags for configuring property-based tests:

- `-rapidx.seed` - Random seed for test case generation (default: 0, random seed based on current time)
- `-rapidx.examples` - Number of test cases to generate (default: 100)
- `-rapidx.maxshrink` - Maximum number of shrinking steps (default: 400)
- `-rapidx.shrink.strategy` - Shrinking strategy: "bfs" or "dfs" (default: "bfs")
- `-rapidx.shrink.subtests` - Use Go's subtest functionality (default: true)
- `-rapidx.shrink.parallel` - Number of parallel workers (default: 1)

Example usage:
```bash
go test -rapidx.examples=1000 -rapidx.maxshrink=500 -rapidx.shrink.strategy=dfs
```