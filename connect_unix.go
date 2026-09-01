//go:build !windows

package main

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Suspends the TUI and runs ssh in the current terminal; the TUI
// resumes when the session ends.
func connectSSHCmd(p Profile) tea.Cmd {
	c := exec.Command("ssh", sshArgs(p)...)
	c.Env = sshEnv(p)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return sshFinishedMsg{err}
	})
}
