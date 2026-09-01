//go:build !windows

package main

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Suspends the TUI and runs ssh in the current terminal; the TUI
// resumes when the session ends.
func connectSSHCmd(p Profile) tea.Cmd {
	setup, err := prepareConnection(p)
	if err != nil {
		return func() tea.Msg { return sshFinishedMsg{err} }
	}

	c := exec.Command("ssh", setup.args...)
	c.Env = setup.env
	return tea.ExecProcess(c, func(err error) tea.Msg {
		setup.cleanup()
		return sshFinishedMsg{err}
	})
}
