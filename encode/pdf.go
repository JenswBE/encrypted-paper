package encode

import (
	"fmt"
	"time"

	"github.com/signintech/gopdf"

	"github.com/JenswBE/encrypted-paper/assets"
)

const FontSize = 12

func GeneratePDF(outputPath string, title string, qrCodes [][]byte) (err error) {
	// Init PDF
	pdf := gopdf.GoPdf{}
	pageSize := *gopdf.PageSizeA4
	pdf.Start(gopdf.Config{PageSize: pageSize})

	// Set font
	fontName := "dejavu"
	err = pdf.AddTTFFontData(fontName, assets.DejaVuSansTTF)
	if err != nil {
		return fmt.Errorf("failed to add TTF font %s: %w", fontName, err)
	}
	err = pdf.SetFont(fontName, "", FontSize)
	if err != nil {
		return fmt.Errorf("failed to set font to %s: %w", fontName, err)
	}

	// Set header
	pdf.AddHeader(func() {
		pdf.SetY(30)
		err = pdf.CellWithOption(&gopdf.Rect{W: pageSize.W, H: FontSize + 2}, title, gopdf.CellOption{Align: gopdf.Center})
		if err != nil {
			err = fmt.Errorf("failed to add title in header: %w", err)
		}
	})
	if err != nil {
		return err
	}

	// Set footer
	pdf.AddFooter(func() {
		err = pdf.SetFontSize(8)
		if err != nil {
			err = fmt.Errorf("failed to set font size in footer: %w", err)
			return
		}
		pdf.SetX(20)
		pdf.SetY(pageSize.H - 20)
		err = pdf.Cell(nil, "Generated with https://github.com/JenswBE/encrypted-paper on "+time.Now().Format("02 Jan 2006 15:04 -0700"))
		if err != nil {
			err = fmt.Errorf("failed to set project URL in footer: %w", err)
			return
		}
		pdf.SetX(pageSize.W - 75)
		pdf.SetY(pageSize.H - 20)
		err = pdf.Cell(nil, fmt.Sprintf("Page %d of %d", pdf.GetNumberOfPages(), len(qrCodes)))
		if err != nil {
			err = fmt.Errorf("failed to set page number in footer: %w", err)
			return
		}
	})
	if err != nil {
		return err
	}

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
