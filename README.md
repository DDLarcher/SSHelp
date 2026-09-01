# SSHelp

A quick project to help manage and store SSH connections, written in Go.

Terminal-based SSH profile manager with AES-256-GCM encrypted storage. Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Security

- **AES-256-GCM** encryption for the profile store, with **Argon2id** key derivation (64 MiB, 3 passes). Stores written by earlier versions used PBKDF2 and are re-encrypted the first time they are opened
- The store lives in the user's own configuration directory (`~/.config/sshelp/profiles.json`, `%AppData%\sshelp` on Windows) with mode `0600`, and is written through a temporary file so an interrupted save cannot destroy it. Older versions kept it next to the executable, where a shared or synced folder would expose it; such a file is read once and migrated
- Master password required on every launch
- Input sanitization: only `a-z A-Z 0-9 . - _ @ :` allowed in all text fields
- Command injection prevention on SSH launch

## Collections

A profile can belong to a named collection (`production`, `staging`, a customer
name...). The list groups profiles under a heading per collection, with
collections in alphabetical order first and ungrouped profiles last.

A collection is created simply by typing its name in the `Collection` field of a
profile, and disappears when its last profile leaves it. Profile names only have
to be unique **within** a collection, so `web` can exist in both `production` and
`staging`.

## Saved passwords

A profile can store its SSH password, which SSHelp fills in on connect. Because
nobody is watching when that happens, the feature is deliberately fenced in.

**The host key must be verified first.** A password can only be saved for a host
that is already in your `known_hosts`, so the first connection is an ordinary
one where `ssh` shows you the fingerprint and waits for you. The matching
`known_hosts` entries are then pinned to the profile, and every connection that
uses the stored password runs with `StrictHostKeyChecking=yes` against a private
`known_hosts` containing only those pinned keys. A host that cannot prove it
holds one of them is refused before authentication, so a spoofed DNS record or a
hijacked hostname never gets the password. The pinned fingerprints are shown in
the profile details; if the server legitimately rotates its key, update your own
`known_hosts` the usual way and save the profile again to re-pin.

**The password is not passed in the environment or on a command line.** SSHelp
writes it to a file readable only by you, and points OpenSSH at itself through
`SSH_ASKPASS`. The helper deletes the file as it reads it, so the password is
handed out at most once per connection and is gone from disk seconds later; a
file left behind by a connection that never asked expires after two minutes and
is swept at the next launch.

**Only a password prompt is answered.** The prompt text is compared against the
exact wordings OpenSSH uses (`…'s password:`, `password:`, `enter passphrase for
key …`). Under keyboard-interactive the prompt comes from the remote server, so
anything else it asks — a verification code, a second prompt after the password
— gets no answer, and host key confirmations are never auto-approved.

Public key authentication is left enabled: if an agent key or `IdentityFile`
gets you in, the password is never sent.

Leave the password empty to keep the normal behaviour of being prompted by
`ssh`. Requires OpenSSH 8.4+ (for `SSH_ASKPASS_REQUIRE`); with an older client
the password is simply not auto-filled and `ssh` prompts as usual.

### What this does not protect against

- Anything already running as your user can read the credential file during the
  short window it exists, and a `ProxyCommand` from a hostile `~/.ssh/config` is
  started by `ssh` early enough to do so.
- Passwords live in memory as Go strings while SSHelp runs and cannot be wiped
  reliably; they may reach a core dump or swap.

## Requirements

- Go 1.21+
- SSH client installed on the system, `ssh-keygen` included (OpenSSH 8.4+ for saved-password auto-fill)
- Windows or Linux/macOS

## Build from source

```bash
git clone https://github.com/DDLarcher/SSHelp.git
cd SSHelp
go mod tidy
go build -ldflags="-s -w" -trimpath -o SSHelp .    # Linux/macOS
go build -ldflags="-s -w" -trimpath -o SSHelp.exe . # Windows
```

## Usage

On first launch, create a master password. On subsequent launches, enter it to unlock.

Each profile card shows `(key)` when an identity file is set and `(pwd)` when a
password is saved. Profiles are listed grouped by collection.

### Key bindings

| Key | Action |
|-----|--------|
| `a` | Add new profile |
| `↑↓` `j` `k` | Navigate profiles |
| `Enter` / click | Open action menu on profile |
| `C` | Connect via SSH (new window on Windows; in the current terminal on Linux/macOS, returning to SSHelp when the session ends) |
| `E` | Edit profile |
| `D` | Delete profile |
| `V` | View details |
| `Esc` | Go back / cancel |
| `q` | Quit |
