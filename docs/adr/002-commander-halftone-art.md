# ADR-002: Halftone Art for Commander Cards

## Status
Accepted

## Context
Le proxy attuali sono puramente testuali. Per le carte commander (identificate dal tag `[Commander{...}]` nella decklist) vogliamo mostrare l'illustrazione della carta in stile halftone (mezzitoni), posizionata tra il nome e la riga del tipo. Questo rende il commander immediatamente riconoscibile senza consumare troppo inchiostro (~50% di copertura con grid size 8).

## Decision

### Parsing del tag Commander
- Il parser attualmente scarta i tag `[...]`. Deve invece riconoscere `Commander` (case-insensitive) tra i tag e propagare un flag `IsCommander` su `DeckEntry`.
- Il tag `{top}` o simili modificatori dentro `Commander{...}` vengono ignorati per ora (riguardano l'ordinamento, non l'art).

### Modello
- `DeckEntry`: aggiungere `IsCommander bool`
- `Card`: aggiungere `ArtCropURL string` (URL dell'art crop da Scryfall)
- `DeckCard`: nessuna modifica, porta già `Card` e si aggiunge `IsCommander bool`

### Scryfall Client
- La risposta JSON contiene `image_uris.art_crop` (URL diretto a un JPG). Parsare questo campo e popolarlo su `Card.ArtCropURL`.
- Per carte multi-faced, usare `card_faces[0].image_uris.art_crop` se `image_uris` top-level non esiste.

### Halftone Processing
- Nuovo package `internal/halftone` con una funzione:
  ```go
  func Apply(src image.Image, gridSize int) image.Image
  ```
- Algoritmo: griglia di celle `gridSize x gridSize` px, luminanza media per cella, cerchio nero proporzionale alla darkness. Grid size = 8.
- Input: JPEG dall'URL art_crop. Output: `image.Gray` (bianco/nero).

### Image Fetching
- Il fetch dell'immagine avviene nel main loop, solo per carte con `IsCommander == true` e `ArtCropURL != ""`.
- Usare `net/http` con User-Agent `mtg-proxy/1.0`, seguire redirect.
- L'immagine viene processata in memoria, non salvata su disco.

### PDF Rendering
- Per le carte commander, il layout cambia:
  ```
  ┌─────────────────────────┐
  │ Name              Mana  │  <- header (invariato)
  ├─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─┤
  │                         │
  │    [halftone art]       │  <- NUOVO: art crop in halftone
  │                         │
  ├─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─┤
  │ Type Line               │  <- type line (invariato)
  ├─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─┤
  │ Oracle text...          │  <- oracle text (meno spazio)
  │                         │
  │              P/T or Loy │  <- footer (invariato)
  └─────────────────────────┘
  ```
- L'immagine halftone viene registrata nel PDF con `fpdf.RegisterImageOptionsReader` e posizionata con `fpdf.ImageOptions`.
- Altezza art: ~30mm (su 88mm totali di card height), larghezza: `cardW - 2*padding`.
- L'oracle text occupa lo spazio rimanente tra type line e footer.

### Interfacce
- `Card` acquisisce `ArtCropURL` ma il campo immagine processata non va nel modello di dominio.
- Il renderer riceve l'immagine halftone come dato aggiuntivo. Opzioni:
  - `DeckCard` acquisisce un campo `ArtImage image.Image` (nullable, presente solo per commander)
  - Il renderer controlla `ArtImage != nil` per decidere il layout.

### Edge Cases
- Commander senza `image_uris` (es. card appena spoilerata): skip dell'art, render come carta normale.
- Commander multi-faced: usare l'art della prima faccia.
- Art crop con aspect ratio variabile: scalare per riempire la larghezza disponibile, centrare verticalmente, clippare se eccede l'altezza.

### Non in scope
- Art per carte non-commander
- Scelta del grid size da CLI
- Outline/edge detection (scartato: troppo rumoroso su art complesse)
- Floyd-Steinberg dithering (buona qualità ma ~66% ink, halftone-8 preferito a ~50%)

## Consequences
- Il tempo di generazione aumenta leggermente per il fetch dell'immagine del commander (1 richiesta HTTP extra)
- Il PDF sarà più pesante per le pagine con commander (immagine embedded)
- L'ink coverage per la carta commander sale a ~50% nella zona art, ma il resto del deck rimane text-only
