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
		log.Fatal("usage: wlcli <stopId> [<stopId> ...]")
	}

	rows, err := loadStops()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("stops loaded:", len(rows))

	stopIDs := strings.Join(os.Args[1:], "&stopId=")
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
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return rows, nil
}
