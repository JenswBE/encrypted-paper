package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "encrypted-paper",
	Short: "Compress, encrypt and convert data into QR codes.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(encodeCmd, decodeCmd)
}
