package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

type row struct {
	LineName   string
	Direction  string
	Countdowns []string
}

func main() {
	log.SetFlags(0) // Disable timestamps in log output
	log.SetPrefix("wlcli: ")

	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println("wlcli", version())
		return
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: wlcli <stop name>")
		fmt.Fprintln(os.Stderr, "       wlcli -v | --version")
		os.Exit(1)
	}

	rows, err := loadStops()
	if err != nil {
		log.Fatal(err)
	}
	// The shell splits the arguments, so we need to join them back together, e.g. "Kagraner Platz"
	name := strings.Join(os.Args[1:], " ")
	ids := findStopIDs(rows, name)
	if len(ids) == 0 {
		log.Fatalf("no stop found: %q", name)
	}
	if len(ids) > maxStopIDs {
		log.Fatalf("%q matches %d stops (max %d), please be more specific.", name, len(ids), maxStopIDs)
	}

	err = showDepartures(ids)
	if err != nil {
		log.Fatal(err)
	}
}

// lineColor returns the official Wiener Linien color for U-Bahn lines
func lineColor(lineName string) color.Color {
	switch lineName {
	case "U1":
		return lipgloss.Color("#E3000F")
	case "U2":
		return lipgloss.Color("#A862A4")
	case "U3":
		return lipgloss.Color("#EF7C00")
	case "U4":
		return lipgloss.Color("#00963F")
	case "U5":
		return lipgloss.Color("#008F95")
	case "U6":
		return lipgloss.Color("#9D6830")
	default:
		return lipgloss.NoColor{}
	}
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return info.Main.Version
}

// lineSortKey returns the group and number used to sort a line name
// Group 0 is U-Bahn, 1 is tram and bus, 2 is night lines
func lineSortKey(name string) (group int, number int) {
	group = 1
	rest := name

	if len(name) > 0 {
		switch name[0] {
		case 'U':
			group = 0
			rest = name[1:]
		case 'N':
			group = 2
			rest = name[1:]
		}
	}

	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}

	if end > 0 {
		number, _ = strconv.Atoi(rest[:end]) // rest[:end] contains only digits, so Atoi cannot fail
	}

	return group, number
}

// compareLines compares two line names for sorting: U-Bahn first,
// then trams and buses, then night lines, each by number
func compareLines(a, b string) int {
	groupA, numberA := lineSortKey(a)
	groupB, numberB := lineSortKey(b)

	if groupA != groupB {
		return groupA - groupB
	}

	if numberA != numberB {
		return numberA - numberB
	}

	return strings.Compare(a, b)
}

// showDepartures fetches the departures for the given stop IDs and prints them as a table
func showDepartures(ids []string) error {
	result, err := fetchDepartures(ids)
	if err != nil {
		return err
	}

	var tableRows []row

	for _, m := range result.Data.Monitors {
		for _, l := range m.Lines {
			var countdowns []string
			for _, d := range l.Departures.Departure {
				countdowns = append(countdowns, strconv.Itoa(d.DepartureTime.Countdown))
				if len(countdowns) >= 3 {
					break
				}
			}
			tableRows = append(tableRows, row{
				LineName:   l.Name,
				Direction:  strings.TrimSpace(l.Towards),
				Countdowns: countdowns,
			})
		}
	}
	slices.SortFunc(tableRows, func(a, b row) int {
		return compareLines(a.LineName, b.LineName)
	})

	t := table.New()
	t.Headers("LINE", "DIRECTION", "NEXT")
	t.Border(lipgloss.RoundedBorder())
	t.BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8")))
	t.StyleFunc(func(rowIndex, colIndex int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)

		if rowIndex == table.HeaderRow {
			return style.Bold(true)
		}
		if colIndex == 0 {
			name := tableRows[rowIndex].LineName
			return style.Bold(true).Foreground(lineColor(name))
		}
		return style
	})

	for _, r := range tableRows {
		t.Row(r.LineName, r.Direction, strings.Join(r.Countdowns, ", "))
	}
	lipgloss.Println(t)
	return nil
}
