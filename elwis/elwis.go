package elwis

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/net/html"
)

// calcTablePosition(22)
// calcTablePosition(23)
// calcTablePosition(0)
// calcTablePosition(1)
// fmt.Println("22: ", fetchPrognosis(22))
// fmt.Println("23: ", fetchPrognosis(23))
// fmt.Println("00: ", fetchPrognosis(0))
// fmt.Println("01: ", fetchPrognosis(1))

func calcTablePosition(hour int) (row int, col int) {

	col = 0
	row = (hour + 1) / 2

	if row > 11 {
		col = 1
		row = 0
	}

	// last: overall offset in table
	row += 2
	col += 5

	// row 2 has one additional (front, multi-row) column
	if row == 2 {
		col++
	}

	// fmt.Printf("row: %v, col: %v\n", row, col)
	return row, col
}

func FetchPrognosis(hour int) int {

	row, col := calcTablePosition(hour)

	url := "https://www.elwis.de/DE/dynamisch/gewaesserkunde/wasserstaende/index.php?target=1&pegelId=a6ee8177-107b-47dd-bcfd-30960ccc6e9c"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting data")
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading body")
		log.Fatal(err)
	}

	z := html.NewTokenizer(bytes.NewReader(body))
	rowcnt := 0
	colcnt := 0
	inRow := false
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			log.Fatal(z.Err())
			return -1
		case html.StartTagToken:
			tn, _ := z.TagName()
			if !inRow {
				if string(tn) == "tr" {
					if rowcnt == row {
						inRow = true
						// fmt.Println("In target row: ", rowcnt)
						colcnt = 0
					} else {
						// fmt.Println("Got table row ", rowcnt)
						rowcnt++
					}
				}
			} else {
				if string(tn) == "td" {
					if colcnt == col {
						// fmt.Println("In target row: ", rowcnt, " col: ", colcnt)
						tt = z.Next()
						if tt != html.TextToken {
							// seems we are in a <b> tag in first column
							_ = z.Next()
						}
						// if rowcnt != 2 {
						// 	// need to skip one <b>
						// 	_ = z.Next()
						// }
						// _ = z.Next()
						valstr := string(z.Text())
						val, err := strconv.Atoi(valstr)
						if err != nil {
							log.Fatal(err)
						}
						return val
					} else {
						// fmt.Println("Got table row ", rowcnt, " col ", colcnt)
						colcnt++
					}
				}
			}
		}
	}
}
