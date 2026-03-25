package decklist_test

import (
	"testing"

	"github.com/epalmerini/mtg-proxy/internal/card"
	"github.com/epalmerini/mtg-proxy/internal/decklist"
)

func TestParseStandardEntries(t *testing.T) {
	input := "4 Lightning Bolt\n2 Counterspell\n1 Tarmogoyf\n"

	parser := decklist.NewParser()
	entries, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []card.DeckEntry{
		{Name: "Lightning Bolt", Quantity: 4},
		{Name: "Counterspell", Quantity: 2},
		{Name: "Tarmogoyf", Quantity: 1},
	}

	assertEntries(t, entries, expected)
}

func TestParseIgnoresEmptyLines(t *testing.T) {
	input := "\n4 Lightning Bolt\n\n2 Counterspell\n\n"

	entries, err := decklist.NewParser().Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEntries(t, entries, []card.DeckEntry{
		{Name: "Lightning Bolt", Quantity: 4},
		{Name: "Counterspell", Quantity: 2},
	})
}

func TestParseIgnoresComments(t *testing.T) {
	input := "# Mainboard\n4 Lightning Bolt\n// Sideboard\n1 Counterspell\n"

	entries, err := decklist.NewParser().Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEntries(t, entries, []card.DeckEntry{
		{Name: "Lightning Bolt", Quantity: 4},
		{Name: "Counterspell", Quantity: 1},
	})
}

func TestParseCardNameWithComma(t *testing.T) {
	input := "1 Jace, the Mind Sculptor\n"

	entries, err := decklist.NewParser().Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEntries(t, entries, []card.DeckEntry{
		{Name: "Jace, the Mind Sculptor", Quantity: 1},
	})
}

func TestParseReturnsEmptyForBlankInput(t *testing.T) {
	entries, err := decklist.NewParser().Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestParseReturnsErrorForInvalidLine(t *testing.T) {
	input := "Lightning Bolt\n"

	_, err := decklist.NewParser().Parse(input)
	if err == nil {
		t.Fatal("expected error for line without quantity, got nil")
	}
}

func TestParseReturnsErrorForZeroQuantity(t *testing.T) {
	input := "0 Lightning Bolt\n"

	_, err := decklist.NewParser().Parse(input)
	if err == nil {
		t.Fatal("expected error for zero quantity, got nil")
	}
}

func TestParseReturnsErrorForNegativeQuantity(t *testing.T) {
	input := "-1 Lightning Bolt\n"

	_, err := decklist.NewParser().Parse(input)
	if err == nil {
		t.Fatal("expected error for negative quantity, got nil")
	}
}

func assertEntries(t *testing.T, got, want []card.DeckEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d entries, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i].Name != want[i].Name {
			t.Errorf("[%d] name: got %q, want %q", i, got[i].Name, want[i].Name)
		}
		if got[i].Quantity != want[i].Quantity {
			t.Errorf("[%d] quantity: got %d, want %d", i, got[i].Quantity, want[i].Quantity)
		}
	}
}
