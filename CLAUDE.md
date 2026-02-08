# profile

## Project Overview

Go CLI application built with Cobra.

- **Module**: `github.com/inovacc/profile`
- **Go version**: 1.25.5
- **Architecture**: Hexagonal/Clean (cmd/, internal/, pkg/)

## Build & Run

```bash
task build          # Build to dist/
task run            # Run the application
go run .            # Run directly
```

## Testing

```bash
task test           # Run all tests with coverage
task test:unit      # Unit tests only (skip integration)
task test:cover     # Show coverage percentage
task test:coverage  # Generate HTML coverage report
```

## Code Quality

```bash
task fmt            # Format code (go fmt + goimports)
task vet            # Static analysis
task lint           # golangci-lint
task lint:fix       # Lint with auto-fix
task check          # All quality checks (fmt, vet, lint, test)
```

## Dependencies

```bash
task deps           # Download, tidy, verify
task deps:upgrade   # Upgrade all to latest
```

## Release

```bash
task build:dev          # Snapshot build with goreleaser
task release            # Production release (requires git tag)
task release:snapshot   # Snapshot release
task release:check      # Validate goreleaser config
```

## Project Structure

```
profile/
├── cmd/            # CLI commands (Cobra)
│   ├── output.go   # JSON/text output helper (outputResult)
│   ├── cmdtree.go  # Command tree visualization
│   └── aicontext.go # AI context documentation generator
├── internal/       # Private application code
│   ├── crypto/     # AES-256-GCM encryption, HKDF key derivation, machine ID
│   └── sentinel/   # Time-limited release session management
├── pkg/vault/      # Public vault API (types, errors, operations)
├── docs/           # Documentation
├── Taskfile.yml    # Task runner configuration
├── .golangci.yml   # Linter configuration
├── .goreleaser.yaml # Release configuration
└── main.go         # Entry point
```

## Conventions

- Use `task` (Taskfile) for all automation
- Use `glix install` instead of `go install` for CLI tools; use `glix install .` to install from a local directory
- Table-driven tests, 80% coverage minimum
- Mute unused returns: `_, _ = fmt.Fprintln(w, output)`
- Use `log/slog` for structured logging
- All commands use `outputResult(cmd, jsonData, textFn)` for JSON/text dual output
- Global `--json` persistent flag on rootCmd inherited by all subcommands
- `visibleSubcommands()` in cmdtree.go filters hidden commands and "help"
