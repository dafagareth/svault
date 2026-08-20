# Security

This document describes how `svault` protects your secrets, the trade-offs it makes, and how to report a vulnerability.

## Cryptography

- **Encryption**: AES-256-GCM, an authenticated cipher. Any tampering with the vault file is detected on decryption.
- **Key derivation**: Argon2id, the recommended algorithm for password-based key derivation. Each vault uses a fresh random 16 byte salt.
- **Randomness**: all salts, nonces, and generated passwords use `crypto/rand`. The package never uses `math/rand` for anything security related.

The `vault.enc` file layout is:

```
[ 16 bytes salt ][ 12 bytes nonce ][ ciphertext ]
```

The salt is stored in the clear, which is standard practice. It is not secret; its purpose is to make precomputed attacks against the password infeasible.

## Guarantees

- **The master password cannot be recovered.** If you forget it, there is no way into the vault. This is by design. Keep a backup of your password somewhere safe.
- Secret values are **never** written to logs, temp files, or any output other than `svault get`, `svault copy`, `svault env`, and `svault export`.
- Every write to the vault first creates a `vault.enc.bak` backup plus several timestamped rollback copies, so a bad write cannot destroy an earlier good copy.
- Concurrent svault processes serialize on an exclusive file lock, so two writes can never clobber each other and silently lose a secret.
- The session token file is created with mode `0600`, readable only by the owning user.
- Secret keys are validated to be valid shell variable names, so `svault export` and `svault env` output is always safe to eval.

## Known trade-off: session key in /tmp

When the vault is unlocked, the derived key is stored as plaintext in `/tmp/.svault_session` with permission `0600`. This is a deliberate trade-off between convenience (not retyping the password constantly) and security.

Implications:

- Other users on the same system cannot read this file, since it is protected by `0600` permissions.
- **Root, or any process with root-equivalent privileges, can read this file.**
- The file is deleted automatically after the TTL (default 30 minutes) or when `svault lock` runs.

If you work on a shared server or an environment where others have root access, run `svault lock` as soon as you finish using the vault rather than relying on the TTL to expire.

## Reporting a vulnerability

If you find a security issue, do not open a public GitHub issue.

Instead, email **dafagareth@gmail.com** with:

- A description of the issue
- Steps to reproduce
- The potential impact

You can expect an acknowledgement within a few days.
