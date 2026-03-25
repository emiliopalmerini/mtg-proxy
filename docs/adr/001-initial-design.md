# ADR-001: MTG Proxy Card Generator

## Status
Proposed

## Context
Serve un tool CLI per generare proxy di carte Magic: The Gathering ottimizzate per la stampa, senza immagini, partendo da una decklist.

## Decision

### Input
- File di testo con formato standard: `<quantity> <card name>`, una riga per carta
- Esempio:
  ```
  4 Lightning Bolt
  2 Counterspell
  1 Black Lotus
  ```
- Righe vuote e commenti (`//` o `#`) vengono ignorati

### Data Source
- Scryfall API (`https://api.scryfall.com/cards/named?exact=<name>`)
- Rate limit: max 10 req/s (rispettare le policy di Scryfall, che chiedono 50-100ms tra richieste)
- I dati usati per ogni carta:
  - `name` — nome
  - `mana_cost` — costo di mana (es. `{2}{U}{U}`)
  - `type_line` — riga del tipo
  - `oracle_text` — testo delle abilità
  - `power`, `toughness` — forza/costituzione (solo creature)
  - `loyalty` — lealtà (solo planeswalker)

### Mana Cost Rendering
- Simboli semplificati a lettere: `{W}` → `W`, `{U}` → `U`, `{B}` → `B`, `{R}` → `R`, `{G}` → `G`
- Mana generico: `{2}` → `2`, `{X}` → `X`
- Mana ibrido: `{W/U}` → `W/U`
- Esempio: `{1}{W}{W}` → `1WW`

### Output
- PDF A4 (210x297mm)
- 9 carte per pagina (griglia 3x3)
- Ogni carta: 63x88mm (dimensione standard MTG)
- Margini pagina per centrare la griglia
- Bordo nero sottile attorno a ogni carta
- Layout carta (dall'alto al basso):
  1. **Header**: Nome (sinistra) + Mana cost (destra)
  2. **Tipo**: type_line
  3. **Testo**: oracle_text (con word wrap)
  4. **Footer**: Power/Toughness (creature) o Loyalty (planeswalker), angolo basso destra
- Font: monospace o sans-serif, leggibile anche in piccolo

### CLI Interface
```
mtg-proxy generate -i decklist.txt -o proxies.pdf
```
- `-i` / `--input`: path del file decklist (required)
- `-o` / `--output`: path del PDF output (default: `proxies.pdf`)

### Error Handling
- Carta non trovata su Scryfall: warning a stderr, skip della carta, continua con le altre
- File input non trovato: errore fatale
- Errore di rete: retry con backoff (max 3 tentativi), poi warning e skip

## Architecture

```
cmd/mtg-proxy/main.go          — entrypoint, CLI parsing
internal/decklist/parser.go     — parsing del file decklist
internal/scryfall/client.go     — client HTTP per Scryfall API
internal/card/model.go          — domain model della carta
internal/pdf/renderer.go        — generazione PDF
```

- Hexagonal: il domain (`card`) definisce le interfacce, gli adapter (`scryfall`, `pdf`, `decklist`) le implementano
- Il parser restituisce `[]DeckEntry{Name string, Quantity int}`
- Il client Scryfall implementa un'interfaccia `CardFetcher` definita nel domain
- Il renderer PDF implementa un'interfaccia `DeckRenderer` definita nel domain

## Consequences
- Nessuna immagine = PDF leggero e veloce da generare
- Dipendenza da Scryfall API (ma è gratuita e stabile)
- Estendibile a formati diversi di decklist in futuro
