# Contributing to aipkg

Thanks for your interest in contributing to aipkg! This document covers everything you need to get started.

## Development setup

### Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [Task](https://taskfile.dev) (task runner)
- [golangci-lint](https://golangci-lint.run/docs/install/) v2
- [pre-commit](https://pre-commit.com/#install)

### Getting started

```sh
git clone https://github.com/tjespers/dofs.git
cd dofs
pre-commit install --hook-type pre-commit --hook-type commit-msg
task build
```

## Development workflow

```sh
task build          # build binary to dist/
task test           # run tests
task lint           # run golangci-lint
task lint:fix       # auto-fix lint issues
task fmt            # format code
task check          # lint + vet + test (full check)
task tidy           # go mod tidy
```

Run `task check` before pushing to catch issues early.

## Project structure

```
cmd/dofs/       # main entry point
internal/       # private packages (all library code goes here)
```

No `/pkg` directory. This is a CLI, not a library.

## Commit conventions

We use [Conventional Commits](https://www.conventionalcommits.org/). The pre-commit hook enforces this automatically.

Common prefixes: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`, `ci:`

### DCO sign-off

All commits must include a DCO sign-off line. Use the `-s` flag:

```sh
git commit -s -m "feat: add manifest validation"
```

This adds a `Signed-off-by` trailer to your commit, certifying you have the right to submit the contribution under the project's license.

### Linking to issues

If your commit resolves a GitHub issue, reference it in the commit body:

```
feat: add status subcommand

Implement a status subcommand that shows DOFS health status.

Closes: #3
```

## Pull requests

- Keep PRs focused on a single change
- Include tests for new functionality
- Make sure `task check` passes
- Write a clear description of what changed and why

## Code style

Go formatting is enforced by `gofmt` and `goimports` via golangci-lint. No manual style decisions needed beyond what the linters enforce.

## License

By contributing, you agree that your contributions will be licensed under the [Apache-2.0 License](LICENSE).
