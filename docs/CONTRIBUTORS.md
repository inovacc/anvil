# Contributors

## Maintainers

| Name        | Role           | GitHub                                         |
|-------------|----------------|------------------------------------------------|
| dyammarcano | Creator & Lead | [@dyammarcano](https://github.com/dyammarcano) |

## How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Write tests for new functionality
4. Ensure `task check` passes (fmt, vet, lint, test)
5. Submit a pull request

## Contribution Guidelines

- Follow existing code patterns and conventions
- Table-driven tests, 80% coverage minimum
- Use `log/slog` for structured logging
- Use `task` (Taskfile) for all automation
- Run `gitleaks detect --source .` before committing
