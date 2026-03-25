package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/epalmerini/mtg-proxy/internal/app"
	"github.com/epalmerini/mtg-proxy/internal/card"
)

// scryfallCard mirrors the relevant fields from Scryfall's API response.
type scryfallCard struct {
	Object     string  `json:"object"`
	Name       string  `json:"name"`
	ManaCost   string  `json:"mana_cost"`
	TypeLine   string  `json:"type_line"`
	OracleText string  `json:"oracle_text"`
	Power      *string `json:"power,omitempty"`
	Toughness  *string `json:"toughness,omitempty"`
	Loyalty    *string `json:"loyalty,omitempty"`
}

func ptr(s string) *string { return &s }

var testCards = map[string]scryfallCard{
	"Lightning Bolt": {
		Object:     "card",
		Name:       "Lightning Bolt",
		ManaCost:   "{R}",
		TypeLine:   "Instant",
		OracleText: "Lightning Bolt deals 3 damage to any target.",
	},
	"Counterspell": {
		Object:     "card",
		Name:       "Counterspell",
		ManaCost:   "{U}{U}",
		TypeLine:   "Instant",
		OracleText: "Counter target spell.",
	},
	"Tarmogoyf": {
		Object:     "card",
		Name:       "Tarmogoyf",
		ManaCost:   "{1}{G}",
		TypeLine:   "Creature — Lhurgoyf",
		OracleText: "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.",
		Power:      ptr("*"),
		Toughness:  ptr("1+*"),
	},
	"Jace, the Mind Sculptor": {
		Object:     "card",
		Name:       "Jace, the Mind Sculptor",
		ManaCost:   "{2}{U}{U}",
		TypeLine:   "Legendary Planeswalker — Jace",
		OracleText: "+2: Look at the top card of target player's library. You may put that card on the bottom of that player's library.\n0: Draw three cards, then put two cards from your hand on top of your library in any order.\n−1: Return target creature to its owner's hand.\n−12: Exile all cards from target player's library, then that player shuffles their hand into their library.",
		Loyalty:    ptr("3"),
	},
}

func newMockScryfall(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("exact")
		card, ok := testCards[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"object":  "error",
				"details": fmt.Sprintf("Card not found: %s", name),
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(card)
	}))
}

func TestGenerateFullPipeline(t *testing.T) {
	server := newMockScryfall(t)
	defer server.Close()

	decklistPath := filepath.Join("..", "testdata", "sample_deck.txt")
	content, err := os.ReadFile(decklistPath)
	if err != nil {
		t.Fatalf("failed to read decklist: %v", err)
	}

	// Step 1: Parse decklist
	parser := app.NewDecklistParser()
	entries, err := parser.Parse(string(content))
	if err != nil {
		t.Fatalf("failed to parse decklist: %v", err)
	}

	expectedEntries := []struct {
		name     card.CardName
		quantity card.Quantity
	}{
		{"Lightning Bolt", 4},
		{"Counterspell", 2},
		{"Tarmogoyf", 1},
		{"Jace, the Mind Sculptor", 1},
	}

	if len(entries) != len(expectedEntries) {
		t.Fatalf("expected %d entries, got %d", len(expectedEntries), len(entries))
	}
	for i, exp := range expectedEntries {
		if entries[i].Name != exp.name {
			t.Errorf("entry[%d] name: expected %q, got %q", i, exp.name, entries[i].Name)
		}
		if entries[i].Quantity != exp.quantity {
			t.Errorf("entry[%d] quantity: expected %d, got %d", i, exp.quantity, entries[i].Quantity)
		}
	}

	// Step 2: Fetch cards from mock Scryfall
	fetcher := app.NewCardFetcher(server.URL)
	var deckCards []card.DeckCard
	for _, entry := range entries {
		c, err := fetcher.FetchCard(entry.Name)
		if err != nil {
			t.Fatalf("failed to fetch card %q: %v", entry.Name, err)
		}
		deckCards = append(deckCards, card.DeckCard{Card: c, Quantity: entry.Quantity})
	}

	// Verify card data was correctly mapped
	assertCard(t, deckCards[0].Card, card.Card{
		Name:       "Lightning Bolt",
		ManaCost:   card.ParseManaCost("{R}"),
		TypeLine:   "Instant",
		OracleText: "Lightning Bolt deals 3 damage to any target.",
		Stats:      nil,
		Loyalty:    nil,
	})
	assertCard(t, deckCards[2].Card, card.Card{
		Name:       "Tarmogoyf",
		ManaCost:   card.ParseManaCost("{1}{G}"),
		TypeLine:   "Creature — Lhurgoyf",
		OracleText: "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.",
		Stats:      &card.Stats{Power: "*", Toughness: "1+*"},
		Loyalty:    nil,
	})

	jLoyalty := card.Loyalty("3")
	assertCard(t, deckCards[3].Card, card.Card{
		Name:       "Jace, the Mind Sculptor",
		ManaCost:   card.ParseManaCost("{2}{U}{U}"),
		TypeLine:   "Legendary Planeswalker — Jace",
		OracleText: card.OracleText(testCards["Jace, the Mind Sculptor"].OracleText),
		Stats:      nil,
		Loyalty:    &jLoyalty,
	})

	// Step 3: Render PDF
	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "proxies.pdf")

	renderer := app.NewDeckRenderer()
	err = renderer.Render(deckCards, outputPath)
	if err != nil {
		t.Fatalf("failed to render PDF: %v", err)
	}

	// Verify PDF was created and is non-empty
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("PDF file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF file is empty")
	}

	// Verify PDF starts with the correct magic bytes
	pdfBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read PDF: %v", err)
	}
	if string(pdfBytes[:5]) != "%PDF-" {
		t.Fatalf("file does not start with PDF header, got: %q", string(pdfBytes[:5]))
	}
}

