# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the anvil project.

## Index

| ADR                                            | Title                                       | Status   |
|------------------------------------------------|---------------------------------------------|----------|
| [0001](0001-machine-bound-encryption.md)       | Machine-Bound Encryption                    | Accepted |
| [0002](0002-password-gated-env-release.md)     | Password-Gated Environment Variable Release | Accepted |
| [0003](0003-tpm-hardware-backed-sealing.md)    | TPM 2.0 Hardware-Backed Master Key Sealing  | Accepted |
| [0004](0004-executable-based-plugin-system.md) | Executable-Based Plugin System              | Accepted |

## Creating a New ADR

Use the template below:

```markdown
# ADR-NNNN: Title

## Status
Proposed | Accepted | Deprecated | Superseded by [ADR-XXXX](link)

## Context
What is the issue motivating this decision?

## Decision
What is the change we're proposing/doing?

## Consequences
### Positive
- Benefit 1

### Negative
- Drawback 1

## Alternatives Considered
Description and why not chosen.
```
