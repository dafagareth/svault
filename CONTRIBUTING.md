# Contributing to svault

Thank you for your interest in contributing to `svault`! We welcome contributions, bug reports, feature requests, and documentation improvements.

---

## 📜 Code of Conduct

All contributors and maintainers are expected to adhere to our [Code of Conduct](CODE_OF_CONDUCT.md). Please report unacceptable behavior in accordance with the guidelines described therein.

---

## 🛠️ Development Setup

### Prerequisites
- **Go**: Version 1.25 or higher
- **Make**: Standard build execution tool
- **Git**

### Getting Started
1. **Fork and Clone the Repository:**
   ```bash
   git clone https://github.com/YOUR-USERNAME/svault.git
   cd svault
   ```

2. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

3. **Verify Development Setup:**
   ```bash
   go test ./...
   ```

---

## 🔄 Development Workflow

1. Create a logical feature branch from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. Implement your changes alongside corresponding unit tests.
3. Run local code verification checks:
   ```bash
   make test       # Run complete unit test suite
   make vet        # Run static analysis check
   ```
4. Verify race detector compliance:
   ```bash
   go test -race ./...
   ```
5. Commit your changes adhering to conventional commit specifications.
6. Push your feature branch to your fork and open a Pull Request against `main`.

---

## 📝 Commit Conventions

We enforce the [Conventional Commits specification](https://www.conventionalcommits.org/). Commit messages must follow the `<type>(<scope>): <description>` format.

**Examples:**
- `feat(cmd): add rotate command to re-encrypt vault payload`
- `fix(storage): resolve race condition on concurrent session writes`
- `docs(readme): update namespace resolution order documentation`
- `test(crypto): add edge case validation for empty nonces`

---

## 📐 Code & Style Guidelines

- **Language**: Write all code, comments, documentation, and user-facing messages strictly in clear, concise English.
- **Error Handling**: Pass errors explicitly up the execution stack. Avoid using `panic` in core package libraries or CLI commands.
- **Error Formatting**: Error strings must be formatted in lowercase without trailing punctuation (e.g., `fmt.Errorf("failed to load session: %w", err)`).
- **Cryptographic Security**: Use `crypto/rand` exclusively for security and cryptographic primitives. Never use `math/rand`.
- **Code Cleanliness**: Omit redundant comments that simply restate function signatures or variable identifiers.

---

## 🧪 Testing Standards

- **Test Coverage**: Every package and CLI subcommand under `cmd/` must maintain accompanying `_test.go` test suites.
- **Environment Isolation**: Always utilize `t.TempDir()` and `t.Setenv()` to isolate filesystem targets and environment variables during test execution.
- **Environment Safety**: Unit tests **must never** modify the real host system environment (`~/.svault`) or system clipboards.
- **Testing Helpers**: Utilize the `setupVault` helper in [`cmd/helpers_test.go`](cmd/helpers_test.go) to construct isolated test vault fixtures.

---

## 🐛 Bug Reports & Feature Requests

Submit bug reports and feature requests via the [GitHub Issue Tracker](https://github.com/dafagareth/svault/issues).

When reporting a bug, please include:
- Operating System and System Architecture
- Go version (`go version`)
- Exact command invocation and flags executed
- Complete error log output and steps to reproduce

> [!IMPORTANT]
> **Security Vulnerabilities:** Do **NOT** open public GitHub issues for security vulnerabilities. Please follow the disclosure protocol outlined in [SECURITY.md](SECURITY.md).
