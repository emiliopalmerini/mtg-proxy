package pdf

import (
	"fmt"

	"github.com/epalmerini/mtg-proxy/internal/card"
	"github.com/go-pdf/fpdf"
)

const (
	// A4 dimensions in mm
	pageW = 210.0
	pageH = 297.0

	// Standard MTG card size in mm
	cardW = 63.0
	cardH = 88.0

	cols = 3
	rows = 3

	cardsPerPage = cols * rows

	// Margins to center the 3x3 grid on A4
	marginX = (pageW - cols*cardW) / 2
	marginY = (pageH - rows*cardH) / 2

	// Internal card padding
	padding = 2.0

	fontFamily = "Helvetica"
)

// Renderer generates a printable PDF of proxy cards.
type Renderer struct {
	compress bool
}

type Option func(*Renderer)

// WithCompression enables or disables PDF stream compression.
func WithCompression(on bool) Option {
	return func(r *Renderer) { r.compress = on }
}

func NewRenderer(opts ...Option) *Renderer {
	r := &Renderer{compress: true}
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *Renderer) Render(cards []card.DeckCard, outputPath string) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(r.compress)
	pdf.SetAutoPageBreak(false, 0)

	expanded := expandDeck(cards)

	if len(expanded) == 0 {
		pdf.AddPage()
	}

	for i, c := range expanded {
		if i%cardsPerPage == 0 {
			pdf.AddPage()
		}

		pos := i % cardsPerPage
		col := pos % cols
		row := pos / cols

		x := marginX + float64(col)*cardW
		y := marginY + float64(row)*cardH

		renderCard(pdf, c, x, y)
	}

	return pdf.OutputFileAndClose(outputPath)
}

func expandDeck(cards []card.DeckCard) []card.Card {
	var expanded []card.Card
	for _, dc := range cards {
		for i := 0; i < int(dc.Quantity); i++ {
			expanded = append(expanded, dc.Card)
		}
	}
	return expanded
}

func renderCard(pdf *fpdf.Fpdf, c card.Card, x, y float64) {
	// Card border
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.3)
	pdf.Rect(x, y, cardW, cardH, "D")

	innerX := x + padding
	innerW := cardW - 2*padding
	cursorY := y + padding

	// Header: Name (left) + Mana cost (right)
	pdf.SetFont(fontFamily, "B", 7)
	pdf.SetXY(innerX, cursorY)

	manaCost := c.ManaCost.String()
	nameW := innerW
	if manaCost != "" {
		costW := pdf.GetStringWidth(manaCost) + 1
		nameW = innerW - costW
		pdf.CellFormat(nameW, 4, string(c.Name), "", 0, "L", false, 0, "")
		pdf.CellFormat(costW, 4, manaCost, "", 0, "R", false, 0, "")
	} else {
		pdf.CellFormat(nameW, 4, string(c.Name), "", 0, "L", false, 0, "")
	}
	cursorY += 5

	// Separator line
	pdf.SetLineWidth(0.1)
	pdf.Line(innerX, cursorY, x+cardW-padding, cursorY)
	cursorY += 1

	// Type line
	pdf.SetFont(fontFamily, "", 6)
	pdf.SetXY(innerX, cursorY)
	pdf.CellFormat(innerW, 3.5, string(c.TypeLine), "", 0, "L", false, 0, "")
	cursorY += 4.5

	// Separator line
	pdf.Line(innerX, cursorY, x+cardW-padding, cursorY)
	cursorY += 1

	// Oracle text
	pdf.SetFont(fontFamily, "", 5.5)
	pdf.SetXY(innerX, cursorY)

	oracleMaxH := y + cardH - padding - cursorY
	if c.Stats != nil || c.Loyalty != nil {
		oracleMaxH -= 5
	}

	// Use MultiCell for word-wrapped oracle text
	pdf.MultiCell(innerW, 3, string(c.OracleText), "", "L", false)
	cursorY += oracleMaxH

	// Footer: Stats or Loyalty
	if c.Stats != nil {
		footer := fmt.Sprintf("%s/%s", c.Stats.Power, c.Stats.Toughness)
		pdf.SetFont(fontFamily, "B", 7)
		pdf.SetXY(innerX, y+cardH-padding-4)
		pdf.CellFormat(innerW, 4, footer, "", 0, "R", false, 0, "")
	} else if c.Loyalty != nil {
		footer := fmt.Sprintf("Loyalty: %s", *c.Loyalty)
		pdf.SetFont(fontFamily, "B", 7)
		pdf.SetXY(innerX, y+cardH-padding-4)
		pdf.CellFormat(innerW, 4, footer, "", 0, "R", false, 0, "")
	}
}
