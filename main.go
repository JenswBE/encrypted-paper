package main

import (
	"bytes"
	"flag"
	"log/slog"
	"os"

	"github.com/JenswBE/encrypted-paper/compress"
)

func main() {
	// 0. Validate deps available: xz, qrencode, zbarimg

	// ENCODE
	// 1. Compress with XZ
	// 2. Encrypt using Argon2 and XChaCha20
	// 3. Convert to QR code (include metadata)
	// 4. Validate if output is decodeable and yields same as input

	// Parse and validate flags
	inputPath := flag.String("i", "", "Input file")
	flag.Parse()
	if *inputPath == "" {
		slog.Error("Input file path is a mandatory parameter")
		os.Exit(1)
	}

	// Open input file
	inputFile, err := os.Open(*inputPath)
	if err != nil {
		slog.Error("Failed to open input file", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := inputFile.Close(); closeErr != nil {
			slog.Error("Failed to close input file", "error", err)
		}
	}()

	// Compress input file
	var compressedInput bytes.Buffer
	err = compress.Compress(inputFile, &compressedInput)
	if err != nil {
		slog.Error("Failed to compress input file", "error", err)
		os.Exit(1)
	}

	// DECODE
	// 1. Read QR code
	// 2. Decrypt
	// 3. Decompress
}