func TestGenerateWithUnknownCard(t *testing.T) {
	server := newMockScryfall(t)
	defer server.Close()

	decklist := "4 Lightning Bolt\n1 Nonexistent Card\n2 Counterspell\n"

	parser := app.NewDecklistParser()
	entries, err := parser.Parse(decklist)
	if err != nil {
		t.Fatalf("failed to parse decklist: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Unknown card should return an error from fetcher
	fetcher := app.NewCardFetcher(server.URL)
	_, err = fetcher.FetchCard(entries[1].Name)
	if err == nil {
		t.Fatal("expected error for unknown card, got nil")
	}
}

func TestGenerateCorrectCardCount(t *testing.T) {
	server := newMockScryfall(t)
	defer server.Close()

	decklist := "4 Lightning Bolt\n2 Counterspell\n"

	parser := app.NewDecklistParser()
	entries, err := parser.Parse(decklist)
	if err != nil {
		t.Fatalf("failed to parse decklist: %v", err)
	}

	fetcher := app.NewCardFetcher(server.URL)
	var deckCards []card.DeckCard
	for _, entry := range entries {
		c, err := fetcher.FetchCard(entry.Name)
		if err != nil {
			t.Fatalf("failed to fetch card %q: %v", entry.Name, err)
		}
		deckCards = append(deckCards, card.DeckCard{Card: c, Quantity: entry.Quantity})
	}

	// Total cards should be 4 + 2 = 6, which fits on 1 page (9 per page)
	totalCards := 0
	for _, dc := range deckCards {
		totalCards += int(dc.Quantity)
	}
	if totalCards != 6 {
		t.Errorf("expected 6 total cards, got %d", totalCards)
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "proxies.pdf")

	renderer := app.NewDeckRenderer()
	err = renderer.Render(deckCards, outputPath)
	if err != nil {
		t.Fatalf("failed to render PDF: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("PDF not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("PDF is empty")
	}
}

func TestGenerateMultiplePages(t *testing.T) {
	server := newMockScryfall(t)
	defer server.Close()

	// 10 cards total = 2 pages (9 per page)
	decklist := "10 Lightning Bolt\n"

	parser := app.NewDecklistParser()
	entries, err := parser.Parse(decklist)
	if err != nil {
		t.Fatalf("failed to parse decklist: %v", err)
	}

	fetcher := app.NewCardFetcher(server.URL)
	var deckCards []card.DeckCard
	for _, entry := range entries {
		c, err := fetcher.FetchCard(entry.Name)
		if err != nil {
			t.Fatalf("failed to fetch: %v", err)
		}
		deckCards = append(deckCards, card.DeckCard{Card: c, Quantity: entry.Quantity})
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "proxies.pdf")

	renderer := app.NewDeckRenderer()
	err = renderer.Render(deckCards, outputPath)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("PDF not created: %v", err)
	}
	// 2-page PDF should be larger than a 1-page one, but just check it exists and is non-trivial
	if info.Size() < 100 {
		t.Fatalf("PDF seems too small for 2 pages: %d bytes", info.Size())
	}
}

func TestGenerateEmptyDecklist(t *testing.T) {
	parser := app.NewDecklistParser()

	// Empty or comment-only decklist should return no entries
	entries, err := parser.Parse("# just a comment\n\n// another comment\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}

	// Rendering empty deck should not error
	renderer := app.NewDeckRenderer()
	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "empty.pdf")
	err = renderer.Render(nil, outputPath)
	if err != nil {
		t.Fatalf("rendering empty deck should not error: %v", err)
	}
}

func assertCard(t *testing.T, got, want card.Card) {
	t.Helper()
	if got.Name != want.Name {
		t.Errorf("name: got %q, want %q", got.Name, want.Name)
	}
	if got.ManaCost.String() != want.ManaCost.String() {
		t.Errorf("mana cost: got %q, want %q", got.ManaCost, want.ManaCost)
	}
	if got.TypeLine != want.TypeLine {
		t.Errorf("type line: got %q, want %q", got.TypeLine, want.TypeLine)
	}
	if got.OracleText != want.OracleText {
		t.Errorf("oracle text: got %q, want %q", got.OracleText, want.OracleText)
	}
	if (got.Stats == nil) != (want.Stats == nil) {
		t.Errorf("stats: got %v, want %v", got.Stats, want.Stats)
	} else if got.Stats != nil && *got.Stats != *want.Stats {
		t.Errorf("stats: got %v, want %v", *got.Stats, *want.Stats)
	}
	if (got.Loyalty == nil) != (want.Loyalty == nil) {
		t.Errorf("loyalty: got %v, want %v", got.Loyalty, want.Loyalty)
	} else if got.Loyalty != nil && *got.Loyalty != *want.Loyalty {
		t.Errorf("loyalty: got %q, want %q", *got.Loyalty, *want.Loyalty)
	}
}
