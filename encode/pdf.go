package encode

import (
	"fmt"

	"github.com/signintech/gopdf"
)

const FontSize = 12

func GeneratePDF(outputPath string, title string, qrCodes [][]byte) error {
	// Init PDF
	pdf := gopdf.GoPdf{}
	pageSize := *gopdf.PageSizeA4
	pdf.Start(gopdf.Config{PageSize: pageSize})

	// Set font
	fontName := "dejavu"
	fontPath := "assets/DejaVuSans.ttf"
	err := pdf.AddTTFFont(fontName, fontPath)
	if err != nil {
		return fmt.Errorf("failed to add TTF font %s as %s: %w", fontPath, fontName, err)
	}
	err = pdf.SetFont(fontName, "", FontSize)
	if err != nil {
		return fmt.Errorf("failed to set font to %s: %w", fontName, err)
	}

	// Set header
	pdf.AddHeader(func() {
		pdf.SetY(30)
		pdf.CellWithOption(&gopdf.Rect{W: pageSize.W, H: FontSize + 2}, title, gopdf.CellOption{Align: gopdf.Center})
	})

	// Set footer
	pdf.AddFooter(func() {
		pdf.SetFontSize(8)
		pdf.SetY(pageSize.H - 20)
		pdf.Cell(nil, "Generated with https://github.com/JenswBE/encrypted-paper")

		pdf.SetX(pageSize.W - 75)
		pdf.Cell(nil, fmt.Sprintf("Page %d of %d", pdf.GetNumberOfPages(), len(qrCodes)))
	})

	// Generate pages
	for i, qrCode := range qrCodes {
		pdf.AddPage()
		pdf.SetY(50)

		holder, err := gopdf.ImageHolderByBytes(qrCode)
		if err != nil {
			return fmt.Errorf("failed to convert QR code image %d to holder: %w", i+1, err)
		}

		imageMargin := float64(25)
		imageSize := pageSize.W - imageMargin*2
		imageYPos := pageSize.H/2 - imageSize/2
		err = pdf.ImageByHolder(holder, imageMargin, imageYPos, &gopdf.Rect{W: imageSize, H: imageSize})
		if err != nil {
			return fmt.Errorf("failed to add QR code %d image holder to PDF file: %w", i+1, err)
		}
	}

	// Write PDF file
	err = pdf.WritePdf(outputPath)
	if err != nil {
		return fmt.Errorf("failed to write PDF file to path %s: %w", outputPath, err)
	}
	return nil
}
