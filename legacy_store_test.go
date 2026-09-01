package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"
	"testing"
)

// Builds a store in the pre-Argon2id format: salt || nonce || ciphertext with
// no magic prefix and a PBKDF2 key.
func writeLegacyStore(t *testing.T, profiles []Profile, password string) []byte {
	t.Helper()

	salt := make([]byte, saltLen)
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		t.Fatal(err)
	}

	block, err := aes.NewCipher(deriveKeyLegacy(password, salt))
	if err != nil {
		t.Fatal(err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(profiles)
	if err != nil {
		t.Fatal(err)
	}

	out := append([]byte{}, salt...)
	out = append(out, nonce...)
	return append(out, aesgcm.Seal(nil, nonce, data, nil)...)
}
