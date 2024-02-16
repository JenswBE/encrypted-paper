package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/JenswBE/encrypted-paper/compress"
	"github.com/JenswBE/encrypted-paper/encode"
	"github.com/JenswBE/encrypted-paper/encrypt"
)

var (
	decodeFlagOutput string
	decodeFlagForce  bool
	decodeCmd        = &cobra.Command{
		Use:          "decode [flags] input_file ...",
		Short:        "Parse QR code, decrypt and decompress data",
		RunE:         runDecode,
		SilenceUsage: true,
	}
)

func init() {
	decodeCmd.Flags().StringVarP(&decodeFlagOutput, "output", "o", "", "Output file name")
	decodeCmd.Flags().BoolVar(&decodeFlagForce, "force", false, "Force overwrite output file if exists")
}

// DECODE
// 1. Read QR codes
// 2. Decrypt
// 3. Decompress
func runDecode(cmd *cobra.Command, args []string) error {
	// Validate flags
	if decodeFlagOutput == "" {
		return errors.New("output is a mandatory parameter")
	}
	if len(args) == 0 {
		return errors.New("at least 1 input file should be provided")
	}

	// Check output file already exists
	_, err := os.Stat(decodeFlagOutput)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check if output file exists: %w", err)
	}
	if err == nil && !decodeFlagForce {
		return errors.New("output file already exists: either set flag --force or use another output file")
	}

	// Read input files
	inputFilesContents := make([][]byte, len(args))
	for i, inputFile := range args {
		inputFilesContents[i], err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", inputFile, err)
		}
	}

	// Request password
	password, err := encrypt.GetPassword(false)
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	// Decode QR codes
	output, err := decodeQRCodes(inputFilesContents, password)
	if err != nil {
		return fmt.Errorf("failed to decode QR codes: %w", err)
	}

	// Write output file
	err = os.WriteFile(decodeFlagOutput, output, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	return nil
}

func decodeQRCodes(qrCodes [][]byte, password string) ([]byte, error) {
	// Scan and combine QR codes
	encryptedData, salt, err := encode.ScanQRCodes(qrCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to scan and combine QR codes: %w", err)
	}

	// Generate authenticated encryption cipher
	aead, err := encrypt.AEADFromPassword(password, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher from password and salt: %w", err)
	}

	// Decode data
	compressedData, err := encrypt.Decrypt(encryptedData, aead)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Decompress input file
	var data bytes.Buffer
	err = compress.Decompress(bytes.NewReader(compressedData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}
	return data.Bytes(), nil
}
