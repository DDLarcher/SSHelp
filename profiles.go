package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Profile struct {
	Name       string   `json:"name"`
	Group      string   `json:"group,omitempty"`
	User       string   `json:"user"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	KeyPath    string   `json:"key_path,omitempty"`
	Password   string   `json:"password,omitempty"`
	HostKeys   []string `json:"host_keys,omitempty"`
	LastAccess string   `json:"last_access,omitempty"`
}

const maxPasswordLen = 256

var validInputRE = regexp.MustCompile(`^[a-zA-Z0-9._@:-]+$`)

func isValidInput(s string) bool {
	return s != "" && validInputRE.MatchString(s)
}

// Passwords are handed to ssh through the environment, never through a shell,
// so any printable character is fine; control characters are not.
func isValidPassword(s string) bool {
	if utf8.RuneCountInString(s) > maxPasswordLen {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

// A pinned known_hosts entry, as produced by ssh-keygen -F.
func isValidHostKey(s string) bool {
	if s == "" || len(s) > 2048 {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

type ProfileError struct{ msg string }

func (e *ProfileError) Error() string { return e.msg }

func ValidateProfile(p Profile) error {
	if !isValidInput(p.Name) {
		return &ProfileError{"invalid profile name"}
	}
	if !isValidInput(p.User) {
		return &ProfileError{"invalid username"}
	}
	if !isValidInput(p.Host) {
		return &ProfileError{"invalid hostname"}
	}
	if p.Port < 1 || p.Port > 65535 {
		return &ProfileError{"invalid port (1-65535)"}
	}
	if p.Group != "" && !isValidInput(p.Group) {
		return &ProfileError{"invalid collection name"}
	}
	if p.KeyPath != "" && !isValidInput(p.KeyPath) {
		return &ProfileError{"invalid key path characters"}
	}
	if !isValidPassword(p.Password) {
		return &ProfileError{"invalid password (max 256 printable characters)"}
	}
	for _, k := range p.HostKeys {
		if !isValidHostKey(k) {
			return &ProfileError{"invalid pinned host key"}
		}
	}
	return nil
}

func filterValid(profiles []Profile) ([]Profile, int) {
	out := make([]Profile, 0, len(profiles))
	for _, p := range profiles {
		if ValidateProfile(p) == nil {
			out = append(out, p)
		}
	}
	return out, len(profiles) - len(out)
}

func sanitizeForShell(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"', '&', '|', ';', '!', '%', '^', '`', '\n', '\r':
			continue
		default:
			b = append(b, c)
		}
	}
	return string(b)
}

// Where older versions kept the store: next to the executable, which exposes
// it whenever the binary lives in a shared, synced or removable directory.
func legacyProfilesPath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(exe), "profiles.json")
}

// The store now belongs in the user's own configuration directory.
func configProfilesPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "sshelp", "profiles.json")
}

// Saves always go to the new location, so a store left next to the executable
// is migrated the first time anything is written.
func profilesWritePath() string {
	if path := configProfilesPath(); path != "" {
		return path
	}
	return legacyProfilesPath()
}

func profilesReadPath() string {
	if path := configProfilesPath(); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	if old := legacyProfilesPath(); old != "" {
		if _, err := os.Stat(old); err == nil {
			return old
		}
	}
	return profilesWritePath()
}

func joinWarnings(warnings []string) string {
	return strings.Join(warnings, "; ")
}

// Directory for the per-profile pinned known_hosts files and the one-shot
// credential files, created private to the user.
func stateDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "sshelp")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func (p Profile) HostInfo() string {
	info := p.User + "@" + p.Host + ":" + strconv.Itoa(p.Port)
	if p.KeyPath != "" {
		info += " (key)"
	}
	if p.Password != "" {
		info += " (pwd)"
	}
	return info
}

func (p Profile) UpdatedNow() Profile {
	p.LastAccess = time.Now().Format("2006-01-02 15:04")
	return p
}

// Collections come first in alphabetical order, ungrouped profiles last; within
// a collection profiles are sorted by name. listView relies on profiles of the
// same collection being contiguous.
func sortProfiles(profiles []Profile) {
	sort.Slice(profiles, func(i, j int) bool {
		a, b := profiles[i], profiles[j]
		if a.Group != b.Group {
			if a.Group == "" || b.Group == "" {
				return b.Group == ""
			}
			return a.Group < b.Group
		}
		return a.Name < b.Name
	})
}

// Profile names must be unique within their collection, so the same name may be
// reused across collections. skip is the index of the profile being edited, or
// -1 when adding a new one.
func nameTaken(profiles []Profile, p Profile, skip int) bool {
	for i, other := range profiles {
		if i != skip && other.Name == p.Name && other.Group == p.Group {
			return true
		}
	}
	return false
}

// Label of the section a profile is listed under.
func (p Profile) GroupLabel() string {
	if p.Group == "" {
		return "Ungrouped"
	}
	return p.Group
}

func hasGroups(profiles []Profile) bool {
	for _, p := range profiles {
		if p.Group != "" {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
