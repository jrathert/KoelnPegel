package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func readEnvironment(fname string) {
	file, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	input := bufio.NewScanner(file)
	lineNo := 0
	for input.Scan() {
		lineNo++
		line := input.Text()
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		elems := strings.SplitN(line, "=", 2)
		if len(elems) != 2 {
			log.Printf("Syntax error reading environment file %v in line %v ('%v') - ignoring\n", fname, lineNo, line)
		} else {
			os.Setenv(elems[0], elems[1])
		}
	}
}
