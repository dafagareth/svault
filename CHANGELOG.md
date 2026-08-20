# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/), and the project uses
[semantic versioning](https://semver.org/).

## [2.2.1] - 2026-06-08

### Fixed

- `svault` with no arguments shows help again instead of listing keys; only
  flag-leading invocations (for example `svault -l`) route to `list`.
- `list -l` (long format) no longer reveals values. It shows the key and value
  length only; add `-v`/`--values` to reveal values, or `-m` to mask them.

## [2.2.0] - 2026-06-08

### Added

- `list` is now the default command. Running the bare binary (`svault`) lists
  keys like `ls`, and `svault -l`, `svault -m`, `svault --json`, etc. are passed
  straight to `list`. `svault -v`/`--version` and `svault -h`/`--help` are
  unchanged.

## [2.1.0] - 2026-06-08

### Added

- `list` gained flags: `--values/-v` (show values), `--long/-l` (key, value
  length, value), `--mask/-m` (masked values for screen sharing),
  `--filter/-F` (filter keys by substring), `--count/-c` (count only),
  `--json` (machine-readable output), and `--all/-a` shorthand. Boolean flags
  combine like `ls`, for example `svault list -lm` or `svault list -ac`.
- `get` now accepts multiple keys (`svault get DB_HOST DB_PORT`) and a
  `--clip/-c` flag that copies a value to the clipboard, clearing it after
  `--timeout` seconds (default 20).
- `set --clip` reads the value from the clipboard, keeping it out of shell
  history (like `--stdin`).
- `env --format/-f` outputs `shell` (default), `dotenv`, `json`, or `yaml`.
- `completion` command generates scripts for bash, zsh, fish, and powershell.
- `status` now reports namespace count, total keys, the active namespace, and
  the vault file size when unlocked.
- `check` gained `--extra` (list vault keys not present in the file) and
  `--strict` (exit non-zero when any key is missing, useful in CI).

### Changed

- `list`, `get`, `set`, `check`, and `status` now show a longer description and
  usage examples in `--help`, matching the existing `env` help.

## [2.0.1] - 2026-06-07

### Fixed

- Error messages were printed twice (for example on an unknown command). Cobra
  already prints the error itself, and `Execute` printed it again. The duplicate
  print is removed, and usage output is silenced so a plain runtime error no
  longer dumps the full help text.

## [2.0.0] - 2026-06-07

### Changed (breaking)

- Renamed the tool from `stash` to `svault` to avoid clashing with other
  projects of the same name (notably on the AUR and with `git stash`).
  - The command is now `svault` instead of `stash`.
  - The vault directory moved from `~/.stash` to `~/.svault`.
  - The session file moved from `/tmp/.stash_session` to `/tmp/.svault_session`.
  - Environment variables are now `SVAULT_SESSION_TTL` and `SVAULT_NS`.

### Migration

If you used a previous version, move your existing vault into the new location:

```bash
mv ~/.stash ~/.svault
```

Then unlock again with `svault unlock`.

## [1.1.1] - 2026-06-07

### Changed

- Extracted `.env` file parsing into a shared `internal/envfile` package used by
  both `import` and `check`, removing duplicated logic. No user-facing change.

## [1.1.0] - 2026-06-07

### Added

- New commands: `copy`, `open`, `generate`, `env`, `edit`, `search`, `rename`,
  `move`, `ns` (list/delete/rename), and `status`.
- `set --stdin` reads the value from stdin so secrets stay out of shell history.
- `set KEY=VALUE` form in addition to `set KEY VALUE`.
- Automatic namespace detection from the current git repository, with a
  priority order of flag, `SVAULT_NS`, git repo, config, then `default`.
- File locking on every write, so concurrent svault processes can no longer
  clobber each other and silently lose a secret.
- Key validation: secret keys must be valid shell variable names, keeping
  `export` and `env` output safe to eval.
- Timestamped rollback backups (up to five kept) on every write.
- Tests for the command layer and shell completion documentation.

### Changed

- All user-facing output translated to English.
- Documentation split into README, CONTRIBUTING, and SECURITY.

### Fixed

- CI and release workflows now use the Go version declared in `go.mod`.

## [1.0.0] - 2026-06-05

### Added

- Initial release: `init`, `unlock`, `lock`, `set`, `get`, `delete`, `list`,
  `use`, `export`, `import`, `check`, `exec`, `backup`, `restore`, `diff`,
  `rotate`, `info`, and `log`.
- AES-256-GCM encryption with Argon2id key derivation.
- Namespace support and a session model with a configurable TTL.
- Append-only audit log.

[2.2.1]: https://github.com/dafagareth/svault/releases/tag/v2.2.1
[2.2.0]: https://github.com/dafagareth/svault/releases/tag/v2.2.0
[2.1.0]: https://github.com/dafagareth/svault/releases/tag/v2.1.0
[2.0.1]: https://github.com/dafagareth/svault/releases/tag/v2.0.1
[2.0.0]: https://github.com/dafagareth/svault/releases/tag/v2.0.0
[1.1.1]: https://github.com/dafagareth/svault/releases/tag/v1.1.1
[1.1.0]: https://github.com/dafagareth/svault/releases/tag/v1.1.0
[1.0.0]: https://github.com/dafagareth/svault/releases/tag/v1.0.0
