//go:build windows

package main

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Opens the SSH session in a new console window via `cmd /c start`.
func connectSSHCmd(p Profile) tea.Cmd {
	return func() tea.Msg {
		safeName := sanitizeForShell(p.Name)
		if safeName == "" {
			safeName = "SSH"
		}
		args := append([]string{"/c", "start", "SSH: " + safeName, "ssh"}, sshArgs(p)...)
		c := exec.Command("cmd", args...)
		c.Env = sshEnv(p)
		return sshFinishedMsg{c.Start()}
	}
}
