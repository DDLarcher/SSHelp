package sshconn

import (
	"fmt"
	"os"
	"strings"
)

// Answers the first password prompt of a connection and nothing else. The
// credential file is consumed as it is read, so anything asked afterwards - a
// server following the password with a 2FA prompt, say - gets no answer.
func Askpass(credPath, prompt string) int {
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
