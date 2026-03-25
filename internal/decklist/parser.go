package decklist

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/epalmerini/mtg-proxy/internal/card"
)

// Parser implements card.DecklistParser for the standard decklist format.
type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(content string) ([]card.DeckEntry, error) {
	var entries []card.DeckEntry

	for i, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		qtyStr, name, ok := strings.Cut(line, " ")
		if !ok {
			return nil, fmt.Errorf("line %d: expected format '<quantity> <card name>', got %q", i+1, line)
		}

		qty, err := strconv.Atoi(qtyStr)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid quantity %q: %w", i+1, qtyStr, err)
		}
		if qty <= 0 {
			return nil, fmt.Errorf("line %d: quantity must be positive, got %d", i+1, qty)
		}

		name = strings.TrimSpace(name)
		entries = append(entries, card.DeckEntry{
			Name:     card.CardName(name),
			Quantity: card.Quantity(qty),
		})
	}

	return entries, nil
}
