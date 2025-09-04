# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records (ADRs) for the RapidX project. ADRs document important architectural decisions, the context in which they were made, and their consequences.

## What are ADRs?

Architecture Decision Records are documents that capture important architectural decisions along with their context and consequences. They help teams understand why certain decisions were made and provide a historical record of the project's evolution.

## ADR Index

- [ADR-001: Serial Shrinking in Parallel Test Execution](adr-001-serial-shrinking.md)
  - Documents the decision to implement serial shrinking within parallel workers
  - Explains the trade-offs between parallel and serial shrinking approaches

- [ADR-002: Shrinking Strategy Selection (BFS vs DFS)](adr-002-shrinking-strategies.md)
  - Documents the decision to support both BFS and DFS shrinking strategies
  - Explains the characteristics and use cases for each strategy

- [ADR-003: Replay Functionality via Command Line](adr-003-replay-command-line.md)
  - Documents the decision to implement replay functionality for reproducing test failures
  - Explains the seed-based approach and command-line integration

- [ADR-004: Simple State Machine Testing](adr-004-simple-state-machine.md) **[PROPOSED]**
  - Proposes the implementation of state machine testing capabilities
  - Explains the API design and implementation approach for testing stateful systems

## ADR Template

When creating new ADRs, use the following template:

```markdown
# ADR-XXX: [Title]

## Status
[Proposed | Accepted | Rejected | Superseded]

## Context
[Describe the context and problem statement]

## Decision
[Describe the decision and its implementation]

## Consequences
### Positive
[Positive consequences]

### Negative
[Negative consequences]

### Neutral
[Neutral consequences]

## Alternatives Considered
[Describe alternatives that were considered and why they were rejected]

## Implementation Notes
[Any implementation-specific notes]

## References
[Links to relevant documentation, papers, or resources]

## Related ADRs
[Links to related ADRs]
```

## Guidelines

1. **One Decision Per ADR**: Each ADR should focus on a single architectural decision
2. **Clear Status**: Always include the current status of the decision
3. **Context First**: Provide sufficient context for understanding the decision
4. **Consequences**: Be honest about both positive and negative consequences
5. **Alternatives**: Document alternatives that were considered and why they were rejected
6. **References**: Include links to relevant documentation or resources

## Contributing

When making significant architectural decisions:

1. Create a new ADR using the template above
2. Use the next available ADR number
3. Update this README to include the new ADR
4. Submit the ADR for review before implementing the decision
5. Update the status once the decision is accepted and implemented