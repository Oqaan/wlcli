package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type MonitorResponse struct {
	Data Data `json:"data"`
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

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: wlcli <stop name>")
	}

	rows, err := loadStops()
	if err != nil {
		log.Fatal(err)
	}
	// The shell splits the arguments, so we need to join them back together, e.g. "Kagraner Platz"
	name := strings.Join(os.Args[1:], " ")
	ids := findStopIDs(rows, name)
	if len(ids) == 0 {
		log.Fatalf("no stop found: %s", name)
	}

	stopIDs := strings.Join(ids, "&stopId=")
	apiURL := "https://www.wienerlinien.at/ogd_realtime/monitor?stopId=" + stopIDs

	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, body)
	}
	var result MonitorResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range result.Data.Monitors {
		for _, l := range m.Lines {
			for _, d := range l.Departures.Departure {
				fmt.Printf("%s %s %d\n", l.Name, l.Towards, d.DepartureTime.Countdown)
			}
		}
	}
}

// loadStops downloads the stops CSV and returns all rows
func loadStops() ([][]string, error) {
	resp, err := http.Get("https://www.wienerlinien.at/ogd_realtime/doku/ogd/wienerlinien-ogd-haltepunkte.csv")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';' // The Wiener Linien CSV uses semicolons instead of commas
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// findStopIDs returns the stop IDs of all rows whose name matches name
func findStopIDs(rows [][]string, name string) []string {
	var ids []string
	for _, row := range rows {
		// CSV columns: 0 = StopID, 2 = StopText (name)
		if row[2] == name {
			ids = append(ids, row[0])
		}
	}
	return ids
}
