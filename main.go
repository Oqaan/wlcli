package main

import (
	"fmt"
	"log"
	"os"
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

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: wlcli <stop name>")
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

	result, err := fetchDepartures(ids)
	if err != nil {
		log.Fatal(err)
	}

	t := table.New()
	t.Headers("LINE", "DIRECTION", "NEXT")
	t.Border(lipgloss.RoundedBorder())

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
		return strings.Compare(a.LineName, b.LineName)
	})
	for _, r := range tableRows {
		t.Row(r.LineName, r.Direction, strings.Join(r.Countdowns, ", "))
	}
	lipgloss.Println(t)
}
