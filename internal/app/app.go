package app

import (
	"github.com/epalmerini/mtg-proxy/internal/card"
	"github.com/epalmerini/mtg-proxy/internal/decklist"
	"github.com/epalmerini/mtg-proxy/internal/pdf"
	"github.com/epalmerini/mtg-proxy/internal/scryfall"
)

func NewDecklistParser() card.DecklistParser {
	return decklist.NewParser()
}

func NewCardFetcher(baseURL string) card.CardFetcher {
	return scryfall.NewClient(baseURL)
}

func NewDeckRenderer(opts ...pdf.Option) card.DeckRenderer {
	return pdf.NewRenderer(opts...)
}
