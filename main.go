package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	resp, err := http.Get("https://www.wienerlinien.at/ogd_realtime/monitor?stopId=4116")
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
	fmt.Println(len(result.Data.Monitors))
}
