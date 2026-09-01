package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// Carries the path of the one-shot credential file, never the password
	// itself: the environment of an ssh process is readable through
	// /proc/<pid>/environ by anything running as the same user, and is
	// inherited by every helper ssh spawns.
	askpassFileEnv = "SSHELP_ASKPASS_FILE"

	credPrefix   = "cred-"
	credLifetime = 2 * time.Minute
)

type sshFinishedMsg struct{ err error }

// Everything needed to launch one ssh session, plus the cleanup that removes
// the credential file once the session is over.
type connSetup struct {
	args    []string
	env     []string
	cleanup func()
}

func sshTarget(p Profile) []string {
	var args []string
	if p.Port > 0 && p.Port != 22 {
		args = append(args, "-p", strconv.Itoa(p.Port))
	}
	return append(args, p.User+"@"+p.Host)
}

func sshArgs(p Profile, knownHosts string) []string {
	var args []string
	if p.Port > 0 && p.Port != 22 {
		args = append(args, "-p", strconv.Itoa(p.Port))
	}
	if p.KeyPath != "" {
		args = append(args, "-i", p.KeyPath)
	}
	if p.Password != "" {
		// A stored password is offered without anyone watching, so the host has
		// to prove it is the one whose key was verified when the password was
		// saved. An unknown or changed key aborts before authentication.
		args = append(args, "-o", "StrictHostKeyChecking=yes",
			"-o", "NumberOfPasswordPrompts=1")
		if knownHosts != "" {
			args = append(args, "-o", `UserKnownHostsFile="`+knownHosts+`"`)
		}
	}
	return append(args, p.User+"@"+p.Host)
}

func prepareConnection(p Profile) (connSetup, error) {
	s := connSetup{cleanup: func() {}}

	knownHosts := ""
	if p.Password != "" && len(p.HostKeys) > 0 {
		path, err := writePinnedKnownHosts(p)
		// ssh splits an option value on whitespace unless it is quoted, and a
		// quote in the path itself would break the quoting; in that case fall
		// back to the user's own known_hosts, still under StrictHostKeyChecking.
		if err == nil && !strings.Contains(path, `"`) {
			knownHosts = path
		}
	}
	s.args = sshArgs(p, knownHosts)

	if p.Password == "" {
		return s, nil
	}

	exe, err := os.Executable()
	if err != nil {
		return s, err
	}
	cred, err := writeCredential(p.Password)
	if err != nil {
		return s, err
	}

	s.env = append(os.Environ(),
		askpassFileEnv+"="+cred,
		"SSH_ASKPASS="+exe,
		"SSH_ASKPASS_REQUIRE=force",
	)

	// The helper deletes the file as it reads it; the timer covers the case
	// where ssh never asks at all, so the password does not sit on disk.
	timer := time.AfterFunc(credLifetime, func() { os.Remove(cred) })
	s.cleanup = func() {
		timer.Stop()
		os.Remove(cred)
	}
	return s, nil
}

// The password is handed over in a file readable only by its owner, which the
// askpass helper consumes on first read. It therefore never appears on a
// command line nor in any process environment, and it can be collected only
// once per connection.
func writeCredential(password string) (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	sweepCredentials()

	f, err := os.CreateTemp(dir, credPrefix+"*")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := f.Chmod(fileMode); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	if _, err := f.WriteString(password); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return filepath.Clean(f.Name()), nil
}
