# ADR-001: Serial Shrinking in Parallel Test Execution

## Status
Accepted

## Context

RapidX implements property-based testing with automatic test case generation and intelligent shrinking of counterexamples. The library supports parallel execution of test cases to improve performance, but faces a design decision regarding how to handle the shrinking process when tests run in parallel.

### Key Requirements
- Generate multiple test cases in parallel for better performance
- Automatically shrink counterexamples when tests fail
- Maintain deterministic and reproducible results
- Keep the implementation simple and maintainable

### Technical Constraints
- Shrinking algorithms typically maintain internal state (queues, seen sets, etc.)
- Shrinking steps often depend on the result of previous steps
- Go's testing framework (`testing.T`) has specific threading requirements
- Random number generation must be thread-safe

## Decision

**We will implement serial shrinking within each parallel worker, while maintaining parallel generation and initial test execution.**

### Implementation Details

1. **Parallel Generation**: Multiple goroutines generate and execute initial test cases concurrently
2. **Serial Shrinking**: When a test fails, the worker that found the failure performs shrinking sequentially
3. **Independent Workers**: Each worker maintains its own shrinking state and process
4. **Thread Safety**: Shared resources (random number generator) are protected by mutexes

### Code Structure

```go
// Parallel execution with serial shrinking per worker
func runParallel[T any](t *testing.T, cfg Config, g gen.Generator[T], body func(*testing.T, T), seed int64, r *rand.Rand) {
    // ... setup parallel workers ...
    
    for i := 0; i < cfg.Parallelism; i++ {
        go func(workerID int) {
            for testIndex := range testChan {
                // 1. Generate test case (parallel)
                val, shrink := g.Generate(r, gen.Size{})
                
                // 2. Execute test (parallel)
                passed := t.Run(name, func(st *testing.T) { body(st, val) })
                
                if passed {
                    continue
                }
                
                // 3. Shrink counterexample (serial within worker)
                min := val
                for steps < cfg.MaxShrink {
                    next, ok := shrink(acceptedPrev)
                    // ... sequential shrinking logic ...
                }
            }
        }(i)
    }
}
```

## Consequences

### Positive
- **Simplicity**: Avoids complex synchronization between shrinking steps
- **Determinism**: Each worker's shrinking process is predictable and reproducible
- **Performance**: Still benefits from parallel test generation and execution
- **Isolation**: Workers don't interfere with each other's shrinking process
- **Maintainability**: Easier to debug and reason about shrinking behavior

### Negative
- **Potential Bottleneck**: Shrinking can become a bottleneck if `MaxShrink` is very high
- **Resource Utilization**: Some workers may finish early while others are still shrinking
- **Limited Parallelism**: No parallelization of the shrinking process itself

### Neutral
- **Memory Usage**: Each worker maintains its own shrinking state
- **Complexity**: Moderate complexity increase compared to fully serial execution

## Alternatives Considered

### 1. Fully Parallel Shrinking
**Approach**: Parallelize the shrinking process itself across multiple workers.

**Rejected because**:
- Shrinking algorithms have sequential dependencies
- Complex synchronization requirements
- Risk of race conditions and non-deterministic behavior
- Significant implementation complexity

### 2. Centralized Shrinking
**Approach**: Collect all failures and perform shrinking in a single thread.

**Rejected because**:
- Loses parallelism benefits for the most time-consuming part
- Requires collecting and queuing all failures
- Memory overhead for storing intermediate results

### 3. Hybrid Approach
**Approach**: Parallel shrinking with shared state management.

**Rejected because**:
- Extremely complex synchronization requirements
- High risk of deadlocks and race conditions
- Difficult to maintain and debug
- Minimal performance benefit for the added complexity

## Implementation Notes

### Thread Safety Considerations
- Random number generator is protected by `sync.Mutex`
- Each worker operates on independent data structures
- No shared state between shrinking processes

### Performance Characteristics
- Parallelism is most effective when test generation is the bottleneck
- Shrinking typically represents a small portion of total execution time
- The approach scales well with the number of available CPU cores

### Future Considerations
- Could be extended to support parallel shrinking for specific generator types
- May benefit from work-stealing algorithms for better load balancing
- Consider adaptive shrinking strategies based on failure patterns

## References

- [Property-Based Testing with QuickCheck](https://www.cse.chalmers.se/~rjmh/QuickCheck/manual.html)
- [Go Testing Package Documentation](https://pkg.go.dev/testing)
- [Concurrent Programming in Go](https://golang.org/doc/effective_go.html#concurrency)

## Related ADRs

- ADR-002: Shrinking Strategy Selection (BFS vs DFS)
- ADR-003: Random Number Generator Thread Safety