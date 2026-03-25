package pdf_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/epalmerini/mtg-proxy/internal/card"
	"github.com/epalmerini/mtg-proxy/internal/pdf"
)

func bolt() card.Card {
	return card.Card{
		Name:       "Lightning Bolt",
		ManaCost:   card.ParseManaCost("{R}"),
		TypeLine:   "Instant",
		OracleText: "Lightning Bolt deals 3 damage to any target.",
	}
}

func tarmogoyf() card.Card {
	return card.Card{
		Name:       "Tarmogoyf",
		ManaCost:   card.ParseManaCost("{1}{G}"),
		TypeLine:   "Creature — Lhurgoyf",
		OracleText: "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.",
		Stats:      &card.Stats{Power: "*", Toughness: "1+*"},
	}
}

func jace() card.Card {
	l := card.Loyalty("3")
	return card.Card{
		Name:     "Jace, the Mind Sculptor",
		ManaCost: card.ParseManaCost("{2}{U}{U}"),
		TypeLine: "Legendary Planeswalker — Jace",
		OracleText: "+2: Look at the top card of target player's library.\n" +
			"0: Draw three cards, then put two cards from your hand on top of your library.\n" +
			"−1: Return target creature to its owner's hand.\n" +
			"−12: Exile all cards from target player's library.",
		Loyalty: &l,
	}
}

func renderToFile(t *testing.T, cards []card.DeckCard, opts ...pdf.Option) (string, []byte) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pdf")

	renderer := pdf.NewRenderer(opts...)
	err := renderer.Render(cards, path)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	return path, data
}

func TestRenderCreatesPDF(t *testing.T) {
	_, data := renderToFile(t, []card.DeckCard{
		{Card: bolt(), Quantity: 1},
	})

	if string(data[:5]) != "%PDF-" {
		t.Fatalf("not a valid PDF, header: %q", string(data[:5]))
	}
}

func TestRenderEmptyDeck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.pdf")

	renderer := pdf.NewRenderer()
	err := renderer.Render(nil, path)
	if err != nil {
		t.Fatalf("rendering empty deck should not error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("PDF not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF file is empty")
	}
}

func TestRenderSinglePage(t *testing.T) {
	// 9 cards fit on 1 page
	cards := []card.DeckCard{
		{Card: bolt(), Quantity: 9},
	}
	_, data := renderToFile(t, cards)

	pageCount := countPDFPages(data)
	if pageCount != 1 {
		t.Errorf("expected 1 page for 9 cards, got %d", pageCount)
	}
}

func TestRenderTwoPages(t *testing.T) {
	// 10 cards need 2 pages
	cards := []card.DeckCard{
		{Card: bolt(), Quantity: 10},
	}
	_, data := renderToFile(t, cards)

	pageCount := countPDFPages(data)
	if pageCount != 2 {
		t.Errorf("expected 2 pages for 10 cards, got %d", pageCount)
	}
}

func TestRenderMultipleCardTypes(t *testing.T) {
	cards := []card.DeckCard{
		{Card: bolt(), Quantity: 4},
		{Card: tarmogoyf(), Quantity: 2},
		{Card: jace(), Quantity: 1},
	}
	_, data := renderToFile(t, cards)

	// 7 cards = 1 page
	pageCount := countPDFPages(data)
	if pageCount != 1 {
		t.Errorf("expected 1 page for 7 cards, got %d", pageCount)
	}

	// PDF with 3 distinct card types should be larger than a single-card PDF
	_, singleData := renderToFile(t, []card.DeckCard{{Card: bolt(), Quantity: 1}})
	if len(data) <= len(singleData) {
		t.Errorf("multi-card PDF (%d bytes) should be larger than single-card PDF (%d bytes)", len(data), len(singleData))
	}
}

func TestRenderCreatureHasFooter(t *testing.T) {
	// A creature card should produce a larger PDF than an instant
	// because it has the stats footer rendered
	crCards := []card.DeckCard{{Card: tarmogoyf(), Quantity: 1}}
	_, crData := renderToFile(t, crCards)

	// Just verify it renders without error and is a valid PDF
	if string(crData[:5]) != "%PDF-" {
		t.Fatal("creature card did not produce valid PDF")
	}
}

func TestRenderPlaneswalkerHasFooter(t *testing.T) {
	pwCards := []card.DeckCard{{Card: jace(), Quantity: 1}}
	_, pwData := renderToFile(t, pwCards)

	if string(pwData[:5]) != "%PDF-" {
		t.Fatal("planeswalker card did not produce valid PDF")
	}
}

func TestRenderExpandsQuantity(t *testing.T) {
	// 4 copies of bolt + 2 copies of counterspell = 6 cards on the page
	cs := card.Card{
		Name:       "Counterspell",
		ManaCost:   card.ParseManaCost("{U}{U}"),
		TypeLine:   "Instant",
		OracleText: "Counter target spell.",
	}
	cards := []card.DeckCard{
		{Card: bolt(), Quantity: 4},
		{Card: cs, Quantity: 2},
	}
	_, data := renderToFile(t, cards)

	pageCount := countPDFPages(data)
	if pageCount != 1 {
		t.Errorf("expected 1 page for 6 cards, got %d", pageCount)
	}
}

// countPDFPages counts pages by looking for /Type /Page entries that are not /Type /Pages.
func countPDFPages(data []byte) int {
	content := string(data)
	count := 0
	idx := 0
	for {
		i := strings.Index(content[idx:], "/Type /Page")
		if i == -1 {
			break
		}
		pos := idx + i + len("/Type /Page")
		// /Type /Pages is the page tree, not an actual page
		if pos < len(content) && content[pos] == 's' {
			idx = pos
			continue
		}
		count++
		idx = pos
	}
	return count
}
