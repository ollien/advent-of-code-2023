package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Coordinate struct {
	row int
	col int
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s inputfile\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	inputFile, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("failed to open input file: %s", err))
	}

	inputBytes, err := io.ReadAll(inputFile)
	if err != nil {
		panic(fmt.Sprintf("failed to read input file: %s", err))
	}

	input := string(inputBytes)
	inputLines := strings.Split(strings.TrimSpace(input), "\n")

	fmt.Printf("Part 1: %d\n", part1(inputLines))
}

func part1(inputLines []string) int {
	symbolPositions := findSymbols(inputLines)
	partNumberCandidates := []Coordinate{}
	for _, symbolPos := range symbolPositions {
		candidates := findNumbersAdjacentTo(inputLines, symbolPos)
		partNumberCandidates = append(partNumberCandidates, candidates...)
	}

	total := 0
	scanned := map[Coordinate]struct{}{}
	for _, candidate := range partNumberCandidates {
		if _, ok := scanned[candidate]; ok {
			continue
		}

		line := inputLines[candidate.row]

		backwardsBuf := strings.Builder{}
		// Scan backward
		for i := candidate.col - 1; i >= 0; i-- {
			char := line[i]
			if !unicode.IsDigit(rune(char)) {
				break
			}

			scanned[Coordinate{row: candidate.row, col: i}] = struct{}{}
			backwardsBuf.WriteByte(char)
		}

		forwardsBuf := strings.Builder{}
		// Scan forward
		for i := candidate.col; i < len(line); i++ {
			char := line[i]
			if !unicode.IsDigit(rune(char)) {
				break
			}

			scanned[Coordinate{row: candidate.row, col: i}] = struct{}{}
			forwardsBuf.WriteByte(char)
		}

		fullNumber := reverseString(backwardsBuf.String()) + forwardsBuf.String()
		scannedNumber, err := strconv.Atoi(fullNumber)
		if err != nil {
			panic(fmt.Sprintf("%s was not a number", fullNumber))
		}

		total += scannedNumber
	}

	return total
}

func findSymbols(inputLines []string) []Coordinate {
	coords := []Coordinate{}
	for row, line := range inputLines {
		for col, char := range line {
			if isSymbol(char) {
				coord := Coordinate{
					row: row,
					col: col,
				}

				coords = append(coords, coord)
			}
		}
	}

	return coords
}

func findNumbersAdjacentTo(inputLines []string, coordinate Coordinate) []Coordinate {
	coords := []Coordinate{}
	for dRow := -1; dRow <= 1; dRow++ {
		for dCol := -1; dCol <= 1; dCol++ {
			if dRow == 0 && dCol == 0 {
				continue
			}

			row := coordinate.row + dRow
			col := coordinate.col + dCol
			if unicode.IsDigit(rune(inputLines[row][col])) {
				coord := Coordinate{
					row: row,
					col: col,
				}

				coords = append(coords, coord)
			}
		}
	}

	return coords
}

func isSymbol(r rune) bool {
	return r != '.' && !unicode.IsDigit(r)
}

func reverseString(s string) string {
	buffer := strings.Builder{}
	for i := len(s) - 1; i >= 0; i-- {
		buffer.WriteByte(s[i])
	}

	return buffer.String()
}
