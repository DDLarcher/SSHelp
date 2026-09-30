package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DDLarcher/SSHelp/internal/sshconn"
	"github.com/DDLarcher/SSHelp/internal/tui"
)

func main() {
	// OpenSSH runs SSH_ASKPASS with the prompt as its only argument. Requiring
	// that argument means a bare launch always starts the TUI, so a stray
	// SSHELP_ASKPASS_FILE in the environment cannot silently turn the binary
	// into a helper.
	if cred := os.Getenv(sshconn.AskpassFileEnv); cred != "" && len(os.Args) == 2 {
		os.Exit(sshconn.Askpass(cred, os.Args[1]))
	}

	sshconn.SweepCredentials()

	p := tea.NewProgram(tui.New(), tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
