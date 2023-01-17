package main

import (
	"fmt"
	"log"
)

func main() {

	readEnvironment("kpg.env")

	loadHistory()

	current := retrieveCurrentData()

	var tendency string
	var icon rune
	diff, err := levelDifference(current, 240)
	if err != nil {
		tendency = "?"
		icon = 0
	} else {
		tendency, icon = getTendencyString(diff)
	}

	statusText := prepareStatusString(current, tendency, icon)

	doPost := checkIfPostNow(current)
	if doPost {
		id, err := postToMastodon(statusText)
		if err != nil {
			log.Printf("%v [%v: %v] - error: %v\n", current, diff, tendency, err)
		} else {
			fmt.Printf("%v [%v: %v] - %v\n", current, diff, tendency, id)
			savePostTime(current)
		}
	} else {
		fmt.Printf("%v [%v: %v] - no post\n", current, diff, tendency)
	}

	if len(history) == 0 || history[len(history)-1].Timestamp.Before(current.Timestamp) {
		history = append(history, current)
		saveHistory()
	}
}
