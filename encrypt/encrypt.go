package encrypt

import (
	"crypto/cipher"
	cryptorand "crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

func GenerateSalt() ([]byte, error) {
	var salt = make([]byte, 16)
	_, err := cryptorand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

func AEADFromPassword(password string, salt []byte) (cipher.AEAD, error) {
	key := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create new XChaCha20 Poly1305 AEAD: %w", err)
	}
	return aead, nil
}

func Encrypt(msg []byte, aead cipher.AEAD) ([]byte, error) {
	// Based on https://pkg.go.dev/golang.org/x/crypto/chacha20poly1305#example-NewX
	// Select a random nonce, and leave capacity for the ciphertext.
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(msg)+aead.Overhead())
	if _, err := cryptorand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to select a random nonce: %w", err)
	}

	// Encrypt the message and append the ciphertext to the nonce.
	encryptedMsg := aead.Seal(nonce, nonce, msg, nil)
	return encryptedMsg, nil
}

func Decrypt(encryptedMsg []byte, aead cipher.AEAD) ([]byte, error) {
	// Based on https://pkg.go.dev/golang.org/x/crypto/chacha20poly1305#example-NewX
	// Validate length of encrypted message
	if len(encryptedMsg) < aead.NonceSize() {
		return nil, fmt.Errorf("cipher text (len %d) must be longer than nonce size (len %d)", len(encryptedMsg), aead.NonceSize())
	}

	// Split nonce and ciphertext
	nonce, ciphertext := encryptedMsg[:aead.NonceSize()], encryptedMsg[aead.NonceSize():]

	// Decrypt the message and check it wasn't tampered with
	msg, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt and authenticate cipher text: %w", err)
	}
	return msg, nil
}
