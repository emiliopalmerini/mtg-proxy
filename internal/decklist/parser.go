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

		qtyStr, rest, ok := strings.Cut(line, " ")
		if !ok {
			return nil, fmt.Errorf("line %d: expected format '<quantity> <card name>', got %q", i+1, line)
		}

		qtyStr = strings.TrimSuffix(qtyStr, "x")
		qty, err := strconv.Atoi(qtyStr)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid quantity %q: %w", i+1, qtyStr, err)
		}
		if qty <= 0 {
			return nil, fmt.Errorf("line %d: quantity must be positive, got %d", i+1, qty)
		}

		name, setCode, collectorNum, isCommander := parseCardDetails(rest)

		entries = append(entries, card.DeckEntry{
			Name:            card.CardName(name),
			Quantity:        card.Quantity(qty),
			SetCode:         card.SetCode(setCode),
			CollectorNumber: card.CollectorNumber(collectorNum),
			IsCommander:     isCommander,
		})
	}

	return entries, nil
}

// parseCardDetails extracts card name, set code, and collector number from
// the remainder of a decklist line after the quantity.
// Handles both simple ("Lightning Bolt") and rich ("Ankh of Mishra (6ed) 273 [Slug]") formats.
func parseCardDetails(s string) (name, setCode, collectorNum string, isCommander bool) {
	s = strings.TrimSpace(s)

	// Extract and strip tags: [...]
	if idx := strings.Index(s, "["); idx != -1 {
		closeIdx := strings.Index(s[idx:], "]")
		if closeIdx != -1 {
			tagContent := s[idx+1 : idx+closeIdx]
			for _, tag := range strings.Split(tagContent, ",") {
				tag = strings.TrimSpace(tag)
				// Strip modifiers like {top} from tag
				if braceIdx := strings.Index(tag, "{"); braceIdx != -1 {
					tag = tag[:braceIdx]
				}
				if strings.EqualFold(tag, "Commander") {
					isCommander = true
				}
			}
		}
		s = strings.TrimSpace(s[:idx])
	}

	// Extract set code and collector number: Name (set) number
	if idx := strings.Index(s, "("); idx != -1 {
		name = strings.TrimSpace(s[:idx])
		rest := s[idx+1:]

		closeIdx := strings.Index(rest, ")")
		if closeIdx != -1 {
			setCode = rest[:closeIdx]
			collectorNum = strings.TrimSpace(rest[closeIdx+1:])
		}
		return
	}

	name = s
	return
}
