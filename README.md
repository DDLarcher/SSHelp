# SSHelp

A quick project to help manage and store SSH connections, written in Go.

Terminal-based SSH profile manager with AES-256-GCM encrypted storage. Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Security

- **AES-256-GCM** encryption for `profiles.json` with PBKDF2 key derivation (100,000 iterations)
- Master password required on every launch
- Input sanitization: only `a-z A-Z 0-9 . - _ @ :` allowed in all text fields
- Command injection prevention on SSH launch
- Saved SSH passwords are stored only inside the encrypted `profiles.json`, are never shown in the UI (only `••••••••`), and never appear on a command line

## Saved passwords

A profile can store its SSH password. On connect, SSHelp hands it to OpenSSH
through `SSH_ASKPASS`: `ssh` runs the SSHelp binary as its prompt helper and
receives the password over a pipe, with the value passed in the environment
rather than on the command line (so it is not visible in `ps`).

The helper answers **password and passphrase prompts only**. It never answers a
host key confirmation or a 2FA prompt, so nothing is auto-approved on your
behalf. Because the host key question can therefore not be answered
interactively, profiles with a saved password connect with
`StrictHostKeyChecking=accept-new`: an unknown host is added to `known_hosts`
automatically, while a **changed** host key still aborts the connection.

Leave the password empty to keep the normal behaviour of being prompted by
`ssh`. Requires OpenSSH 8.4+ (for `SSH_ASKPASS_REQUIRE`); with an older client
the password is simply not auto-filled and `ssh` prompts as usual.

## Requirements

- Go 1.21+
- SSH client installed on the system (OpenSSH 8.4+ for saved-password auto-fill)
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
password is saved.

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
