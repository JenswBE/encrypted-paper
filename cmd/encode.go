package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/JenswBE/encrypted-paper/compress"
	"github.com/JenswBE/encrypted-paper/encode"
	"github.com/JenswBE/encrypted-paper/encrypt"
)

var (
	encodeFlagTitle          string
	encodeFlagOutput         string
	encodeFlagMaxOutputFiles uint
	encodeCmd                = &cobra.Command{
		Use:          "encode [flags] input_file",
		Short:        "Compress, encrypt and convert data into QR codes",
		Args:         cobra.ExactArgs(1),
		RunE:         runEncode,
		SilenceUsage: true,
	}
)

func init() {
	encodeCmd.Flags().StringVarP(&encodeFlagTitle, "title", "t", "", "Title on each output page")
	encodeCmd.Flags().StringVarP(&encodeFlagOutput, "output", "o", "encrypted-paper.pdf", "Output file name")
	encodeCmd.Flags().UintVar(&encodeFlagMaxOutputFiles, "max-output-files", 10, "Maximum number of output files to generate. Set to 0 to disable limit.")
}

func runEncode(_ *cobra.Command, args []string) error {
	// Parse encode config
	config, err := parseEncodeConfig(encodeFlagTitle, args[0], encodeFlagOutput, encodeFlagMaxOutputFiles)
	if err != nil {
		return fmt.Errorf("failed to parse encode config: %w", err)
	}

	// Request password
	password, err := encrypt.GetPassword(true)
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	// Marshal data
	err = marshal(config, password)
	if err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}
	return nil
}

type EncodeConfig struct {
	InputPath      string
	MaxOutputFiles uint
	OutputFileName string
}

func parseEncodeConfig(title, inputFile, outputFileName string, maxOutputFiles uint) (EncodeConfig, error) {
	// Validate flags
	if title == "" {
		return EncodeConfig{}, errors.New("title is a mandatory parameter")
	}
	if inputFile == "" {
		return EncodeConfig{}, errors.New("input file is a mandatory parameter")
	}
	if outputFileName == "" {
		return EncodeConfig{}, errors.New("output file name cannot be empty")
	}
	if filepath.Ext(outputFileName) != ".pdf" {
		return EncodeConfig{}, errors.New("output file must have extension .pdf")
	}

	// Ensure input file is readable
	if _, err := os.Stat(inputFile); err != nil {
		return EncodeConfig{}, fmt.Errorf("unable to read input file: %w", err)
	}

	// Build and return flags
	return EncodeConfig{
		InputPath:      inputFile,
		MaxOutputFiles: maxOutputFiles,
		OutputFileName: outputFileName,
	}, nil
}

// MARSHAL
//  1. Compress with XZ
//  2. Encrypt using Argon2 and XChaCha20
//  3. Convert to QR code (include metadata)
//  4. Validate if output is decodeable and yields same as input
func marshal(config EncodeConfig, password string) error {
	// Read input file
	inputFileContents, err := os.ReadFile(config.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file contents: %w", err)
	}

	// Compress input file
	var compressedInput bytes.Buffer
	err = compress.Compress(bytes.NewReader(inputFileContents), &compressedInput)
	if err != nil {
		return fmt.Errorf("failed to compress input file: %w", err)
	}

	// Generate salt
	salt, err := encrypt.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate authenticated encryption cipher
	aead, err := encrypt.AEADFromPassword(password, salt)
	if err != nil {
		return fmt.Errorf("failed to create cipher from password and salt: %w", err)
	}

	// Encrypt input file
	encryptedInput, err := encrypt.Encrypt(compressedInput.Bytes(), aead)
	if err != nil {
		return fmt.Errorf("failed to encrypt input: %w", err)
	}

	// Encode into QR codes
	qrCodes, err := encode.GenerateQRCodes(salt, encryptedInput, int(config.MaxOutputFiles))
	if err != nil {
		return fmt.Errorf("failed to encode data into QR code: %w", err)
	}

	// Ensure QR codes are decodable
	decodedData, err := decodeQRCodes(qrCodes, password)
	if err != nil {
		return fmt.Errorf("failed to decode generated QR codes for validation: %w", err)
	}

	// Compare input data and decoded QR codes
	if !bytes.Equal(inputFileContents, decodedData) {
		return errors.New("input data and decoded QR data are different")
	}

	// Generate PDF
	err = encode.GeneratePDF(config.OutputFileName, encodeFlagTitle, qrCodes)
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}
	return nil
}
