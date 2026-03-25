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

func renderCard(p *fpdf.Fpdf, c card.Card, x, y float64) {
	// Dashed card border (saves ink)
	p.SetDrawColor(128, 128, 128)
	p.SetLineWidth(0.2)
	p.SetDashPattern([]float64{1.5, 1.0}, 0)
	p.Rect(x, y, cardW, cardH, "D")
	p.SetDashPattern([]float64{}, 0)

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
	p.Line(innerX, cursorY, x+cardW-padding, cursorY)
	cursorY += 1

	// Type line
	p.SetFont(fontName, "", 6)
	p.SetXY(innerX, cursorY)
	p.CellFormat(innerW, 3.5, string(c.TypeLine), "", 0, "L", false, 0, "")
	cursorY += 4.5

	// Separator
	p.Line(innerX, cursorY, x+cardW-padding, cursorY)
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
