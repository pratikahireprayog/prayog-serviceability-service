---
description: 
globs: 
alwaysApply: true
---
# Go Project Structure Rule (Go 1.24.2)

This project follows the [Standard Go Project Layout](mdc:https:/github.com/golang-standards/project-layout) and is intended for Go version **1.24.2**.

## Directory Structure

- `/cmd` — Main applications for this project. Each application has its own subdirectory.
- `/internal` — Private application and library code. This code is not importable by other projects.
- `/pkg` — Public libraries intended to be used by external applications.
- `/api` — API definitions (OpenAPI/Swagger specs, JSON schema, protocol definitions).
- `/configs` — Configuration file templates or default configs.
- `/scripts` — Build, install, analysis, and other scripts.
- `/build` — Packaging and CI/CD configurations and scripts.
- `/deployments` — Deployment configurations and templates (e.g., Docker, Kubernetes).
- `/test` — Additional external test apps and test data.
- `/docs` — Design and user documentation.
- `/tools` — Supporting tools for this project.
- `/examples` — Example applications and/or public libraries.
- `/third_party` — External helper tools, forked code, and other 3rd party utilities.
- `/githooks` — Git hooks.
- `/assets` — Other assets (images, logos, etc).
- `/website` — Project website data (if not using GitHub Pages).

## Coding Guidelines

- **Go Version:** All code must be compatible with Go 1.24.2.
- **APIs:** Use only the Go standard library for building APIs (e.g., `net/http`).
- **SOLID Principles:** Code must adhere to SOLID principles:
  - Single Responsibility Principle
  - Open/Closed Principle
  - Liskov Substitution Principle
  - Interface Segregation Principle
  - Dependency Inversion Principle
- **Clean Code:**
  - Write readable, maintainable, and well-documented code.
  - Use meaningful names, small functions, and clear logic.
  - Avoid code duplication and large functions/files.
- **Design Patterns:**
  - Apply Go-idiomatic design patterns where appropriate (e.g., Factory, Strategy, Adapter, Decorator, etc.).
  - Use interfaces to abstract dependencies and enable testing.
- **Testing:**
  - Place unit and integration tests in the `/test` directory or alongside the code with `_test.go` suffix.
  - Use the standard `testing` package.
- **Documentation:**
  - Document all public types, functions, and packages using GoDoc conventions.
  - Maintain additional documentation in `/docs`.

## References
- [Standard Go Project Layout](mdc:https:/github.com/golang-standards/project-layout)
- [GoDoc](mdc:https:/pkg.go.dev/golang.org/x/tools/cmd/godoc)
- [SOLID Principles](mdc:https:/en.wikipedia.org/wiki/SOLID)
- [Go Design Patterns](mdc:https:/refactoring.guru/design-patterns/go)

## Example
```text
project-root/
  cmd/
  internal/
  pkg/
  api/
  configs/
  scripts/
  build/
  deployments/
  test/
  docs/
  tools/
  examples/
  third_party/
  githooks/
  assets/
  website/
  README.md
  go.mod
  go.sum
```

This rule should be referenced for all new Go code and when reviewing project structure or code quality.
