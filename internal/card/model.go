package card

import (
	"fmt"
	"regexp"
	"strings"
)

// ManaCost represents a card's mana cost, parsed from Scryfall format.
type ManaCost struct {
	symbols []ManaSymbol
}

// ManaSymbol represents a single mana symbol (e.g., W, U, 2, X, W/U).
type ManaSymbol struct {
	raw string
}

var manaSymbolRegex = regexp.MustCompile(`\{([^}]+)\}`)

// ParseManaCost parses a Scryfall mana cost string like "{1}{W}{W}" into a ManaCost.
func ParseManaCost(raw string) ManaCost {
	matches := manaSymbolRegex.FindAllStringSubmatch(raw, -1)
	symbols := make([]ManaSymbol, 0, len(matches))
	for _, m := range matches {
		symbols = append(symbols, ManaSymbol{raw: m[1]})
	}
	return ManaCost{symbols: symbols}
}

// String returns the simplified text representation (e.g., "1WW", "W/U").
func (mc ManaCost) String() string {
	var b strings.Builder
	for _, s := range mc.symbols {
		b.WriteString(s.raw)
	}
	return b.String()
}

// IsEmpty returns true if the card has no mana cost (e.g., lands).
func (mc ManaCost) IsEmpty() bool {
	return len(mc.symbols) == 0
}

// CardName represents a card's name.
type CardName string

func (n CardName) String() string { return string(n) }

// TypeLine represents a card's type line (e.g., "Creature — Human Wizard").
type TypeLine string

func (t TypeLine) String() string { return string(t) }

// IsCreature returns true if the type line contains "Creature".
func (t TypeLine) IsCreature() bool {
	return strings.Contains(string(t), "Creature")
}

// IsPlaneswalker returns true if the type line contains "Planeswalker".
func (t TypeLine) IsPlaneswalker() bool {
	return strings.Contains(string(t), "Planeswalker")
}

// OracleText represents a card's rules text.
type OracleText string

func (o OracleText) String() string { return string(o) }

// Stats represents a creature's power and toughness.
type Stats struct {
	Power     string
	Toughness string
}

func (s Stats) String() string {
	return fmt.Sprintf("%s/%s", s.Power, s.Toughness)
}

// Loyalty represents a planeswalker's starting loyalty.
type Loyalty string

func (l Loyalty) String() string { return string(l) }

// Quantity represents a card count in a decklist.
type Quantity int

// Card represents a Magic: The Gathering card with its relevant fields.
type Card struct {
	Name       CardName
	ManaCost   ManaCost
	TypeLine   TypeLine
	OracleText OracleText
	Stats      *Stats
	Loyalty    *Loyalty
}

// DeckEntry represents a card entry in a decklist with its quantity.
type DeckEntry struct {
	Name     CardName
	Quantity Quantity
}

// DeckCard combines a resolved card with its quantity in the deck.
type DeckCard struct {
	Card     Card
	Quantity Quantity
}

// CardFetcher retrieves card data by name.
type CardFetcher interface {
	FetchCard(name CardName) (Card, error)
}

// DeckRenderer renders a list of deck cards to an output file.
type DeckRenderer interface {
	Render(cards []DeckCard, outputPath string) error
}

// DecklistParser parses a decklist from raw text.
type DecklistParser interface {
	Parse(content string) ([]DeckEntry, error)
}
