//go:build windows

package sshconn

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DDLarcher/SSHelp/internal/profile"
)

// Opens the SSH session in a new console window via `cmd /c start`.
func Connect(p profile.Profile) tea.Cmd {
	return func() tea.Msg {
		setup, err := prepareConnection(p)
		if err != nil {
			return FinishedMsg{err}
		}

		safeName := sanitizeForShell(p.Name)
		if safeName == "" {
			safeName = "SSH"
		}

		c := exec.Command("cmd", append([]string{"/c", "start", "SSH: " + safeName, "ssh"}, setup.args...)...)
		c.Env = setup.env
		if err := c.Start(); err != nil {
			setup.cleanup()
			return FinishedMsg{err}
		}
		// The session outlives this call in its own console window, so the
		// credential file is left to the helper and to its expiry timer.
		return FinishedMsg{nil}
	}
}

// Strips the characters cmd.exe treats specially from the window title.
func sanitizeForShell(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"', '&', '|', ';', '!', '%', '^', '`', '\n', '\r':
			continue
		default:
			b = append(b, c)
		}
	}
	return string(b)
}
