# Security Policy and Cryptographic Architecture

This document specifies the cryptographic design, security guarantees, operational trade-offs, and vulnerability reporting procedures for `svault`.

## Cryptographic Design

`svault` employs authenticated symmetric encryption with AES-256-GCM. Decryption operations verify authentication tags, detecting any data corruption or tampering.

Key derivation uses Argon2id with a 16-byte random salt generated per vault initialization.

Cryptographic randomness, nonces, and password generation rely exclusively on system entropy via `crypto/rand`.

Encrypted storage payload structure:
`[ 16-byte Salt ][ 12-byte Nonce ][ Ciphertext + Auth Tag ]`

The salt is stored unencrypted at the header of the vault payload to prevent precomputed rainbow table attacks.

## Security Guarantees

Master password recovery is impossible by design. Forgetting the master password results in permanent loss of access to vault data.

Secret values are omitted from audit logs, temporary files, and debug outputs. Values are exposed only when executing `get`, `copy`, `env`, or `export` commands.

Key names in audit logs are masked using a SHA-256 hash prefix (`id:xxxxxxxx`) so that plaintext key names never appear in `vault.log`.

Vault modifications perform write operations through exclusive file locks, preventing race conditions between concurrent processes.

All write operations (vault, session, config) use atomic file replacement (`CreateTemp`, `Sync`, `Rename`) to prevent partial writes from corrupting stored data on power loss or process interruption.

Automatic rollback backups are generated prior to vault modifications to protect against data corruption.

Key names are restricted to valid POSIX environment variable names, ensuring environment export operations remain safe for shell evaluation.

## Operational Trade-offs

When unlocked, the derived encryption key is cached at `~/.svault/.session` with permissions `0600` (user-readable only). The directory `~/.svault/` is created with permissions `0700`.

Implications of session caching:
- Permissions restrict access to the owning user only.
- Processes with root privileges can access any file on the system.
- Session files automatically expire after the configured TTL (default 30 minutes, configurable via `SVAULT_SESSION_TTL` environment variable) or upon executing `svault lock`.

Users operating in shared or multi-tenant environments with administrative access should execute `svault lock` immediately after completing work.

## Vulnerability Disclosure Protocol

Report security vulnerabilities via email to `dafagareth@gmail.com`. Include a technical description of the vulnerability, reproduction steps, and potential impact. Do not submit public GitHub issues for security vulnerabilities.
