package profile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Isolates the profile store and the state directory from the user's real ones.
func isolate(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("relies on the XDG directory variables")
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
}

func TestStoreIsPrivateAndUpgradesFromTheLegacyFormat(t *testing.T) {
	isolate(t)
	profiles := []Profile{{Name: "web", User: "root", Host: "web.example.com", Port: 22, Password: "s3cret"}}

	if err := Save(profiles, "master"); err != nil {
		t.Fatal(err)
	}
	path := profilesReadPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != FileMode {
		t.Errorf("store mode = %o, want %o", info.Mode().Perm(), FileMode)
	}
	if dir, _ := filepath.Split(path); strings.Contains(dir, "go-build") {
		t.Errorf("store was written next to the executable: %s", path)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), string(magicV2)) {
		t.Error("store was not written in the Argon2id format")
	}
	if strings.Contains(string(raw), "s3cret") {
		t.Error("the store is not encrypted")
	}

	got, err := Load("master")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Password != "s3cret" {
		t.Fatalf("round trip lost data: %+v", got)
	}
	if _, err := Load("wrong"); err == nil {
		t.Error("the wrong master password decrypted the store")
	}
}

func TestLegacyPBKDF2StoreIsReadAndRewritten(t *testing.T) {
	isolate(t)
	legacy := writeLegacyStore(t, []Profile{{Name: "old", User: "root", Host: "a.example.com", Port: 22}}, "master")

	if err := os.MkdirAll(filepath.Dir(profilesWritePath()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profilesWritePath(), legacy, 0o644); err != nil {
		t.Fatal(err)
	}
	os.Chmod(profilesWritePath(), 0o644)

	got, err := Load("master")
	if err == nil {
		t.Error("the format upgrade should be reported to the user")
	} else if _, ok := err.(*Error); !ok {
		t.Fatalf("upgrade failed: %v", err)
	}
	if len(got) != 1 || got[0].Name != "old" {
		t.Fatalf("legacy profiles lost: %+v", got)
	}

	raw, err := os.ReadFile(profilesReadPath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), string(magicV2)) {
		t.Error("the legacy store was not rewritten in the new format")
	}
	if info, _ := os.Stat(profilesReadPath()); info.Mode().Perm() != FileMode {
		t.Errorf("rewritten store mode = %o, want %o", info.Mode().Perm(), FileMode)
	}
}
