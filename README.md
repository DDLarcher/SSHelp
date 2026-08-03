# SSHelp

A quick project to help manage and store SSH connections, written in Go.

Terminal-based SSH profile manager with AES-256-GCM encrypted storage. Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Security

- **AES-256-GCM** encryption for `profiles.json` with PBKDF2 key derivation (100,000 iterations)
- Master password required on every launch
- Input sanitization: only `a-z A-Z 0-9 . - _ @ :` allowed in all text fields
- Command injection prevention on SSH launch

## Requirements

- Go 1.21+
- SSH client installed on the system
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
