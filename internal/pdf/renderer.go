package pdf

import (
	"fmt"

	"github.com/epalmerini/mtg-proxy/internal/card"
	"github.com/go-pdf/fpdf"
)

const (
	pageW = 210.0
	pageH = 297.0

	cardW = 63.0
	cardH = 88.0

	cols = 3
	rows = 3

	cardsPerPage = cols * rows

	marginX = (pageW - cols*cardW) / 2
	marginY = (pageH - rows*cardH) / 2

	padding  = 2.0
	fontName = "dejavu"
)

// Renderer generates a printable PDF of proxy cards.
type Renderer struct {
	compress bool
}

type Option func(*Renderer)

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
	p := fpdf.New("P", "mm", "A4", "")
	p.SetCompression(r.compress)
	p.SetAutoPageBreak(false, 0)

	p.AddUTF8FontFromBytes(fontName, "", dejaVuRegular)
	p.AddUTF8FontFromBytes(fontName, "B", dejaVuBold)

	expanded := expandDeck(cards)

	if len(expanded) == 0 {
		p.AddPage()
	}

	for i, c := range expanded {
		if i%cardsPerPage == 0 {
			p.AddPage()
			remaining := len(expanded) - i
			cardsOnPage := remaining
			if cardsOnPage > cardsPerPage {
				cardsOnPage = cardsPerPage
			}
			drawGrid(p, cardsOnPage)
		}

		pos := i % cardsPerPage
		col := pos % cols
		row := pos / cols

		x := marginX + float64(col)*cardW
		y := marginY + float64(row)*cardH

		renderCard(p, c, x, y)
	}

	return p.OutputFileAndClose(outputPath)
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

func drawGrid(p *fpdf.Fpdf, cardCount int) {
	p.SetDrawColor(0, 0, 0)
	p.SetLineWidth(0.2)
	p.SetDashPattern([]float64{0.3, 1.5}, 0)

	usedRows := (cardCount + cols - 1) / cols

	// Horizontal lines
	for r := 0; r <= usedRows; r++ {
		y := marginY + float64(r)*cardH
		colsInRow := cols
		if r == usedRows {
			// last row: only draw as wide as the cards in the previous row
			lastRowCards := cardCount - (usedRows-1)*cols
			colsInRow = lastRowCards
		}
		if r < usedRows {
			colsInRow = cols
		}
		p.Line(marginX, y, marginX+float64(colsInRow)*cardW, y)
	}

	// Vertical lines
	for r := 0; r < usedRows; r++ {
		cardsInRow := cols
		if r == usedRows-1 {
			cardsInRow = cardCount - r*cols
		}
		for c := 0; c <= cardsInRow; c++ {
			x := marginX + float64(c)*cardW
			yTop := marginY + float64(r)*cardH
			yBot := yTop + cardH
			p.Line(x, yTop, x, yBot)
		}
	}

	p.SetDashPattern([]float64{}, 0)
}

func renderCard(p *fpdf.Fpdf, c card.Card, x, y float64) {
	innerX := x + padding
	innerW := cardW - 2*padding
	cursorY := y + padding

	// Header: Name (left) + Mana cost (right)
	p.SetFont(fontName, "B", 7)
	p.SetXY(innerX, cursorY)

	manaCost := c.ManaCost.String()
	nameW := innerW
	if manaCost != "" {
		costW := p.GetStringWidth(manaCost) + 1
		nameW = innerW - costW
		p.CellFormat(nameW, 4, string(c.Name), "", 0, "L", false, 0, "")
		p.CellFormat(costW, 4, manaCost, "", 0, "R", false, 0, "")
	} else {
		p.CellFormat(nameW, 4, string(c.Name), "", 0, "L", false, 0, "")
	}
	cursorY += 5

	// Separator
	p.SetDrawColor(0, 0, 0)
	p.SetLineWidth(0.1)
	p.SetDashPattern([]float64{0.3, 1.5}, 0)
	p.Line(innerX, cursorY, x+cardW-padding, cursorY)
	p.SetDashPattern([]float64{}, 0)
	cursorY += 1

	// Type line
	p.SetFont(fontName, "", 6)
	p.SetXY(innerX, cursorY)
	p.CellFormat(innerW, 3.5, string(c.TypeLine), "", 0, "L", false, 0, "")
	cursorY += 4.5

	// Separator
	p.SetDashPattern([]float64{0.3, 1.5}, 0)
	p.Line(innerX, cursorY, x+cardW-padding, cursorY)
	p.SetDashPattern([]float64{}, 0)
	cursorY += 1

	// Oracle text - clip to available height
	footerH := 0.0
	if c.Stats != nil || c.Loyalty != nil {
		footerH = 5.0
	}
	oracleMaxY := y + cardH - padding - footerH

	p.SetFont(fontName, "", 5.5)
	p.ClipRect(innerX, cursorY, innerW, oracleMaxY-cursorY, false)
	p.SetXY(innerX, cursorY)
	p.MultiCell(innerW, 3, string(c.OracleText), "", "L", false)
	p.ClipEnd()

	// Footer: Stats or Loyalty
	if c.Stats != nil {
		footer := fmt.Sprintf("%s/%s", c.Stats.Power, c.Stats.Toughness)
		p.SetFont(fontName, "B", 7)
		p.SetXY(innerX, y+cardH-padding-4)
		p.CellFormat(innerW, 4, footer, "", 0, "R", false, 0, "")
	} else if c.Loyalty != nil {
		footer := fmt.Sprintf("Loyalty: %s", *c.Loyalty)
		p.SetFont(fontName, "B", 7)
		p.SetXY(innerX, y+cardH-padding-4)
		p.CellFormat(innerW, 4, footer, "", 0, "R", false, 0, "")
	}
}
