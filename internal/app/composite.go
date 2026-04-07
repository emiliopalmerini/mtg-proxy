package app

import (
	"github.com/epalmerini/mtg-proxy/internal/bulkdata"
	"github.com/epalmerini/mtg-proxy/internal/card"
	"github.com/epalmerini/mtg-proxy/internal/scryfall"
)

// CompositeFetcher tries the local bulk data store first, then falls back to
// the Scryfall API for cards not found locally or for set/collector lookups.
type CompositeFetcher struct {
	local  *bulkdata.Store
	remote *scryfall.Client
}

// NewCompositeFetcher creates a CompositeFetcher with a local SQLite store and
// a Scryfall API client as fallback. Returns the fetcher and a cleanup function
// that closes the database.
func NewCompositeFetcher(dbPath, bulkBaseURL, scryfallBaseURL string) (*CompositeFetcher, func(), error) {
	store, err := bulkdata.NewStore(dbPath, bulkBaseURL)
	if err != nil {
		return nil, nil, err
	}

	client := scryfall.NewClient(scryfallBaseURL)

	f := &CompositeFetcher{local: store, remote: client}
	cleanup := func() { store.Close() }
	return f, cleanup, nil
}

// FetchCard looks up a card by trying the local database first, then falling
// back to the Scryfall API if the card is not found locally.
func (f *CompositeFetcher) FetchCard(entry card.DeckEntry) (card.Card, error) {
	c, err := f.local.FetchCard(entry)
	if err == nil {
		return c, nil
	}

	return f.remote.FetchCard(entry)
}
