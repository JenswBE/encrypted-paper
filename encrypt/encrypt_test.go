package encrypt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateEncryptionRoundtrip(t *testing.T) {
	// Data
	password := "MY_VERY_SECURE_PASSWORD" // #nosec G101
	msg := "Should not be public"

	// Generate salt
	salt, err := GenerateSalt()
	require.NoError(t, err)

	// Create encryption AEAD
	aeadEnc, err := AEADFromPassword(password, salt)
	require.NoError(t, err)

	// Encrypt message
	encryptedMsg, err := Encrypt([]byte(msg), aeadEnc)
	require.NoError(t, err)

	// Create decryption AEAD => Ensures AEAD is ephemeral
	aeadDec, err := AEADFromPassword(password, salt)
	require.NoError(t, err)

	// Decrypt message
	decryptedMsg, err := Decrypt(encryptedMsg, aeadDec)
	require.NoError(t, err)

	// Validate result
	require.Equal(t, msg, string(decryptedMsg))
}
