//go:build !windows

package sshconn

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DDLarcher/SSHelp/internal/profile"
)

// Suspends the TUI and runs ssh in the current terminal; the TUI
// resumes when the session ends.
func Connect(p profile.Profile) tea.Cmd {
	setup, err := prepareConnection(p)
	if err != nil {
		return func() tea.Msg { return FinishedMsg{err} }
	}

	c := exec.Command("ssh", setup.args...)
	c.Env = setup.env
	return tea.ExecProcess(c, func(err error) tea.Msg {
		setup.cleanup()
		return FinishedMsg{err}
	})
}
