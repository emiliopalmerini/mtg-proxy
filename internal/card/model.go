package card

import (
	"fmt"
	"image"
	"regexp"
	"strings"
)

type ManaCost struct { symbols []ManaSymbol }
type ManaSymbol struct { raw string }
var manaSymbolRegex = regexp.MustCompile(`\{([^}]+)\}`)

func ParseManaCost(raw string) ManaCost {
	matches := manaSymbolRegex.FindAllStringSubmatch(raw, -1)
	symbols := make([]ManaSymbol, 0, len(matches))
	for _, m := range matches { symbols = append(symbols, ManaSymbol{raw: m[1]}) }
	return ManaCost{symbols: symbols}
}
func (mc ManaCost) String() string { var b strings.Builder; for _, s := range mc.symbols { b.WriteString(s.raw) }; return b.String() }
func (mc ManaCost) IsEmpty() bool { return len(mc.symbols) == 0 }

type CardName string
func (n CardName) String() string { return string(n) }
type TypeLine string
func (t TypeLine) String() string { return string(t) }
func (t TypeLine) IsCreature() bool { return strings.Contains(string(t), "Creature") }
func (t TypeLine) IsPlaneswalker() bool { return strings.Contains(string(t), "Planeswalker") }
func (t TypeLine) IsBasicLand() bool { s := string(t); return strings.Contains(s, "Basic Land") || strings.Contains(s, "Basic Snow Land") }
type OracleText string
func (o OracleText) String() string { return string(o) }
type Stats struct { Power string; Toughness string }
func (s Stats) String() string { return fmt.Sprintf("%s/%s", s.Power, s.Toughness) }
type Loyalty string
func (l Loyalty) String() string { return string(l) }
type Quantity int
type SetCode string
func (s SetCode) String() string { return string(s) }
func (s SetCode) IsEmpty() bool { return s == "" }
type CollectorNumber string
func (c CollectorNumber) String() string { return string(c) }
func (c CollectorNumber) IsEmpty() bool { return c == "" }

type CardFace struct {
	Name CardName
	ManaCost ManaCost
	TypeLine TypeLine
	OracleText OracleText
	Stats *Stats
	Loyalty *Loyalty
	ArtCropURL string
}

type Card struct {
	Faces []CardFace
	ArtCropURL string
}
func (c Card) Front() CardFace { return c.Faces[0] }
func (c Card) IsMultiFaced() bool { return len(c.Faces) > 1 }
func (c Card) IsBasicLand() bool { for _, f := range c.Faces { if f.TypeLine.IsBasicLand() { return true } }; return false }

type DeckEntry struct {
	Name CardName
	Quantity Quantity
	SetCode SetCode
	CollectorNumber CollectorNumber
	IsCommander bool
}

type DeckCard struct {
	Card Card
	Quantity Quantity
	IsCommander bool
	// ArtImages is aligned with Card.Faces. Single-faced cards contain one image.
	ArtImages []image.Image
}

type CardFetcher interface { FetchCard(entry DeckEntry) (Card, error) }
type DeckRenderer interface { Render(cards []DeckCard, outputPath string) error }
type DecklistParser interface { Parse(content string) ([]DeckEntry, error) }
