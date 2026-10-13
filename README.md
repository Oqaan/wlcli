# wlcli

[![CI](https://github.com/Oqaan/wlcli/actions/workflows/ci.yml/badge.svg)](https://github.com/Oqaan/wlcli/actions/workflows/ci.yml)

Real-time Vienna public transport departures in your terminal, written in Go.

Type a stop name and `wlcli` shows the next departures of every line that stops there (U-Bahn, tram and bus), using the official Wiener Linien real-time API.

## Installation

Requires [Go](https://go.dev/dl/) 1.25 or newer.

```sh
go install github.com/Oqaan/wlcli@latest
```

If you get `command not found` afterwards, your terminal doesn't know where Go puts installed programs. Add Go's `bin` folder to your `PATH`:

```sh
# Linux / WSL (bash)
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc
source ~/.bashrc

# macOS (zsh)
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc
source ~/.zshrc
```

## Usage

```sh
wlcli <stop name>
```

Example:

```sh
wlcli schwedenplatz
```

```text
╭──────┬────────────────────────┬────────────╮
│ LINE │ DIRECTION              │ NEXT       │
├──────┼────────────────────────┼────────────┤
│ 1    │ Stefan-Fadinger-Platz  │ 0, 10, 16  │
│ 1    │ Prater Hauptallee      │ 0, 11, 23  │
│ 2    │ Friedrich-Engels-Platz │ 10, 20, 30 │
│ 2    │ Dornbach               │ 1, 13, 25  │
│ U1   │ Leopoldau              │ 3, 11, 19  │
│ U1   │ Oberlaa                │ 4, 13, 19  │
│ U4   │ Heiligenstadt          │ 1, 8, 15   │
│ U4   │ Hütteldorf             │ 6, 13, 20  │
╰──────┴────────────────────────┴────────────╯
```

`NEXT` shows the next three departures in minutes. In the terminal, U-Bahn lines are shown in their official Wiener Linien colors.

- The search is case-insensitive and matches partial names, so `wlcli kagraner platz` also finds the tram and bus stops named "Kagraner Platz U".
- Quotes around the stop name are optional: `wlcli Kagraner Platz` and `wlcli "Kagraner Platz"` do the same thing.
- Very generic searches (e.g. `platz`) match hundreds of stops. `wlcli` rejects searches with more than 100 matching stop IDs and asks you to be more specific.

## How it works

1. Downloads the [list of stops](https://www.wienerlinien.at/ogd_realtime/doku/ogd/wienerlinien-ogd-haltepunkte.csv) and finds all stop IDs whose name contains the search term.
2. Requests the real-time departures for these stop IDs from the `/monitor` endpoint.
3. Prints one row per line and direction, sorted by line, as a table rendered with [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Data source

Data source: [Wiener Linien](https://www.wienerlinien.at/open-data), licensed under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).

## License

The code is licensed under the [MIT License](LICENSE). This does not apply to the data, see [Data source](#data-source).
