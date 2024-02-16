package encode

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/fxamacker/cbor/v2"

	"github.com/JenswBE/encrypted-paper/encrypt"
	"github.com/JenswBE/encrypted-paper/utils"
)

// See https://en.wikipedia.org/wiki/QR_code#Information_capacity
const (
	MaxBytesInQRCode = 2953
	MaxPageCount     = math.MaxUint8
)

type QRHeader struct {
	Salt      []byte `json:"salt"`
	PageCount uint8  `json:"page_count"`
}

type QRData struct {
	Header     *QRHeader `json:"header,omitempty"`
	PageNumber uint8     `json:"page_number"`
	Data       []byte    `json:"data"`
}

func getQRDataOverhead(withHeader bool) int {
	qrData := QRData{
		Data:       []byte{1},
		PageNumber: MaxPageCount,
	}
	if withHeader {
		qrData.Header = &QRHeader{
			Salt:      make([]byte, encrypt.SaltSizeBytes),
			PageCount: MaxPageCount,
		}
	}
	output, err := cbor.Marshal(qrData)
	if err != nil {
		panic(fmt.Sprintf("Failed to calculate QR data overhead: %v", err))
	}
	return len(output)
}

func GenerateQRCodes(salt []byte, data []byte, maxOutputPages int) ([][]byte, error) {
	// Calculate overhead
	maxDataSizeWithHeader := MaxBytesInQRCode - getQRDataOverhead(true)
	maxDataSizeWithoutHeader := MaxBytesInQRCode - getQRDataOverhead(false)
	pageCount := calcPageCount(maxDataSizeWithHeader, maxDataSizeWithoutHeader, len(data))
	if pageCount > math.MaxUint8 {
		return nil, fmt.Errorf("page count is %d, but maximum supported page count in header is %d", pageCount, MaxPageCount)
	}

	// Validate max output pages
	if pageCount > maxOutputPages {
		return nil, fmt.Errorf("%d expected output pages is more than configured maximum of %d allowed output pages", pageCount, maxOutputPages)
	}

	// Generate QR codes
	var cursor int
	var qrData QRData
	var err error
	output := make([][]byte, pageCount)
	for i := range pageCount {
		pageNumber := i + 1 // 1 for zero indexed
		if pageNumber == 1 {
			// Stage first page
			qrData = QRData{
				Header: &QRHeader{
					Salt:      salt,
					PageCount: uint8(pageCount),
				},
				PageNumber: 1,
				Data:       data,
			}
			if pageCount > 1 {
				// Data is too big for a single page
				qrData.Data = data[:maxDataSizeWithHeader]
				cursor = maxDataSizeWithHeader
			}
		} else {
			// Stage remaining pages
			readUntil := cursor + maxDataSizeWithoutHeader
			if readUntil > len(data) {
				readUntil = len(data)
			}
			qrData = QRData{
				PageNumber: uint8(pageNumber),
				Data:       data[cursor:readUntil],
			}
			cursor = readUntil
		}

		// Marchal to CBOR and generate QR code
		output[i], err = marshalAndCreateQR(qrData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate page %d: %w", pageNumber, err)
		}
	}
	return output, nil
}

func marshalAndCreateQR(qrData QRData) ([]byte, error) {
	// Marshal into CBOR
	var cborData bytes.Buffer
	err := cbor.NewEncoder(&cborData).Encode(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data as CBOR: %w", err)
	}

	// Encode as QR code
	var qrCode bytes.Buffer
	err = utils.RunCommand("generate QR code", &cborData, &qrCode, "qrencode", "--8bit", "--level=L", "--output=-", "--dpi=300", "--size=10")
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	return qrCode.Bytes(), nil
}

func calcPageCount(maxDataSizeWithHeader, maxDataSizeWithoutHeader, totalDataSize int) int {
	if totalDataSize <= maxDataSizeWithHeader {
		return 1
	}
	remainingSize := totalDataSize - maxDataSizeWithHeader
	return remainingSize/maxDataSizeWithoutHeader + 2 // 1 for page with header and 1 for int decimal cutoff
}

func ScanQRCodes(qrCodes [][]byte) (data, salt []byte, err error) {
	// Scan and unmarshal QR codes
	qrDatas := make([]QRData, len(qrCodes))
	for i, qrCode := range qrCodes {
		// Scan QR code
		var cborData bytes.Buffer
		err = utils.RunCommand("scan QR code", bytes.NewReader(qrCode), &cborData, "zbarimg", "--raw", "--oneshot", "--set=binary", "-")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan QR code: %w", err)
		}

		// Unmarshal from CBOR
		err = cbor.NewDecoder(&cborData).Decode(&qrDatas[i])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode data as CBOR: %w", err)
		}
	}

	// Sort and combine codes
	slices.SortFunc(qrDatas, func(a, b QRData) int { return cmp.Compare(a.PageNumber, b.PageNumber) })
	var buf bytes.Buffer
	buf.Grow(MaxBytesInQRCode * len(qrCodes)) // Ignore overhead of metadata to keep code KISS
	for i, qrData := range qrDatas {
		if qrData.PageNumber != uint8(i+1) {
			return nil, nil, fmt.Errorf("page %d is missing", i+1)
		}
		if i == 0 {
			// First page contains salt and page count
			if qrData.Header == nil {
				return nil, nil, errors.New("header with metadata not found in first page")
			}
			if len(qrData.Header.Salt) != encrypt.SaltSizeBytes {
				return nil, nil, fmt.Errorf("salt in header is %d bytes, but salt must be %d bytes", len(qrData.Header.Salt), encrypt.SaltSizeBytes)
			}
			salt = qrData.Header.Salt
			if int(qrData.Header.PageCount) != len(qrCodes) {
				return nil, nil, fmt.Errorf("%d qr codes received, but accordingly to header, there must be %d qr codes", len(qrCodes), qrData.Header.PageCount)
			}
		}
		buf.Write(qrData.Data)
	}
	return buf.Bytes(), salt, nil
}
