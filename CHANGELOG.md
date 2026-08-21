# Changelog

All notable changes to this project are documented in this file.
The format follows [Keep a Changelog](https://keepachangelog.com/) and [Semantic Versioning](https://semver.org/).

## [1.0.0] - 2026-08-21

### Fixed
- Fixed `svault backup` writing the backup payload non-atomically, preventing corrupted backups on process crash.
- Fixed `svault export --output` writing the `.env` output file non-atomically.
- Fixed `svault ns use` silently accepting non-existent namespace names; it now correctly validates against the vault.
- Fixed `svault doctor` duplicating clipboard detection logic and failing to respect `common` configurations.
- Refactored `svault` base command (`rootCmd`) to return an explicit error instead of hard-exiting with `os.Exit(1)` when the vault is missing, improving testability.

### Added

- Initial release of `svault`, a zero-dependency CLI application for managing encrypted local environment secrets.
- Authenticated symmetric encryption using AES-256-GCM with key derivation via Argon2id and a 16-byte random salt per vault initialization.
- Isolated namespaces for project environment management with automatic Git repository root directory detection.
- Session token caching in user-private directory (`~/.svault/.session`, mode 0600) with configurable TTL (default 30 minutes via `SVAULT_SESSION_TTL`).
- Subprocess environment injection (`svault exec`) to run child processes without writing plaintext secrets to disk.
- Export and import support for `.env`, JSON, and YAML formats.
- Atomic file writing (`CreateTemp`, `Sync`, `Rename`) across vault payload, session, and configuration stores for crash consistency.
- Cross-process file locking (`syscall.Flock` on Unix, `LockFileEx` on Windows) to prevent concurrent write race conditions.
- Automatic rolling backup of vault payload prior to each write operation, retained at up to five timestamped copies.
- Append-only operational audit log (`vault.log`) with automated key name privacy masking via SHA-256 hash prefix (`id:xxxxxxxx`).
- Password generator using `crypto/rand` with configurable length and optional symbol exclusion.
- System clipboard integration with auto-clear timeout across all platforms (Wayland, X11, macOS, Windows).
- Cross-platform browser launcher in `svault open` (`xdg-open` on Linux, `open` on macOS, `start` on Windows).
- Dynamic shell tab-completion for active secret keys and namespace names across `get`, `delete`, `edit`, `rename`, `copy`, `open`, and `ns use` subcommands.
- CLI subcommands grouped in `--help` output into six functional domains: Session and Authentication, Secret Operations, Environment and Integration, Namespaces, Utilities and Password Generation, and Maintenance and Diagnostics.
- Distribution packaging manifests for Arch Linux (AUR source and binary), Homebrew, Scoop, Debian (.deb), and Fedora (.rpm).
- Installation scripts for Linux and macOS (`install.sh`) and Windows (`install.ps1`).
- Developer tooling via `dev.sh` with `check` target (gofmt, go vet, go test -race) and `tidy` target.
