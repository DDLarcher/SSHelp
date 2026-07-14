package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Profile struct {
	Name       string `json:"name"`
	User       string `json:"user"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	KeyPath    string `json:"key_path,omitempty"`
	LastAccess string `json:"last_access,omitempty"`
}

func profilesPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "profiles.json"
	}
	return filepath.Join(filepath.Dir(exe), "profiles.json")
}

func LoadProfiles() ([]Profile, error) {
	data, err := os.ReadFile(profilesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []Profile{}, nil
		}
		return nil, err
	}
	var profiles []Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}
	if profiles == nil {
		profiles = []Profile{}
	}
	return profiles, nil
}

func SaveProfiles(profiles []Profile) error {
	if profiles == nil {
		profiles = []Profile{}
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(profilesPath(), data, 0644)
}

func (p Profile) HostInfo() string {
	info := fmt.Sprintf("%s@%s:%d", p.User, p.Host, p.Port)
	if p.KeyPath != "" {
		info += " (key)"
	}
	return info
}

func (p Profile) UpdatedNow() Profile {
	p.LastAccess = time.Now().Format("2006-01-02 15:04")
	return p
}
