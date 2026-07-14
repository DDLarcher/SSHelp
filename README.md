# SSHelp

Terminal-based SSH profile manager. Save your connections, click and connect.

## Features

- **Profile management** — save user@host:port, key paths, and connection details
- **Quick connect** — select a profile, press Enter, C — opens SSH in a new window
- **Mouse + keyboard** — full mouse click support plus Vim-style navigation
- **Dark theme** — red accent on black, hacker aesthetic
- **Persistent storage** — profiles saved to `profiles.json` alongside the binary

## Usage

Run `SSHelp.exe` from a terminal:

```
SSHelp.exe
```

### Key bindings

| Key | Action |
|-----|--------|
| `a` | Add new profile |
| `↑↓` / `j` `k` | Navigate profiles |
| `Enter` / click | Open action menu |
| `C` | Connect (SSH) |
| `E` | Edit profile |
| `D` | Delete profile |
| `V` | View details |
| `q` | Quit |

## How it works

Profiles are stored in `profiles.json` next to the executable. On connect, SSHelp launches `ssh` in a new terminal window.

### Example profile

```json
{
  "name": "Production",
  "user": "root",
  "host": "server.example.com",
  "port": 2222,
  "key_path": "~/.ssh/id_rsa"
}
```

## Build from source

```bash
go mod tidy
go build -o SSHelp.exe .
```

Requires Go 1.21+.
