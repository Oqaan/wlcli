package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type MonitorResponse struct {
	Data    Data    `json:"data"`
	Message Message `json:"message"`
}

type Message struct {
	Value string `json:"value"`
}

type Data struct {
	Monitors []Monitor `json:"monitors"`
}

type Monitor struct {
	Lines []Line `json:"lines"`
}

type Line struct {
	Name       string     `json:"name"`
	Towards    string     `json:"towards"`
	Departures Departures `json:"departures"`
}

// Departures is an extra wrapper object in the API response around the departure list
type Departures struct {
	Departure []Departure `json:"departure"`
}

type Departure struct {
	DepartureTime DepartureTime `json:"departureTime"`
}

type DepartureTime struct {
	Countdown int `json:"countdown"`
}

type row struct {
	LineName   string
	Direction  string
	Countdowns []string
}

// client is a shared HTTP client with a timeout to avoid hanging requests
var client = &http.Client{Timeout: 10 * time.Second}

// maxStopIDs limits how many stop IDs are sent in one request, because the API rejects requests with too many
const maxStopIDs = 100

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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "LINE\tDIRECTION\tNEXT\n")

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
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.LineName, r.Direction, strings.Join(r.Countdowns, ", "))
	}
	w.Flush()
}

// loadStops downloads the stops CSV and returns all rows
func loadStops() ([][]string, error) {
	resp, err := client.Get("https://www.wienerlinien.at/ogd_realtime/doku/ogd/wienerlinien-ogd-haltepunkte.csv")
	if err != nil {
		return nil, fmt.Errorf("downloading stops: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading stops: unexpected status %s", resp.Status)
	}
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';' // The Wiener Linien CSV uses semicolons instead of commas
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading stops CSV: %w", err)
	}
	return rows, nil
}

// fetchDepartures requests the real-time departures for the given stop IDs
func fetchDepartures(ids []string) (MonitorResponse, error) {
	stopIDs := strings.Join(ids, "&stopId=")
	apiURL := "https://www.wienerlinien.at/ogd_realtime/monitor?stopId=" + stopIDs

	resp, err := client.Get(apiURL)
	if err != nil {
		return MonitorResponse{}, fmt.Errorf("fetching departures: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MonitorResponse{}, fmt.Errorf("reading departure response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var apiResult MonitorResponse
		_ = json.Unmarshal(body, &apiResult) // Try to get the API's error message; the body may not be JSON (e.g. an HTML error page), so ignore errors.
		if apiResult.Message.Value != "" {
			return MonitorResponse{}, fmt.Errorf("fetching departures: %s: %s", resp.Status, apiResult.Message.Value)
		}
		return MonitorResponse{}, fmt.Errorf("fetching departures: unexpected status %s", resp.Status)
	}
	var result MonitorResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return MonitorResponse{}, fmt.Errorf("unmarshaling departure response: %w", err)
	}
	return result, nil
}

// findStopIDs returns the stop IDs of all rows whose name contains name (case-insensitive)
func findStopIDs(rows [][]string, name string) []string {
	var ids []string
	search := strings.ToLower(name)
	for _, record := range rows {
		// CSV columns: 0 = StopID, 2 = StopText (name)
		if strings.Contains(strings.ToLower(record[2]), search) {
			ids = append(ids, record[0])
		}
	}
	return ids
}
