package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

type Profile struct {
	Name       string `json:"name"`
	User       string `json:"user"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	KeyPath    string `json:"key_path,omitempty"`
	Password   string `json:"password,omitempty"`
	LastAccess string `json:"last_access,omitempty"`
}

const maxPasswordLen = 256

var validInputRE = regexp.MustCompile(`^[a-zA-Z0-9._@:-]+$`)

func isValidInput(s string) bool {
	return s != "" && validInputRE.MatchString(s)
}

// Passwords are handed to ssh through the environment, never through a shell,
// so any printable character is fine; control characters are not.
func isValidPassword(s string) bool {
	if len(s) > maxPasswordLen {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] == 0x7f {
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
	if p.KeyPath != "" && !isValidInput(p.KeyPath) {
		return &ProfileError{"invalid key path characters"}
	}
	if !isValidPassword(p.Password) {
		return &ProfileError{"invalid password (max 256 printable characters)"}
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

func profilesPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "profiles.json"
	}
	return filepath.Join(filepath.Dir(exe), "profiles.json")
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

func sortProfiles(profiles []Profile) {
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
