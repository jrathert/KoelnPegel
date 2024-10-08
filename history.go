// This file implements a history of measurements
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// History of measurements
var history []Measurement

// max number of history elements that are stored
const HISTORY_LENGTH = 24 * 4

// name of the file to store history in
const HISTORY_FILE = ".kpg_history"

// Load history from file into global variable
// The whole file is loaded, expecting it was written with the corresponding
// saveHistory function, that truncates numer of items appropriately
// Potential errors include reading data from a file and decoding the JSON
func loadHistory() error {
	input, err := os.ReadFile(HISTORY_FILE)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("Error reading history file:", err)
			return err
		}
	} else {
		if err := json.NewDecoder(bytes.NewReader(input)).Decode(&history); err != nil {
			if err != io.EOF {
				log.Println("Error decoding history:", err)
				return err
			}
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

	data, err := json.MarshalIndent(history, "", "    ")
	if err != nil {
		log.Println("Error encoding history:", err)
		return err
	}
	err = os.WriteFile(HISTORY_FILE, data, 0664)
	if err != nil {
		log.Println("Error writing history file:", err)
		return err
	}
	return nil
}

//
// These two functions were planned to improve reporting, but are not in use
// yet - therefore commented them out for now to avoid any warnings
//
// func hasJustCrossedThreshold(m Measurement, threshold float64) bool {
// 	if m.Level <= threshold {
// 		return false
// 	}
// 	if len(history) == 0 {
// 		return true
// 	} else {
// 		return history[len(history)-1].Level < threshold
// 	}

// }

// func getDiffToLast(m Measurement) float64 {
// 	if len(history) == 0 {
// 		loadHistory()
// 	}
// 	if len(history) == 0 {
// 		return m.Level
// 	} else {
// 		return m.Level - history[len(history)-1].Level
// 	}

// }

func levelDifference(current Measurement, minutes int) (float64, error) {

	if len(history) == 0 {
		errMsg := "cannot determine difference - history is empty"
		return 0.0, errors.New(errMsg)
	}

	searchTime := current.Timestamp.Add(-time.Minute * time.Duration(minutes))

	matchIndex := -1
	// history is sorted by timestamps - identify the biggest index for which
	// the timestamp is maximum, but not bigger than searchtime
	for i := 0; i < len(history); i++ {
		ts := history[i].Timestamp
		if !searchTime.Before(ts) {
			matchIndex = i
		}
	}

	if matchIndex == -1 {
		// all entries in history have a timestamp bigger than searchtime
		// we take the first elem
		errMsg := fmt.Sprintf("cannot determine difference - all elems in history %v", searchTime)
		return 0.0, errors.New(errMsg)
	} else {
		diff := current.Level - history[matchIndex].Level
		return diff, nil
	}
}
