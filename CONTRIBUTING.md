# Contributing Guidelines

This document outlines development procedures, code standards, and verification steps for contributing to `svault`.

## Code of Conduct

Contributors must adhere to the standards described in [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## Development Setup

Requirements:
- Go version 1.25 or higher
- Standard Make utility
- Git

Clone the repository and verify the test suite:
```bash
git clone https://github.com/dafagareth/svault.git
cd svault
go test ./...
```

## Development Workflow

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feat/feature-name
   ```
2. Implement code modifications along with corresponding unit tests.
3. Run all verification checks using the development script:
   ```bash
   ./dev.sh check
   ```
   This runs `gofmt`, `go vet`, and `go test -race ./...` in sequence. All three must pass before opening a Pull Request.
4. Optionally, synchronize Go module dependencies:
   ```bash
   ./dev.sh tidy
   ```
5. Commit changes adhering to the [Conventional Commits](https://www.conventionalcommits.org/) specification.
6. Push the feature branch and open a Pull Request against `main`.

## Commit Specification

Commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/) format: `<type>(<scope>): <description>`.

Examples:
- `feat(secrets): add batch secret retrieval support to get command`
- `fix(storage): resolve concurrent file lock race condition`
- `docs(readme): update namespace resolution hierarchy section`
- `test(crypto): add edge case validation for empty nonces`
- `chore(deps): update golang.org/x/crypto module dependency`

## Code and Architectural Standards

Write code, comments, and documentation strictly in clear English.

Propagate errors explicitly up the call stack. Do not invoke `panic` in package libraries or CLI commands.

Format error strings in lowercase without trailing punctuation, for example: `fmt.Errorf("failed to load session: %w", err)`.

Use `crypto/rand` exclusively for security and cryptographic primitives. Do not use `math/rand`.

All file writes to vault, session, and config must use atomic replacement (`os.CreateTemp`, `Sync`, `Close`, `os.Rename`). Do not use `os.WriteFile` directly for files that must remain consistent on crash.

Avoid redundant comments that restate function signatures or variable names.

## Testing Standards

Packages and CLI subcommands must maintain associated unit tests in `_test.go` files.

Use `t.TempDir()` and `t.Setenv()` to isolate filesystem interactions and environment variables during testing.

Tests must not mutate host configuration directories (`~/.svault`) or the system clipboard.

Use the `SetupVault` helper in `cmd/common/test_helpers.go` to construct isolated test vault fixtures.

## Reporting Issues and Vulnerabilities

Submit functional bugs and feature requests via the GitHub Issue Tracker. Include system architecture, Go version, exact command arguments, and reproduction steps.

Security vulnerabilities must not be reported via public GitHub issues. Follow the disclosure protocol defined in [SECURITY.md](SECURITY.md).
