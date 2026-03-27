package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"

	"github.com/epalmerini/mtg-proxy/internal/app"
	"github.com/epalmerini/mtg-proxy/internal/card"
	"github.com/epalmerini/mtg-proxy/internal/halftone"
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
		if skipBasics && c.IsBasicLand() {
			fmt.Fprintf(os.Stderr, "skipping basic land: %s\n", c.Front().Name)
			continue
		}
		dc := card.DeckCard{Card: c, Quantity: entry.Quantity, IsCommander: entry.IsCommander}
		if entry.IsCommander && c.ArtCropURL != "" {
			art, err := fetchArtCrop(c.ArtCropURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not fetch art for %q: %v\n", entry.Name, err)
			} else {
				dc.ArtImage = halftone.Apply(art, 8)
			}
		}
		deckCards = append(deckCards, dc)
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

func fetchArtCrop(url string) (image.Image, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mtg-proxy/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}
	return img, nil
}
