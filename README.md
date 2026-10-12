# wlcli

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
LINE   DIRECTION                NEXT
1      Stefan-Fadinger-Platz    2, 11, 21
1      Prater Hauptallee        2, 13, 24
2      Friedrich-Engels-Platz   0, 11, 21
2      Dornbach                 3, 13, 23
U1     Leopoldau                2, 7, 13
U1     Oberlaa                  1, 4, 10
U4     Heiligenstadt            3, 9, 14
U4     Hütteldorf               1, 6, 14
```

`NEXT` shows the next three departures in minutes.

- The search is case-insensitive and matches partial names, so `wlcli kagraner platz` also finds the tram and bus stops named "Kagraner Platz U".
- Quotes around the stop name are optional: `wlcli Kagraner Platz` and `wlcli "Kagraner Platz"` do the same thing.
- Very generic searches (e.g. `platz`) match hundreds of stops. `wlcli` rejects searches with more than 100 matching stop IDs and asks you to be more specific.

## How it works

1. Downloads the [list of stops](https://www.wienerlinien.at/ogd_realtime/doku/ogd/wienerlinien-ogd-haltepunkte.csv) and finds all stop IDs whose name contains the search term.
2. Requests the real-time departures for these stop IDs from the `/monitor` endpoint.
3. Prints one row per line and direction, sorted by line.

## Data source

Data source: [Wiener Linien](https://www.wienerlinien.at/open-data), licensed under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).

## License

The code is licensed under the [MIT License](LICENSE). This does not apply to the data, see [Data source](#data-source).
