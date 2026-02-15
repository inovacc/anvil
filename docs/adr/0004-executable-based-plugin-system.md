# ADR-0004: Executable-Based Plugin System

## Status
Accepted

## Context
Anvil needs extensibility for secret lifecycle hooks (e.g., notifying external systems on secret changes) and external secret providers (e.g., fetching secrets from AWS Secrets Manager or HashiCorp Vault). Go's standard `plugin` package only works on Linux and macOS — it does not support Windows. Since anvil is a cross-platform tool (Windows, Linux, macOS), the plugin mechanism must work everywhere.

## Decision
Use an executable-based plugin model where hooks and providers are external programs that communicate with anvil via JSON over stdin/stdout.

### Hook Protocol
Anvil invokes the hook executable with a JSON payload on stdin:
```json
{"secret_key": "API_KEY", "profile_name": "prod", "db_path": "/path/to/vault.db"}
```

The hook writes a JSON response to stdout:
```json
{"allow": true}
```

For pre-event hooks (`pre-set`, `pre-get`, `pre-delete`), returning `{"allow": false, "message": "reason"}` blocks the operation. Post-event hook errors are logged but never block.

### Provider Protocol
Provider executables receive a JSON request on stdin and return the secret value on stdout, enabling integration with external secret stores.

### Configuration
Plugin configuration is stored in `plugins.json` alongside the vault database, separate from the SQLite schema. This avoids schema migrations for plugin changes.

```json
{
  "hooks": [{"event": "post-set", "command": "/usr/local/bin/notify", "args": ["--channel", "ops"]}],
  "providers": [{"name": "aws", "command": "/usr/local/bin/aws-provider", "prefix": "aws/"}]
}
```

### Integration
The `PluginManager` is loaded during `Vault.Open()` and hooks are automatically invoked in `Set()`, `Get()`, and `Delete()` operations.

## Consequences

### Positive
- Works on all platforms (Windows, Linux, macOS) — any executable works
- Language-agnostic: plugins can be written in any language (Go, Python, Bash, etc.)
- Process isolation: plugin crashes don't crash anvil
- Simple protocol: JSON over stdin/stdout is universally supported
- No schema migration needed: config is a separate JSON file

### Negative
- Slower than in-process plugins (~50ms per hook invocation due to process spawn)
- No shared memory or direct function calls
- Plugin errors are harder to debug (separate process, no stack traces)
- No plugin discovery or registry — manual configuration required

## Alternatives Considered
- **Go `plugin` package:** Not supported on Windows; requires plugins compiled with the same Go version and build flags
- **gRPC plugin model (like HashiCorp go-plugin):** More complex, requires gRPC dependency, overkill for simple hooks
- **Shared library (`.so`/`.dll`):** Platform-specific, complex build requirements, ABI compatibility issues
- **Lua/JavaScript embedding:** Adds runtime dependency, security concerns with embedded scripting
