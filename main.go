package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

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

	var showVersion bool
	var watch bool
	refreshInterval := 30 * time.Second // How often watch mode refreshes
	watchUsage := fmt.Sprintf("refresh every %s", refreshInterval)

	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.BoolVar(&showVersion, "v", false, "show version information")
	flag.BoolVar(&watch, "watch", false, watchUsage)
	flag.BoolVar(&watch, "w", false, watchUsage)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: wlcli [-w | --watch] <stop name>")
		fmt.Fprintln(os.Stderr, "       wlcli -v | --version")
	}
	flag.Parse()

	if showVersion {
		fmt.Println("wlcli", version())
		return
	}

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	rows, err := loadStops()
	if err != nil {
		log.Fatal(err)
	}
	// The shell splits the arguments, so we need to join them back together, e.g. "Kagraner Platz"
	name := strings.Join(flag.Args(), " ")
	ids := findStopIDs(rows, name)
	if len(ids) == 0 {
		log.Fatalf("no stop found: %q", name)
	}
	if len(ids) > maxStopIDs {
		log.Fatalf("%q matches %d stops (max %d), please be more specific.", name, len(ids), maxStopIDs)
	}

	if !watch {
		err = showDepartures(ids)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	for {
		// Clear the terminal before printing the new table
		fmt.Print("\033[H\033[2J")
		// in watch mode errors are only printed, so the next refresh can try again
		err := showDepartures(ids)
		if err != nil {
			log.Print(err)
		}
		fmt.Println("updated", time.Now().Format("15:04:05"), "· press Ctrl+C to quit")
		time.Sleep(refreshInterval)
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

	if len(tableRows) == 0 {
		fmt.Println("no departures at the moment")
	} else {
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
	}

	var lines []string
	for _, r := range tableRows {
		lines = append(lines, r.LineName)
	}

	infos, err := fetchTrafficInfos()
	if err != nil {
		log.Print(err)
		return nil
	}

	relevant := relevantTrafficInfos(infos, ids, lines)

	warning := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E5C07B"))
	description := lipgloss.NewStyle().Width(70).PaddingLeft(2)

	for _, info := range relevant {
		title := strings.ReplaceAll(info.Title, "\n", " ")
		lipgloss.Println(warning.Render("⚠ " + title))
		// many infos repeat the title as description, so only print if it is unique
		desc := strings.ReplaceAll(info.Description, "\n", " ")
		if desc != title {
			lipgloss.Println(description.Render(info.Description))
		}
	}

	return nil
}

// relevantTrafficInfos returns the traffic infos for the given stops or lines
func relevantTrafficInfos(infos []TrafficInfo, ids []string, lines []string) []TrafficInfo {
	var relevant []TrafficInfo

	for _, info := range infos {
		if len(info.RelatedStops) > 0 {
			for _, stop := range info.RelatedStops {
				if slices.Contains(ids, strconv.Itoa(stop)) {
					relevant = append(relevant, info)
					break
				}
			}
		} else {
			for _, line := range info.RelatedLines {
				if slices.Contains(lines, line) {
					relevant = append(relevant, info)
					break
				}
			}
		}
	}
	return relevant
}
