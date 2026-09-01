//go:build windows

package main

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Opens the SSH session in a new console window via `cmd /c start`.
func connectSSHCmd(p Profile) tea.Cmd {
	return func() tea.Msg {
		setup, err := prepareConnection(p)
		if err != nil {
			return sshFinishedMsg{err}
		}

		safeName := sanitizeForShell(p.Name)
		if safeName == "" {
			safeName = "SSH"
		}

		c := exec.Command("cmd", append([]string{"/c", "start", "SSH: " + safeName, "ssh"}, setup.args...)...)
		c.Env = setup.env
		if err := c.Start(); err != nil {
			setup.cleanup()
			return sshFinishedMsg{err}
		}
		// The session outlives this call in its own console window, so the
		// credential file is left to the helper and to its expiry timer.
		return sshFinishedMsg{nil}
	}
}
