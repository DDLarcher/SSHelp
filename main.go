package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// OpenSSH runs SSH_ASKPASS with the prompt as its only argument; we point
	// it at this binary and pass the password through the environment.
	if pw := os.Getenv(askpassEnv); pw != "" {
		os.Exit(askpass(pw, os.Args))
	}

	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Answers password prompts only. Anything else (host key confirmation, 2FA
// codes) is left unanswered so nothing is auto-approved on the user's behalf.
func askpass(password string, args []string) int {
	prompt := ""
	if len(args) > 1 {
		prompt = strings.ToLower(args[1])
	}
	if !strings.Contains(prompt, "password") && !strings.Contains(prompt, "passphrase") {
		return 1
	}
	fmt.Println(password)
	return 0
}
