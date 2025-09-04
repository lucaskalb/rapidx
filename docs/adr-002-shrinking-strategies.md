# ADR-002: Shrinking Strategy Selection (BFS vs DFS)

## Status
Accepted

## Context

RapidX implements intelligent shrinking to find minimal counterexamples when property-based tests fail. The shrinking process needs to explore a tree of possible smaller candidates, and we need to decide how to traverse this tree to find the best counterexample.

### Key Requirements
- Find minimal counterexamples efficiently
- Provide predictable and understandable shrinking behavior
- Allow users to choose between different shrinking strategies
- Maintain good performance for various use cases

### Shrinking Process
When a test fails, the shrinking process:
1. Generates a tree of "smaller" candidates from the failing input
2. Tests each candidate to see if it still fails
3. Continues with successful candidates until no smaller failing case is found

## Decision

**We will implement two shrinking strategies: Breadth-First Search (BFS) as default, and Depth-First Search (DFS) as an alternative, with user-configurable selection.**

### Strategy Details

#### BFS (Breadth-First Search) - Default
- **Behavior**: Explores all candidates at the current "level" before moving deeper
- **Advantages**: 
  - Finds counterexamples closer to the original failing input
  - More predictable shrinking path
  - Better for understanding why a property fails
- **Use cases**: Debugging, understanding failure causes, general-purpose testing

#### DFS (Depth-First Search) - Alternative
- **Behavior**: Explores one shrinking path as deeply as possible before trying alternatives
- **Advantages**:
  - Often finds smaller counterexamples faster
  - More aggressive shrinking
  - Better for finding the absolute minimum failing case
- **Use cases**: Finding minimal examples, performance-critical shrinking

### Implementation

```go
// Global shrinking strategy
var shrinkStrategy = "bfs"

func SetShrinkStrategy(s string) {
    if s == "dfs" {
        shrinkStrategy = "dfs"
    } else {
        shrinkStrategy = "bfs"  // Default to BFS
    }
}

// Usage in shrinking logic
pop := func() (T, bool) {
    if len(queue) == 0 { return zero, false }
    if shrinkStrategy == "dfs" {
        // LIFO - Depth-First
        v := queue[len(queue)-1]
        queue = queue[:len(queue)-1]
        return v, true
    }
    // FIFO - Breadth-First (default)
    v := queue[0]
    queue = queue[1:]
    return v, true
}
```

## Consequences

### Positive
- **Flexibility**: Users can choose the strategy that best fits their needs
- **Predictability**: BFS provides more predictable and understandable results
- **Performance**: DFS can find smaller examples faster in many cases
- **Backward Compatibility**: BFS as default maintains expected behavior
- **Simplicity**: Easy to understand and implement

### Negative
- **Configuration Complexity**: Users need to understand the trade-offs
- **Implementation Overhead**: Need to maintain two different traversal strategies
- **Documentation Burden**: Must explain when to use each strategy

### Neutral
- **Memory Usage**: Both strategies use similar memory patterns
- **Code Complexity**: Moderate increase in complexity

## Alternatives Considered

### 1. BFS Only
**Approach**: Implement only breadth-first search.

**Rejected because**:
- DFS often finds smaller counterexamples faster
- Some users prefer more aggressive shrinking
- Limited flexibility for different use cases

### 2. DFS Only
**Approach**: Implement only depth-first search.

**Rejected because**:
- BFS provides more predictable and understandable results
- Better for debugging and understanding failure causes
- More familiar to users coming from other property-based testing tools

### 3. Adaptive Strategy
**Approach**: Automatically choose strategy based on input characteristics.

**Rejected because**:
- Adds significant complexity
- Difficult to predict which strategy will be better
- May lead to inconsistent behavior

### 4. Multiple Strategies
**Approach**: Implement more than two strategies (e.g., random, best-first).

**Rejected because**:
- Diminishing returns on additional strategies
- Increased complexity and maintenance burden
- BFS and DFS cover the main use cases effectively

## Implementation Notes

### Configuration
- Strategy is set globally via `SetShrinkStrategy()`
- Can be configured via command-line flag: `-rapidx.shrink.strategy`
- Default is BFS for predictable behavior

### Performance Characteristics
- **BFS**: More predictable time complexity, better for understanding
- **DFS**: Often faster to find minimal examples, but less predictable
- Both strategies have similar memory usage patterns

### User Guidance
- Recommend BFS for most use cases
- Suggest DFS when finding the smallest possible counterexample is important
- Provide clear documentation about the trade-offs

## Examples

### BFS Example
```
Original: [1, 2, 3, 4, 5]
Level 1:  [1, 2, 3, 4], [1, 2, 3, 5], [2, 3, 4, 5], [1, 3, 4, 5], [1, 2, 4, 5]
Level 2:  [1, 2, 3], [1, 2, 4], [1, 3, 4], [2, 3, 4], ...
```

### DFS Example
```
Original: [1, 2, 3, 4, 5]
Path 1:   [1, 2, 3, 4, 5] → [1, 2, 3, 4] → [1, 2, 3] → [1, 2] → [1]
Path 2:   [1, 2, 3, 4, 5] → [1, 2, 3, 5] → [1, 2, 5] → [1, 5] → [1]
```

## References

- [QuickCheck: Shrinking](https://www.cse.chalmers.se/~rjmh/QuickCheck/manual.html#shrinking)
- [Hypothesis: Shrinking](https://hypothesis.readthedocs.io/en/latest/shrinking.html)
- [Tree Traversal Algorithms](https://en.wikipedia.org/wiki/Tree_traversal)

## Related ADRs

- ADR-001: Serial Shrinking in Parallel Test Execution
- ADR-003: Random Number Generator Thread Safety