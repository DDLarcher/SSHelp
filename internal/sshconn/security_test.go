package sshconn

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/DDLarcher/SSHelp/internal/profile"
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
	p := profile.Profile{Name: "web", User: "root", Host: "web.example.com", Port: 22, Password: "s3cret"}
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
	p := profile.Profile{Name: "web", User: "root", Host: "web.example.com", Port: 22}
	args := strings.Join(sshArgs(p, ""), " ")
	if strings.Contains(args, "StrictHostKeyChecking") || strings.Contains(args, "UserKnownHostsFile") {
		t.Errorf("a profile without a password must keep plain ssh behaviour: %s", args)
	}
}

func TestPasswordIsNotPassedThroughTheEnvironment(t *testing.T) {
	isolate(t)
	p := profile.Profile{Name: "web", User: "root", Host: "web.example.com", Port: 22, Password: "s3cret-value"}

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
	} else if info.Mode().Perm() != profile.FileMode {
		t.Errorf("credential file mode = %o, want %o", info.Mode().Perm(), profile.FileMode)
	}

	if code := Askpass(cred, "root@web.example.com's password: "); code != 0 {
		t.Errorf("first password prompt refused (exit %d)", code)
	}
	if _, err := os.Stat(cred); !os.IsNotExist(err) {
		t.Error("credential file survived the read")
	}
	if code := Askpass(cred, "Password: "); code == 0 {
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
	p := profile.Profile{Name: "web", User: "root", Host: "nonexistent.invalid", Port: 22, Password: "s3cret"}
	if _, err := PinHostKeys(p); err == nil {
		t.Error("a password was accepted for a host nobody has ever verified")
	}

	p.Password = ""
	got, err := PinHostKeys(p)
	if err != nil || got.HostKeys != nil {
		t.Errorf("clearing the password should drop the pin: %v %v", got.HostKeys, err)
	}
}
