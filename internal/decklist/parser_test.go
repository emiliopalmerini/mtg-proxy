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

func TestParseRichFormat(t *testing.T) {
	input := "1x Ankh of Mishra (6ed) 273 [Slug]\n"

	entries, err := decklist.NewParser().Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEntries(t, entries, []card.DeckEntry{
		{Name: "Ankh of Mishra", Quantity: 1, SetCode: "6ed", CollectorNumber: "273"},
	})
}

func TestParseRichFormatMultipleEntries(t *testing.T) {
	input := "1x Ankh of Mishra (6ed) 273 [Slug]\n4 Lightning Bolt\n1x Athreos, God of Passage (plst) JOU-146 [Protection,Creature]\n"

	entries, err := decklist.NewParser().Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEntries(t, entries, []card.DeckEntry{
		{Name: "Ankh of Mishra", Quantity: 1, SetCode: "6ed", CollectorNumber: "273"},
		{Name: "Lightning Bolt", Quantity: 4},
		{Name: "Athreos, God of Passage", Quantity: 1, SetCode: "plst", CollectorNumber: "JOU-146"},
	})
}

func TestParseRichFormatSplitCard(t *testing.T) {
	input := "1x Boggart Trawler // Boggart Bog (mh3) 243 [Land,Creature,Stax]\n"

	entries, err := decklist.NewParser().Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEntries(t, entries, []card.DeckEntry{
		{Name: "Boggart Trawler // Boggart Bog", Quantity: 1, SetCode: "mh3", CollectorNumber: "243"},
	})
}

func TestParseRichFormatLargeQuantity(t *testing.T) {
	input := "12x Plains (ecl) 274 [Land]\n"

	entries, err := decklist.NewParser().Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEntries(t, entries, []card.DeckEntry{
		{Name: "Plains", Quantity: 12, SetCode: "ecl", CollectorNumber: "274"},
	})
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
		if got[i].SetCode != want[i].SetCode {
			t.Errorf("[%d] set: got %q, want %q", i, got[i].SetCode, want[i].SetCode)
		}
		if got[i].CollectorNumber != want[i].CollectorNumber {
			t.Errorf("[%d] collector: got %q, want %q", i, got[i].CollectorNumber, want[i].CollectorNumber)
		}
	}
}
