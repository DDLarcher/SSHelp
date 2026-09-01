package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// OpenSSH runs SSH_ASKPASS with the prompt as its only argument. Requiring
	// that argument means a bare launch always starts the TUI, so a stray
	// SSHELP_ASKPASS_FILE in the environment cannot silently turn the binary
	// into a helper.
	if cred := os.Getenv(askpassFileEnv); cred != "" && len(os.Args) == 2 {
		os.Exit(askpass(cred, os.Args[1]))
	}

	sweepCredentials()

	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Answers the first password prompt of a connection and nothing else. The
// credential file is consumed as it is read, so anything asked afterwards - a
// server following the password with a 2FA prompt, say - gets no answer.
func askpass(credPath, prompt string) int {
	if !isPasswordPrompt(prompt) {
		return 1
	}

	password, err := os.ReadFile(credPath)
	os.Remove(credPath)
	if err != nil || len(password) == 0 {
		return 1
	}

	fmt.Println(string(password))
	return 0
}

// Under keyboard-interactive the prompt text comes from the remote server, so
// it is matched against the exact wordings OpenSSH and password authentication
// use rather than searched for a substring. A server asking for anything else,
// a verification code included, is refused.
func isPasswordPrompt(prompt string) bool {
	p := strings.ToLower(strings.TrimSpace(prompt))
	return strings.HasPrefix(p, "password:") ||
		strings.HasSuffix(p, "'s password:") ||
		strings.HasPrefix(p, "enter passphrase for key")
}
