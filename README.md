# svault

[![CI](https://github.com/dafagareth/svault/actions/workflows/ci.yml/badge.svg)](https://github.com/dafagareth/svault/actions/workflows/ci.yml)
[![Release](https://github.com/dafagareth/svault/actions/workflows/release.yml/badge.svg)](https://github.com/dafagareth/svault/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](go.mod)
[![AUR](https://img.shields.io/aur/version/svault)](https://aur.archlinux.org/packages/svault)
[![Homebrew](https://img.shields.io/badge/Homebrew-dafagareth%2Ftap-fbb040?logo=homebrew&logoColor=white)](https://github.com/dafagareth/homebrew-tap)
[![Scoop](https://img.shields.io/badge/Scoop-svault-4880EC?logo=windows&logoColor=white)](https://github.com/dafagareth/svault/tree/main/bucket)

`svault` is a zero-dependency CLI application written in Go that manages encrypted environment secrets locally. It stores secret keys and values in an authenticated vault file, preventing accidental commits of plaintext environment files to version control systems.

```bash
$ svault set DB_PASSWORD supersecret123
OK  DB_PASSWORD saved

$ svault get DB_PASSWORD
supersecret123
```

## Primary Capabilities

- Authenticated encryption using AES-256-GCM with key derivation via Argon2id.
- Isolated namespaces for project management with automatic Git repository detection.
- Session caching in `~/.svault/.session` (mode 0600) to avoid repetitive master password prompts.
- Subprocess environment injection (`svault exec`) without creating plaintext files on disk.
- Interoperability with standard `.env`, JSON, and YAML formats.
- Append-only local audit logging without exposing secret payloads.
- Cross-platform clipboard integration with auto-clear timeout (Wayland, X11, macOS, Windows).

## Quick Start

Initialize the vault:
```bash
$ svault init
```

Unlock a session:
```bash
$ svault unlock
```

Store and retrieve secrets:
```bash
$ svault set DB_PASSWORD supersecret123
$ svault get DB_PASSWORD
```

Execute a process with secrets injected as environment variables:
```bash
$ svault exec -- npm start
```

Lock the session when done:
```bash
$ svault lock
```

## Documentation and Specifications

Refer to [docs.md](docs.md) for full installation methods, command specifications, and storage layout descriptions.

Refer to [SECURITY.md](SECURITY.md) for the cryptographic threat model and security guarantees.

Refer to [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution standards.

## License

MIT License. See [LICENSE](LICENSE) for terms.

Copyright (c) 2026 dafagareth.
