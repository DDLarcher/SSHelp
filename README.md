# SSHelp

A quick project to help manage and store SSH connections, written in Go.

Terminal-based SSH profile manager. Save your connections, click and connect. Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Requirements

- Go 1.21+
- SSH client installed on the system (built-in on Windows 10+/Linux/macOS)
- A terminal with mouse support (Windows Terminal, iTerm2, Kitty, etc.)

## Install

```bash
git clone https://github.com/DDLarcher/SSHelp.git
cd SSHelp
go mod tidy
go build -o SSHelp.exe .
```

Then run `SSHelp.exe` from your terminal.

## Usage

Launch the program. You'll see the main screen with the SSHelp banner.

### Adding a profile

Press `a` to start the add wizard:

1. **Profile name** — a label for the connection
2. **Username** — e.g. `root`, `admin`, `deploy`
3. **Host** — IP or domain, e.g. `server.example.com`
4. **Port** — SSH port (defaults to 22)
5. **Key path** — optional, path to your SSH private key
6. Review and save

### Connecting

1. Use `↑↓` or click a profile to select it
2. Press `Enter` or click it again to open the action menu
3. Press `C` (or click Connect) — a new terminal window opens with the SSH session

### Editing / Deleting

- `E` from the action menu — modify any field then `Ctrl+S` to save
- `D` from the action menu — confirm deletion
- `V` for full profile details

### Key bindings

| Key | Action |
|-----|--------|
| `a` | Add new profile |
| `↑↓` `j` `k` | Navigate profiles |
| `Enter` / click | Open action menu on profile |
| `C` | Connect via SSH (new window) |
| `E` | Edit profile |
| `D` | Delete profile |
| `V` | View details |
| `Esc` | Go back / cancel |
| `q` | Quit |

## How it works

Profiles are stored in `profiles.json` next to the executable:

```json
[
  {
    "name": "Production",
    "user": "root",
    "host": "server.example.com",
    "port": 2222,
    "key_path": "~/.ssh/id_rsa",
    "last_access": "2026-07-14 09:00"
  }
]
```

On connect, SSHelp launches the system SSH client in a new terminal window:

```
cmd /c start "SSH: Production" ssh -p 2222 root@server.example.com
```

## Build from source

```bash
git clone https://github.com/DDLarcher/SSHelp.git
cd SSHelp
go mod tidy
go build -o SSHelp.exe .
```

For a smaller binary, use:

```bash
go build -ldflags="-s -w" -o SSHelp.exe .
```
