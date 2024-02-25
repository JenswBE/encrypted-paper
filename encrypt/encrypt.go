package encrypt

import (
	"crypto/cipher"
	cryptorand "crypto/rand"
	"fmt"
	"log/slog"
	"strings"
	"syscall"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/term"
)

// Recommended salt size based on https://en.wikipedia.org/wiki/Argon2#Algorithm
const SaltSizeBytes = 16

// Minimum length for password
const MinPasswordLength = 8

func GetPassword(withConfirm bool) (string, error) {
	for {
		password, err := getPassword("Enter your password")
		if err != nil {
			return "", fmt.Errorf("failed to get password: %w", err)
		}
		if len(password) < MinPasswordLength {
			fmt.Printf("\nPassword must at least have a length of %d. Please try again.\n\n", MinPasswordLength)
			continue
		}
		if withConfirm {
			repeatedPassword, err := getPassword("Repeat your password")
			if err != nil {
				return "", fmt.Errorf("failed to get repeated password: %w", err)
			}
			if password != repeatedPassword {
				fmt.Printf("\nRepeated password is different from original password. Please try again.\n\n")
				continue
			}
		}
		return password, nil
	}
}

func getPassword(prompt string) (string, error) {
	// Based on https://stackoverflow.com/a/32768479
	fmt.Print(prompt + ": ")
	password, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()
	return strings.TrimSpace(string(password)), nil
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSizeBytes)
	_, err := cryptorand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

func AEADFromPassword(password string, salt []byte) (cipher.AEAD, error) {
	if len(password) < MinPasswordLength {
		return nil, fmt.Errorf("password shorter than minimum length of %d", MinPasswordLength)
	}

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
		slog.Error("Decryption failed. Please check you password and retry.")
		return nil, fmt.Errorf("failed to decrypt and authenticate cipher text: %w", err)
	}
	return msg, nil
}
