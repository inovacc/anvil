# Known Issues

## Open Issues

| # | Issue | Severity | Workaround |
|---|-------|----------|------------|
| 1 | LICENSE file is empty | Low | Populate with chosen license text |
| 2 | Build workflow targets `develop` branch which does not exist yet | Low | Create `develop` branch or change workflow to target `main` |
| 3 | `sboms` section in `.goreleaser.yaml` requires `syft` which may not be installed in CI runner | Low | SBOM generation silently skipped if `syft` not available |

## Resolved Issues

| Issue | Resolution | Date |
|-------|------------|------|
| goreleaser release failed due to missing `GITHUB_OWNER` env var | Removed explicit `release.github` config; goreleaser auto-detects from git remote | 2026-02-10 |
| goreleaser release header template used build-level variables (`.Os`, `.Arch`) | Simplified installation section to `go install` command | 2026-02-10 |
