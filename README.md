# profile

Machine-bound encrypted vault and profile manager. Store and manage secrets organized by profiles, bound to the current machine.

## Installation

### Go Install

```bash
go install github.com/inovacc/profile@latest
```

### From Source

```bash
git clone https://github.com/inovacc/profile.git
cd profile
go build -o profile .
```

## Quick Start

```bash
# Initialize the vault
profile vault init

# Create a profile
profile vault profile create myapp

# Store a secret
profile vault set API_KEY sk-abc123 -p myapp

# Retrieve a secret
profile vault get API_KEY -p myapp

# List secrets
profile vault list -p myapp
```

## Environment Variable Release

Release secrets as environment variables with password-gated, time-limited access:

```bash
# Set a password
profile vault env password set

# Release secrets (default: 30 minutes)
profile vault env release --ttl 1h

# Export as shell variables
eval $(profile vault env export --format export)

# Revoke access
profile vault env revoke
```

## JSON Output

All commands support structured JSON output via the `--json` flag:

```bash
profile vault status --json
profile vault list --json
profile cmdtree --json
```

## CLI Tools

```bash
# Display command tree
profile cmdtree

# Generate AI-readable documentation
profile aicontext
profile aicontext --compact
profile aicontext --category vault
```

## License

See [LICENSE](LICENSE) for details.
