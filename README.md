# svault

[![CI](https://github.com/dafagareth/svault/actions/workflows/ci.yml/badge.svg)](https://github.com/dafagareth/svault/actions/workflows/ci.yml)
[![Release](https://github.com/dafagareth/svault/actions/workflows/release.yml/badge.svg)](https://github.com/dafagareth/svault/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](go.mod)
[![AUR](https://img.shields.io/aur/version/svault)](https://aur.archlinux.org/packages/svault)
[![Homebrew](https://img.shields.io/badge/Homebrew-dafagareth%2Ftap-fbb040?logo=homebrew&logoColor=white)](https://github.com/dafagareth/homebrew-tap)
[![Scoop](https://img.shields.io/badge/Scoop-svault-4880EC?logo=windows&logoColor=white)](https://github.com/dafagareth/svault/tree/main/bucket)

**A secure, zero-dependency, local encrypted secret vault for developers.**

`svault` stores environment secrets locally in a single encrypted vault file. It eliminates the risk of committing unencrypted `.env` files to version control, operating with zero external server or cloud dependencies.

```bash
$ svault set DB_PASSWORD supersecret123
OK  DB_PASSWORD saved

$ svault get DB_PASSWORD
supersecret123
```

---

## 🌟 Why `svault`?

| Problem | Solution |
| :--- | :--- |
| **Accidental `.env` Git Commits** | All secrets encrypted at rest using **AES-256-GCM** |
| **Overly Complex Cloud Vaults** | Single static binary with zero runtime dependencies |
| **Secret Collisions Across Projects** | Automatic Git-repository-based namespace isolation |
| **Unmonitored Secret Access** | Append-only local audit log tracking every operation |

For detailed information regarding threat modeling and cryptography design, review [SECURITY.md](SECURITY.md).

---

## ✨ Features

- **Authenticated Encryption**: Encrypted with **AES-256-GCM** and key derivation via **Argon2id**.
- **Namespace Isolation**: Manage secrets per project or environment with automatic Git repository detection.
- **Session Authentication**: Temporary session caching (default 30 minutes) to eliminate repetitive password prompts.
- **Zero-File Ingestion (`exec`)**: Inject secrets directly into process environment variables without writing plaintext files to disk.
- **Safe Clipboard Utility**: Copy secrets with automated clipboard clearing after a configurable timeout (default 30s).
- **Format Interoperability**: Effortless import and export with standard `.env`, JSON, and YAML formats.
- **Audit Traceability**: Immutable local operation logging without storing sensitive secret values.
- **Cross-Platform**: Single binary distribution for Linux, macOS, and Windows.

---

## 📦 Installation

### Package Managers

#### Arch Linux (AUR)
```bash
yay -S svault
# or
paru -S svault
```

#### Homebrew (macOS / Linux)
```bash
brew install dafagareth/tap/svault
```

#### Scoop (Windows)
```powershell
scoop bucket add svault https://github.com/dafagareth/svault
scoop install svault
```

---

### One-Line Install Scripts

**Linux / macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/dafagareth/svault/main/install.sh | sh
```
*Detects OS/architecture, verifies checksums, and installs to `/usr/local/bin` or `~/.local/bin`. Set `SVAULT_VERSION=v2.0.1` or `SVAULT_BIN_DIR` to customize installation.*

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/dafagareth/svault/main/install.ps1 | iex
```

---

### Pre-built Binaries & Source Build

Direct pre-compiled binaries are available on the [Releases](https://github.com/dafagareth/svault/releases) page for `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`, and `windows-amd64`.

**Building from Source:**
*Requirements: **Go 1.25+***

```bash
git clone https://github.com/dafagareth/svault.git
cd svault
sudo make install       # Installs binary to /usr/local/bin
```

---

## 🚀 Quick Start

### 1. Initialize Vault
Initialize your encrypted storage (required once):
```bash
$ svault init
Master password: ********
Confirm password: ********
Vault initialized at ~/.svault/vault.enc
```

### 2. Unlock Session
Unlock the vault for interactive use (valid for 30 minutes by default):
```bash
$ svault unlock
Master password: ********
Vault unlocked. Session valid for 30 minutes.
```

### 3. Store & Retrieve Secrets
```bash
# Save secret via arguments or KEY=VALUE syntax
$ svault set DB_PASSWORD supersecret123
$ svault set JWT_SECRET=myjwtsecret

# Retrieve secret
$ svault get DB_PASSWORD
supersecret123

# List stored keys in active namespace
$ svault list
DB_PASSWORD
JWT_SECRET
```

### 4. Execute Workloads & Lock
```bash
# Inject secrets directly into application processes
$ svault exec -- npm start
$ svault exec -- python manage.py runserver

# Lock session when finished
$ svault lock
```

---

## 🏷️ Namespaces & Automatic Detection

`svault` segregates secrets into namespaces. When executed within a Git repository, `svault` automatically detects the repository name and selects the corresponding namespace.

### Namespace Resolution Order
1. The `--ns` flag specified on command invocation
2. The `SVAULT_NS` environment variable
3. Current Git repository directory name
4. Active namespace configured via `svault use`
5. Default namespace (`default`)

```bash
svault use production              # Switch active namespace
svault set --ns staging DB_URL=... # Target specific namespace explicitly
svault ns list                     # View all namespaces and key counts
svault diff staging production     # Compare secrets between namespaces
```

---

## 📖 Command Reference

Run `svault <command> --help` for complete subcommand usage and flag documentation.

### Authentication & Session Management
- `svault init`: Create a new encrypted vault archive.
- `svault unlock`: Authenticate master password and initiate session.
- `svault lock`: Terminate active session immediately.
- `svault status`: Display session lifecycle and vault statistics.

### Secret Operations
- `svault set <KEY> [VALUE]`: Store secret (`--stdin` or `--clip` flags supported).
- `svault get <KEY...>`: Display secret value (`--clip` to copy to clipboard).
- `svault edit <KEY>`: Modify secret content via `$EDITOR`.
- `svault delete <KEY>`: Permanently remove a secret.
- `svault list`: List keys (`-v` values, `-m` masked, `-l` long format, `--json` JSON).
- `svault search <PATTERN>`: Search key names across namespaces.
- `svault rename <OLD> <NEW>`: Rename key within active namespace.
- `svault move <KEY> --to <NS>`: Relocate key to another namespace.

### Integration & Utilities
- `svault exec -- <COMMAND>`: Execute process with injected vault environment variables.
- `svault env`: Output environment variables (`-f shell|dotenv|json|yaml`).
- `svault generate`: Generate cryptographically secure passwords (`crypto/rand`).
- `svault copy <KEY>`: Copy secret to clipboard with auto-clear timeout.
- `svault open <KEY>`: Open associated service URL and copy secret token/password.
- `svault export`: Export secrets to `.env` format.
- `svault import <FILE>`: Import secrets from `.env` format file.
- `svault check <FILE>`: Verify namespace against reference file (e.g. `.env.example`).

### Maintenance & Auditing
- `svault rotate`: Change master password and re-encrypt vault contents.
- `svault backup [PATH]`: Create timestamped or target backup file.
- `svault restore <PATH>`: Restore vault state from backup file.
- `svault log`: View immutable operational audit logs.
- `svault doctor`: Run environment health and security permission diagnostics.

---

## 🔒 Storage Architecture

Default local storage directory structure:
```text
~/.svault/
├── vault.enc       # AES-256-GCM encrypted secret payload
├── vault.log       # Append-only operational audit log
└── config.json     # Configuration and active namespace state

/tmp/
└── .svault_session # Temporary session token (chmod 0600, auto-purged on TTL expiry)
```

Encrypted storage packet payload layout:
```text
[ 16-byte Salt ] [ 12-byte Nonce ] [ Encrypted Ciphertext + Auth Tag ]
```

---

## ⚙️ Configuration Variables

| Environment Variable | Default | Description |
| :--- | :--- | :--- |
| `SVAULT_SESSION_TTL` | `30` | Session validity duration (in minutes) |
| `SVAULT_NS` | *(unset)* | Override active namespace selection |

---

## 🐚 Shell Completion

Generate completion scripts for your shell environment:

```bash
# Bash
svault completion bash | sudo tee /etc/bash_completion.d/svault > /dev/null

# Zsh
svault completion zsh > "${fpath[1]}/_svault"

# Fish
svault completion fish > ~/.config/fish/completions/svault.fish
```

---

## 📄 License & Community

- **License**: [MIT License](LICENSE)
- **Security Protocols**: Refer to [SECURITY.md](SECURITY.md) for vulnerability disclosure procedures.
- **Contribution Guide**: Refer to [CONTRIBUTING.md](CONTRIBUTING.md) for development workflows.
- **Code of Conduct**: Refer to [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

Copyright © 2026 [Dafa](https://github.com/dafagareth).
