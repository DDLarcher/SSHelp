package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen  = 16
	nonceLen = 12
	keyIter  = 100000 // legacy PBKDF2 files only
	fileMode = 0o600
)

// v2 files carry this prefix. A file without it was written by an older version
// with PBKDF2 key derivation and is re-encrypted the first time it is opened.
var magicV2 = []byte("SSHELPv2")

// Argon2id: 64 MiB and 3 passes, which costs an attacker memory as well as
// time. The file holds live SSH passwords, so PBKDF2 iterations alone are no
// longer an adequate barrier.
const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
)

func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, 32)
}

func deriveKeyLegacy(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, keyIter, 32, sha256.New)
}

func encryptProfiles(profiles []Profile, password string) error {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	block, err := aes.NewCipher(deriveKey(password, salt))
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	data, err := json.Marshal(profiles)
	if err != nil {
		return err
	}

	out := append([]byte{}, magicV2...)
	out = append(out, salt...)
	out = append(out, nonce...)
	out = append(out, aesgcm.Seal(nil, nonce, data, nil)...)

	return writeFilePrivate(profilesWritePath(), out)
}

// Writes through a temporary file so that a crash mid-write cannot leave an
// unreadable store behind, and keeps the result readable by its owner only.
func writeFilePrivate(path string, data []byte) error {
	if dir := filepath.Dir(path); dir != "" {
		os.MkdirAll(dir, 0o700)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, fileMode); err != nil {
		return err
	}
	// WriteFile only applies the mode when creating the file.
	if err := os.Chmod(tmp, fileMode); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Chmod(path, fileMode)
}

func decryptProfiles(password string) ([]Profile, bool, error) {
	raw, err := os.ReadFile(profilesReadPath())
	if err != nil {
		return nil, false, err
	}

	legacy := !bytes.HasPrefix(raw, magicV2)
	if !legacy {
		raw = raw[len(magicV2):]
	}
	if len(raw) < saltLen+nonceLen {
		return nil, legacy, &ProfileError{"profile store is truncated"}
	}

	salt, nonce, encrypted := raw[:saltLen], raw[saltLen:saltLen+nonceLen], raw[saltLen+nonceLen:]

	key := deriveKey(password, salt)
	if legacy {
		key = deriveKeyLegacy(password, salt)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, legacy, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, legacy, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, legacy, err
	}

	var profiles []Profile
	if err := json.Unmarshal(plaintext, &profiles); err != nil {
		return nil, legacy, err
	}
	if profiles == nil {
		profiles = []Profile{}
	}
	return profiles, legacy, nil
}

func profilesFileExists() bool {
	_, err := os.Stat(profilesReadPath())
	return err == nil
}

func LoadProfiles(password string) ([]Profile, error) {
	profiles, legacy, err := decryptProfiles(password)
	if err != nil {
		return nil, err
	}

	valid, skipped := filterValid(profiles)

	var warnings []string
	if skipped > 0 {
		warnings = append(warnings, "Loaded "+itoa(len(valid))+" profiles, "+itoa(skipped)+" skipped (invalid)")
	}
	if legacy {
		if err := encryptProfiles(valid, password); err != nil {
			warnings = append(warnings, "Could not upgrade the profile store: "+err.Error())
		} else {
			warnings = append(warnings, "Profile store upgraded to Argon2id at "+profilesWritePath())
			if old := legacyProfilesPath(); old != "" && old != profilesWritePath() {
				os.Chmod(old, fileMode)
				warnings = append(warnings, "the old copy at "+old+" can be deleted")
			}
		}
	}
	if len(warnings) > 0 {
		return valid, &ProfileError{joinWarnings(warnings)}
	}
	return valid, nil
}

func SaveProfiles(profiles []Profile, password string) error {
	return encryptProfiles(profiles, password)
}
