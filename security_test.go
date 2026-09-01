package main

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

func TestPasswordProfileNeverAutoAcceptsAHostKey(t *testing.T) {
	p := Profile{Name: "web", User: "root", Host: "web.example.com", Port: 22, Password: "s3cret"}
	args := strings.Join(sshArgs(p, "/tmp/kh"), " ")

	if strings.Contains(args, "accept-new") {
		t.Error("an unknown host key must not be accepted for a profile that auto-sends a password")
	}
	if !strings.Contains(args, "StrictHostKeyChecking=yes") {
		t.Errorf("missing strict host key checking: %s", args)
	}
	if !strings.Contains(args, `UserKnownHostsFile="/tmp/kh"`) {
		t.Errorf("pinned known_hosts not quoted or not passed: %s", args)
	}
	// Forcing password auth would put the password on the wire even when an
	// agent key would have been accepted.
	if strings.Contains(args, "PreferredAuthentications") {
		t.Errorf("must not disable public key authentication: %s", args)
	}
}

func TestNoSecurityOptionsWithoutAStoredPassword(t *testing.T) {
	p := Profile{Name: "web", User: "root", Host: "web.example.com", Port: 22}
	args := strings.Join(sshArgs(p, ""), " ")
	if strings.Contains(args, "StrictHostKeyChecking") || strings.Contains(args, "UserKnownHostsFile") {
		t.Errorf("a profile without a password must keep plain ssh behaviour: %s", args)
	}
}

func TestPasswordIsNotPassedThroughTheEnvironment(t *testing.T) {
	isolate(t)
	p := Profile{Name: "web", User: "root", Host: "web.example.com", Port: 22, Password: "s3cret-value"}

	setup, err := prepareConnection(p)
	if err != nil {
		t.Fatal(err)
	}
	defer setup.cleanup()

	for _, e := range setup.env {
		if strings.Contains(e, "s3cret-value") {
			t.Errorf("password found in the ssh environment: %s", e)
		}
	}
	for _, a := range setup.args {
		if strings.Contains(a, "s3cret-value") {
			t.Errorf("password found on the command line: %s", a)
		}
	}
}

func TestAskpassAnswersOnlyTheFirstPasswordPrompt(t *testing.T) {
	isolate(t)
	cred, err := writeCredential("s3cret")
	if err != nil {
		t.Fatal(err)
	}

	if info, err := os.Stat(cred); err != nil {
		t.Fatal(err)
	} else if info.Mode().Perm() != fileMode {
		t.Errorf("credential file mode = %o, want %o", info.Mode().Perm(), fileMode)
	}

	if code := askpass(cred, "root@web.example.com's password: "); code != 0 {
		t.Errorf("first password prompt refused (exit %d)", code)
	}
	if _, err := os.Stat(cred); !os.IsNotExist(err) {
		t.Error("credential file survived the read")
	}
	if code := askpass(cred, "Password: "); code == 0 {
		t.Error("a second prompt was answered")
	}
}

func TestAskpassRejectsPromptsTheServerControls(t *testing.T) {
	for _, prompt := range []string{
		"Verification code: ",
		"Password (2FA): ",
		"OTP password: ",
		"Are you sure you want to continue connecting (yes/no/[fingerprint])? ",
		"Enter your password on your phone, then press enter",
		"",
	} {
		if isPasswordPrompt(prompt) {
			t.Errorf("prompt %q must not be answered", prompt)
		}
	}
	for _, prompt := range []string{
		"root@web.example.com's password: ",
		"Password: ",
		"password: ",
		"Enter passphrase for key '/home/me/.ssh/id_ed25519': ",
	} {
		if !isPasswordPrompt(prompt) {
			t.Errorf("prompt %q should be answered", prompt)
		}
	}
}

func TestPasswordRequiresAVerifiedHostKey(t *testing.T) {
	isolate(t)
	p := Profile{Name: "web", User: "root", Host: "nonexistent.invalid", Port: 22, Password: "s3cret"}
	if _, err := pinHostKeys(p); err == nil {
		t.Error("a password was accepted for a host nobody has ever verified")
	}

	p.Password = ""
	got, err := pinHostKeys(p)
	if err != nil || got.HostKeys != nil {
		t.Errorf("clearing the password should drop the pin: %v %v", got.HostKeys, err)
	}
}

func TestStoreIsPrivateAndUpgradesFromTheLegacyFormat(t *testing.T) {
	isolate(t)
	profiles := []Profile{{Name: "web", User: "root", Host: "web.example.com", Port: 22, Password: "s3cret"}}

	if err := SaveProfiles(profiles, "master"); err != nil {
		t.Fatal(err)
	}
	path := profilesReadPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != fileMode {
		t.Errorf("store mode = %o, want %o", info.Mode().Perm(), fileMode)
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

	got, err := LoadProfiles("master")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Password != "s3cret" {
		t.Fatalf("round trip lost data: %+v", got)
	}
	if _, err := LoadProfiles("wrong"); err == nil {
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

	got, err := LoadProfiles("master")
	if err == nil {
		t.Error("the format upgrade should be reported to the user")
	} else if _, ok := err.(*ProfileError); !ok {
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
	if info, _ := os.Stat(profilesReadPath()); info.Mode().Perm() != fileMode {
		t.Errorf("rewritten store mode = %o, want %o", info.Mode().Perm(), fileMode)
	}
}
