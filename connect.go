package main

import (
	"os"
	"strconv"
)

// Env var carrying the stored password to the askpass helper. Its presence
// also tells main() to run in askpass mode instead of starting the TUI.
const askpassEnv = "SSHELP_ASKPASS_PW"

type sshFinishedMsg struct{ err error }

func sshArgs(p Profile) []string {
	var args []string
	if p.Port > 0 && p.Port != 22 {
		args = append(args, "-p", strconv.Itoa(p.Port))
	}
	if p.KeyPath != "" {
		args = append(args, "-i", p.KeyPath)
	}
	if p.Password != "" {
		// The askpass helper only answers password prompts, so the host key
		// confirmation of an unknown host would abort the login. New hosts are
		// added automatically; a changed key still fails, as it should.
		args = append(args, "-o", "StrictHostKeyChecking=accept-new",
			"-o", "NumberOfPasswordPrompts=1")
		if p.KeyPath == "" {
			args = append(args, "-o", "PreferredAuthentications=password,keyboard-interactive")
		}
	}
	return append(args, p.User+"@"+p.Host)
}

// Environment for the ssh process. When the profile stores a password, ssh is
// told to ask this same binary for it (SSH_ASKPASS) instead of prompting.
// Returns nil to inherit the current environment unchanged.
func sshEnv(p Profile) []string {
	if p.Password == "" {
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	return append(os.Environ(),
		askpassEnv+"="+p.Password,
		"SSH_ASKPASS="+exe,
		"SSH_ASKPASS_REQUIRE=force",
	)
}
