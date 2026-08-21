# Technical Documentation

This document provides technical reference for installation, command invocation, namespace resolution, and storage architecture for `svault`.

## Installation Methods

### Arch Linux (AUR)
```bash
yay -S svault
```

### Homebrew
```bash
brew install dafagareth/tap/svault
```

### Scoop
```powershell
scoop bucket add svault https://github.com/dafagareth/svault
scoop install svault
```

### Scripted Installation

Linux and macOS:
```bash
curl -fsSL https://raw.githubusercontent.com/dafagareth/svault/main/install.sh | sh
```

Windows:
```powershell
irm https://raw.githubusercontent.com/dafagareth/svault/main/install.ps1 | iex
```

### Source Compilation

Requires Go 1.25 or higher.

```bash
git clone https://github.com/dafagareth/svault.git
cd svault
sudo make install
```

## Namespace Resolution Hierarchy

`svault` isolates secrets using namespaces. Resolution order from highest to lowest precedence:

1. Explicit `--ns` flag provided at command invocation.
2. `SVAULT_NS` environment variable.
3. Current Git repository root directory name (detected automatically).
4. Active namespace persisted via `svault ns use` (stored in `~/.svault/config.json`).
5. Default namespace named `default`.

## Command Reference

### Session and Authentication

`svault init`
Initializes a new encrypted vault at `~/.svault/vault.enc`. Prompts for a master password with confirmation.

`svault unlock`
Authenticates the master password and writes a session token to `~/.svault/.session`, valid for 30 minutes.

`svault lock`
Removes the active session token immediately.

`svault status`
Reports session state, remaining session time, namespace count, total key count, and vault file size. Accepts `--short` for a compact one-line format suitable for shell prompts.

### Secret Operations

`svault set KEY [VALUE]`
Stores a secret in the active namespace. Accepts `KEY=VALUE` form, `--stdin` to read from standard input, or `--clip` to read from the clipboard. Passing values as CLI arguments exposes them in the process table; prefer `--stdin` or `--clip`.

`svault get KEY [KEY...]`
Prints stored values. A single key prints only the value; multiple keys print `KEY=VALUE` per line. Accepts `--clip` to copy to clipboard with an automatic expiry timeout (default 20 seconds, configurable via `--timeout`).

`svault edit KEY`
Opens the secret value in the editor defined by `$EDITOR`.

`svault delete KEY`
Removes the specified key and its value from the active namespace.

`svault list`
Lists keys in the active namespace. Accepts `-v` (values), `-m` (masked), `-l` (long format with lengths), `-a` (all namespaces), `-c` (count only), `-F PATTERN` (filter), and `--json`.

`svault search PATTERN`
Searches key names by substring across the active namespace or all namespaces when `--all` is supplied.

`svault rename OLD NEW`
Renames a key while preserving its value.

`svault move KEY --to NAMESPACE`
Transfers a key and its value to the specified target namespace.

### Namespaces

`svault ns list`
Lists all namespaces with their key counts. Marks the active namespace.

`svault ns delete NAMESPACE`
Deletes a namespace and all its secrets. The `default` namespace cannot be deleted.

`svault ns rename OLD NEW`
Renames a namespace.

`svault use NAMESPACE`
Sets the active namespace and persists it to `~/.svault/config.json`.

`svault diff NAMESPACE1 NAMESPACE2`
Compares two namespaces, reporting keys present in one but absent in the other.

### Environment and Integration

`svault exec COMMAND [ARGS...]`
Executes the specified command with vault secrets injected as environment variables. Secrets are never written to disk.

`svault env`
Outputs environment variable declarations. Accepts `-f` to select format: `shell`, `dotenv`, `json`, or `yaml`.

`svault export`
Exports secrets from the active namespace as environment variable lines.

`svault import FILE`
Imports key-value pairs from a `.env` file into the active namespace.

`svault check FILE`
Compares the active namespace against a reference file such as `.env.example`, reporting missing or extra keys.

### Utilities and Password Generation

`svault generate`
Generates a cryptographically secure random password using `crypto/rand`. Accepts `--length` (default 24), `--no-symbols`, `--save KEY` to persist in vault, and `--no-copy` to suppress clipboard write.

`svault copy KEY`
Copies the secret value to the system clipboard and schedules a clear after 30 seconds.

`svault open KEY`
Resolves `KEY_URL` (or `KEY` if it starts with `http`) and opens it in the default browser. Copies `KEY_PASS`, `KEY_PASSWORD`, or `KEY_TOKEN` to clipboard if present.

### Maintenance and Diagnostics

`svault rotate`
Re-encrypts the vault with a new master password and refreshes the active session.

`svault backup [FILE]`
Creates a copy of the encrypted vault. Generates a timestamped filename in `~/.svault/` if no path is given.

`svault restore FILE`
Restores the vault from a backup file using an atomic write. Clears the active session.

`svault log`
Displays the append-only operational audit log. Accepts `--tail N` to show the last N entries.

`svault doctor`
Performs environment diagnostics, verifying file permissions and installation integrity.

`svault info`
Displays vault metadata including creation timestamp, namespace count, total key count, and storage file sizes.

`svault version`
Prints the installed `svault` version string.

`svault diff`
Compares the active namespace against another namespace or a reference file, reporting keys present in one but absent in the other.

### Namespace Management

`svault ns list`
Lists all namespaces stored in the vault.

`svault ns delete <NS>`
Removes the specified namespace and all secrets it contains.

`svault ns rename <OLD> <NEW>`
Renames a namespace.

`svault use <NS>`
Sets the active namespace (alias for `svault ns use`). Writes `active_namespace` to `config.json` (precedence 4 in the resolution hierarchy).

`svault completion SHELL`
Generates shell tab-completion scripts for `bash`, `zsh`, `fish`, or `powershell`.

## Storage Architecture

The local state directory `~/.svault/` contains:
- `vault.enc`: AES-256-GCM encrypted secret payload.
- `vault.log`: Append-only audit log with key name masking.
- `config.json`: Persistent configuration storing the active namespace name.
- `.session`: Active session token (permissions 0600, TTL default 30 minutes).

All write operations use atomic file replacement (`CreateTemp`, `Sync`, `Rename`) to guarantee consistency on crash or power loss.

Encrypted binary payload layout of `vault.enc`:
`[ 16-byte Salt ][ 12-byte Nonce ][ Ciphertext + Auth Tag ]`
