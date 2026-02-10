# Known Bugs

## Open Bugs

_No known bugs at this time._

## Resolved Bugs

| Bug | Severity | Resolution | Date |
|-----|----------|------------|------|
| goreleaser `GITHUB_OWNER` env var not set in reusable workflow | High | Replaced with `GITHUB_REPOSITORY_OWNER` (auto-set by GitHub Actions), then removed explicit `release.github` config entirely to let goreleaser auto-detect from git remote | 2026-02-10 |
| goreleaser release header used `.Os`/`.Arch` (build-level vars unavailable in release templates) | Medium | Simplified header to use `go install` command instead of platform-specific download URLs | 2026-02-10 |

## Reporting Bugs

To report a bug, include:
1. Steps to reproduce
2. Expected behavior
3. Actual behavior
4. OS and Go version
5. Relevant error output
