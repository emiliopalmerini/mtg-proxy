# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

mtg-proxy is a Go CLI tool that generates printable PDF proxy cards for Magic: The Gathering. It parses a decklist file, fetches card data from the Scryfall API, and renders text-based proxy cards onto A4 pages (3x3 grid, 63x88mm per card).

## Commands

```bash
# Build
go build ./cmd/mtg-proxy

# Run
go run ./cmd/mtg-proxy -i decklist.txt -o proxies.pdf
go run ./cmd/mtg-proxy -i decklist.txt -skip-basics

# Test
go test ./...                              # all tests
go test ./internal/card/...                # single package
go test ./test/acceptance/...              # acceptance tests (use mock Scryfall server)
go test -run TestGenerateFullPipeline ./test/acceptance/...  # single test
```

## Architecture

Hexagonal architecture — domain types define ports (interfaces) in `internal/card/model.go`, adapters implement them:

- **`internal/card`** — Domain model. Defines `Card`, `CardFace`, `DeckEntry`, `DeckCard`, and value types (`ManaCost`, `TypeLine`, etc.). Declares three port interfaces: `DecklistParser`, `CardFetcher`, `DeckRenderer`.
- **`internal/decklist`** — Adapter implementing `DecklistParser`. Parses two formats: simple (`4 Lightning Bolt`) and rich with set/collector (`1 Ankh of Mishra (6ed) 273`).
- **`internal/scryfall`** — Adapter implementing `CardFetcher`. HTTP client for Scryfall API with 100ms rate limiting. Routes to `/cards/{set}/{number}` when set+collector available, else `/cards/named?exact=`.
- **`internal/pdf`** — Adapter implementing `DeckRenderer`. Uses go-pdf/fpdf with embedded DejaVu fonts. Renders card faces with name, mana cost, type line, oracle text, and stats/loyalty.
- **`internal/app`** — Composition root. Factory functions that wire adapters to ports — keeps `main.go` decoupled from adapter packages.
- **`cmd/mtg-proxy`** — CLI entry point. Orchestrates the parse→fetch→render pipeline.

## Testing

- **Unit tests** live alongside their packages (`internal/*/..._test.go`)
- **Acceptance tests** in `test/acceptance/` run the full pipeline against a mock HTTP server (no real Scryfall calls)
- Test fixtures in `test/testdata/`

## ADRs

New features should start with an ADR in `docs/adr/` before implementation.
