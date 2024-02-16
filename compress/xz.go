package compress

import (
	"io"

	"github.com/JenswBE/encrypted-paper/utils"
)

func Compress(input io.Reader, output io.Writer) error {
	return utils.RunCommand("compress input data", input, output, "xz", "--compress", "-9", "--extreme", "--stdout")
}

func Decompress(input io.Reader, output io.Writer) error {
	return utils.RunCommand("decompress input data", input, output, "xz", "--decompress", "--stdout")
}
