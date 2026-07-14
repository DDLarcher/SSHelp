package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen  = 16
	nonceLen = 12
	keyIter  = 100000
)

func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, keyIter, 32, sha256.New)
}

func encryptProfiles(profiles []Profile, password string) error {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
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

	encrypted := aesgcm.Seal(nil, nonce, data, nil)

	f, err := os.Create(profilesPath())
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(salt)
	f.Write(nonce)
	f.Write(encrypted)
	return nil
}

func decryptProfiles(password string) ([]Profile, error) {
	f, err := os.Open(profilesPath())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(f, salt); err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(f, nonce); err != nil {
		return nil, err
	}

	encrypted, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}

	var profiles []Profile
	if err := json.Unmarshal(plaintext, &profiles); err != nil {
		return nil, err
	}
	if profiles == nil {
		profiles = []Profile{}
	}
	return profiles, nil
}

func profilesFileExists() bool {
	_, err := os.Stat(profilesPath())
	return err == nil
}

func LoadProfiles(password string) ([]Profile, error) {
	profiles, err := decryptProfiles(password)
	if err != nil {
		return nil, err
	}
	valid, skipped := filterValid(profiles)
	if skipped > 0 {
		return valid, &ProfileError{"Loaded " + itoa(len(valid)) + " profiles, " + itoa(skipped) + " skipped (invalid)"}
	}
	return valid, nil
}

func SaveProfiles(profiles []Profile, password string) error {
	return encryptProfiles(profiles, password)
}
