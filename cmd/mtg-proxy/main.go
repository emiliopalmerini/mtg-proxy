package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/epalmerini/mtg-proxy/internal/app"
	"github.com/epalmerini/mtg-proxy/internal/card"
)

const scryfallBaseURL = "https://api.scryfall.com"

func main() {
	var inputPath, outputPath string
	var skipBasics bool
	flag.StringVar(&inputPath, "i", "", "path to decklist file (required)")
	flag.StringVar(&outputPath, "o", "proxies.pdf", "path to output PDF")
	flag.BoolVar(&skipBasics, "skip-basics", false, "skip basic lands")
	flag.Parse()

	if inputPath == "" {
		fmt.Fprintln(os.Stderr, "error: -i flag is required")
		flag.Usage()
		os.Exit(1)
	}

	content, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading decklist: %v\n", err)
		os.Exit(1)
	}

	parser := app.NewDecklistParser()
	entries, err := parser.Parse(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing decklist: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "warning: decklist is empty")
		os.Exit(0)
	}

	fetcher := app.NewCardFetcher(scryfallBaseURL)
	var deckCards []card.DeckCard
	for _, entry := range entries {
		fmt.Fprintf(os.Stderr, "fetching %s...\n", entry.Name)
		c, err := fetcher.FetchCard(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %q: %v\n", entry.Name, err)
			continue
		}
		if skipBasics && c.TypeLine.IsBasicLand() {
			fmt.Fprintf(os.Stderr, "skipping basic land: %s\n", c.Name)
			continue
		}
		deckCards = append(deckCards, card.DeckCard{Card: c, Quantity: entry.Quantity})
	}

	if len(deckCards) == 0 {
		fmt.Fprintln(os.Stderr, "error: no cards resolved, nothing to render")
		os.Exit(1)
	}

	renderer := app.NewDeckRenderer()
	if err := renderer.Render(deckCards, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "error rendering PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "done: %d cards written to %s\n", len(deckCards), outputPath)
}
