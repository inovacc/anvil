# Known Issues

## Open Issues

_No known issues at this time._

## Resolved Issues

| Issue | Resolution | Date |
|-------|------------|------|
| `sboms` section in `.goreleaser.yaml` requires `syft` which may not be installed in CI runner | Removed `sboms` section entirely; SBOM generation not needed | 2026-02-15 |
| LICENSE file is empty | Populated with BSD 3-Clause license | 2026-02-14 |
| Build workflow targets `develop` branch which does not exist yet | Changed workflow to target `main` branch, removed unnecessary GUI deps | 2026-02-14 |
| CLI dumps full usage text on every error, leaks internal error chains, `--json` flag doesn't affect errors | Added `UserError` type with hints, `handleError` for text/JSON formatting, `SilenceErrors`/`SilenceUsage` on rootCmd | 2026-02-10 |
| All vault commands failed with `unable to open database file: out of memory (14)` — `GetApplicationDirectory()` returned a directory path, not a file path | Appended `vault.db` filename via `filepath.Join(dbPath, dbFileName)` in `Init()`, `Open()`, `GetStatus()`, and `DefaultDBPath()` | 2026-02-10 |
| goreleaser release failed due to missing `GITHUB_OWNER` env var | Removed explicit `release.github` config; goreleaser auto-detects from git remote | 2026-02-10 |
| goreleaser release header template used build-level variables (`.Os`, `.Arch`) | Simplified installation section to `go install` command | 2026-02-10 |
