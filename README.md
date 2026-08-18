# mtg-proxy

A CLI tool that generates printable PDF proxy cards for Magic: The Gathering. It parses a decklist, fetches card data from the [Scryfall API](https://scryfall.com/docs/api), and renders printable proxies onto A4 pages (3x3 grid, 63x88mm per card).

Commander cards include art by default. Use `-art-all` to include art on every card. Before embedding art, mtg-proxy converts it to a light-biased four-level grayscale palette and applies Floyd-Steinberg error-diffusion dithering. This keeps cards visually identifiable while limiting ink/toner coverage compared with printing the original full-color art. Double-faced cards use the corresponding art for each face.

## Requirements

- Go 1.26+

## Install

```bash
go install github.com/emiliopalmerini/mtg-proxy/cmd/mtg-proxy@latest
```

Or build from source:

```bash
git clone https://github.com/emiliopalmerini/mtg-proxy.git
cd mtg-proxy
go build ./cmd/mtg-proxy
```

## Usage

```bash
mtg-proxy -i decklist.txt -o proxies.pdf
```

To add dithered art to every card:

```bash
mtg-proxy -i decklist.txt -o proxies.pdf -art-all
```

### Flags

| Flag           | Default       | Description                              |
| -------------- | ------------- | ---------------------------------------- |
| `-i`           | _(required)_  | Path to decklist file                    |
| `-o`           | `proxies.pdf` | Path to output PDF                       |
| `-skip-basics` | `false`       | Exclude basic lands from the output      |
| `-art-all`     | `false`       | Include dithered art on every card       |

## Decklist formats

### Simple

```
4 Lightning Bolt
2 Counterspell
1 Tarmogoyf
```

Lines starting with `#` or `//` are treated as comments.

### Rich (with set code and collector number)

Fetches the exact printing from Scryfall:

```
1x Ankh of Mishra (6ed) 273 [Slug]
1x Archangel of Tithes (otj) 2 [Defense]
1x Sol Ring (ecc) 58 [Acceleration,Artifact]
1x Liesa, Shroud of Dusk (cmr) 286 [Commander{top}]
```

- `1x` quantity prefix (the `x` is optional)
- `(set)` set code in parentheses
- Collector number after the set code
- `[tags]` are optional and ignored by the parser
- `[Commander{top}]` marks the card as a commander and includes its art by default

## Output

The generated PDF uses A4 pages with a 3x3 grid of 63x88mm cards. Each card face includes:

- Card name and mana cost
- Optional dithered, ink-saving card art
- Type line
- Oracle text
- Power/toughness or loyalty

Cut along the card borders and sleeve them in front of a basic land for play-ready proxies.

## License

MIT
