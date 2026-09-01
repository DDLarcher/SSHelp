package main

import (
	"crypto/sha256"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Name a host is filed under in known_hosts: bare for port 22, [host]:port
// otherwise.
func knownHostsName(host string, port int) string {
	if port == 0 || port == 22 {
		return host
	}
	return "[" + host + "]:" + strconv.Itoa(port)
}

// Hostname and port ssh will actually connect to once ~/.ssh/config has been
// applied, so that a host alias is looked up under the name known_hosts uses.
func resolvedTarget(p Profile) (string, int) {
	host, port := p.Host, p.Port
	out, err := exec.Command("ssh", append([]string{"-G"}, sshTarget(p)...)...).Output()
	if err != nil {
		return host, port
	}
	for _, line := range strings.Split(string(out), "\n") {
		key, val, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok {
			continue
		}
		switch key {
		case "hostname":
			host = val
		case "port":
			if n, err := strconv.Atoi(val); err == nil {
				port = n
			}
		}
	}
	return host, port
}

// known_hosts entries the user has already accepted for this host. Empty when
// the host has never been connected to, which is what stops a password from
// being saved for a host whose identity nobody has verified yet.
func importHostKeys(p Profile) []string {
	host, port := resolvedTarget(p)
	out, err := exec.Command("ssh-keygen", "-F", knownHostsName(host, port)).Output()
	if err != nil {
		return nil
	}

	var keys []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !isValidHostKey(line) {
			continue
		}
		keys = append(keys, line)
	}
	return keys
}

// Writes the pinned entries to a private known_hosts of our own, so the profile
// is bound to the key that was verified when the password was saved rather than
// to whatever ~/.ssh/known_hosts happens to hold at connect time.
func writePinnedKnownHosts(p Profile) (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "known_hosts-"+fileSafe(p.GroupLabel()+"-"+p.Name))
	if err := writeFilePrivate(path, []byte(strings.Join(p.HostKeys, "\n")+"\n")); err != nil {
		return "", err
	}
	return path, nil
}

// Profile names allow ':', which is not a legal filename character everywhere.
func fileSafe(s string) string {
	return strings.ReplaceAll(s, ":", "_")
}

// SHA256 fingerprints of the pinned keys, in the format ssh itself prints, so
// the user can compare them with what the server reports.
func hostKeyFingerprints(p Profile) []string {
	var out []string
	for _, entry := range p.HostKeys {
		parts := strings.Fields(entry)
		if len(parts) < 3 {
			continue
		}
		blob, err := base64.StdEncoding.DecodeString(parts[2])
		if err != nil {
			continue
		}
		sum := sha256.Sum256(blob)
		out = append(out, parts[1]+" SHA256:"+strings.TrimRight(base64.StdEncoding.EncodeToString(sum[:]), "="))
	}
	return out
}

// Removes credential files left behind by a connection that was never asked for
// a password (public key auth succeeded, host unreachable) and whose process is
// long gone.
func sweepCredentials() {
	dir, err := stateDir()
	if err != nil {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), credPrefix) {
			continue
		}
		info, err := e.Info()
		if err == nil && time.Since(info.ModTime()) < credLifetime {
			continue
		}
		os.Remove(filepath.Join(dir, e.Name()))
	}
}

// A password may only be stored for a host whose key the user has already
// verified. The matching known_hosts entries are copied into the profile, so
// the password is later offered only to a host that proves it holds that key.
func pinHostKeys(p Profile) (Profile, error) {
	if p.Password == "" {
		p.HostKeys = nil
		return p, nil
	}

	keys := importHostKeys(p)
	if len(keys) == 0 {
		return p, &ProfileError{"unknown host key for " + p.Host + ": connect once without a saved password, check the fingerprint, then save the password"}
	}

	p.HostKeys = keys
	return p, nil
}
