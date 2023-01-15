// This file implements a history of measurements
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// History of measurements
var history []Measurement

// max number of history elements that are stored
const HISTORY_LENGTH = 24 * 4

// name of the file to store history in
const HISTORY_FILE = ".kpg_history"

// Measurement represents in compact form one measurement of the Cologne level
// It has a timestamp, a water level and a water temperature
type Measurement struct {
	Timestamp   time.Time
	Level       float64
	Temperature float64
}

// String representation of a history item
func (m Measurement) String() string {
	return fmt.Sprintf("%v: %v cm (%v °C)", m.Timestamp, m.Level, m.Temperature)
}

// Load history from file into global variable
// The whole file is loaded, expecting it was written with the corresponding
// saveHistory function, that truncates numer of items appropriately
// Potential errors include reading data from a file and decoding the JSON
func loadHistory() error {
	input, err := os.ReadFile(HISTORY_FILE)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error reading history file:", err)
			return err
		}
	} else {
		if err := json.NewDecoder(bytes.NewReader(input)).Decode(&history); err != nil {
			fmt.Println("Error decoding history:", err)
			return err
		}
	}
	return nil
}

// Saves history to file, trunctating unnecessary elements
// Potential errors include encoding da data to JSON and writing to the file
func saveHistory() error {

	// shorten history to max HISTORY_LENGTH entries
	cap := len(history) - HISTORY_LENGTH
	if cap > 0 {
		history = history[cap:]
	}

	data, err := json.Marshal(history)
	if err != nil {
		fmt.Println("Error encoding history:", err)
		return err
	}
	err = os.WriteFile(HISTORY_FILE, data, 0664)
	if err != nil {
		fmt.Println("Error writing history file:", err)
		return err
	}
	return nil
}

func levelDifference(current Measurement, minutes int) (float64, error) {

	searchTime := current.Timestamp.Add(-time.Minute * time.Duration(minutes))

	matchIndex := -1
	for i := len(history) - 1; i >= 0; i-- {
		ts := history[i].Timestamp
		if searchTime.Before(ts) {
			matchIndex = i
		}
	}

	if matchIndex == -1 {
		errMsg := fmt.Sprintf("Error - cannot find %v in history", searchTime)
		return 0.0, errors.New(errMsg)
	} else {
		diff := current.Level - history[matchIndex].Level
		return diff, nil
	}
}
