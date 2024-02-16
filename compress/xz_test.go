package compress

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCompressRoundtrip(t *testing.T) {
	// Data
	input := strings.Repeat("Data to be compressed\n", 100)

	// Compress
	inputReader := strings.NewReader(input)
	var compressedOutput bytes.Buffer
	err := Compress(inputReader, &compressedOutput)
	require.NoError(t, err)
	require.Less(t, compressedOutput.Len(), len([]byte(input)), "Compressed message should be shorter than original message")

	// Decompress
	var decompressedOutput bytes.Buffer
	err = Decompress(&compressedOutput, &decompressedOutput)
	require.NoError(t, err)

	// Validate result
	require.Equal(t, input, decompressedOutput.String())
}
