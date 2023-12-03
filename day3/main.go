package main

import (
	"errors"
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
	fmt.Printf("Part 2: %d\n", part2(inputLines))
}

func part1(inputLines []string) int {
	symbolPositions := findParts(inputLines, isSymbol)
	partNumberCandidates := []Coordinate{}
	for _, symbolPos := range symbolPositions {
		candidates := findNumbersAdjacentTo(inputLines, symbolPos)
		partNumberCandidates = append(partNumberCandidates, candidates...)
	}

	total := 0
	// keep track of digits we've already scanned; it's possible one number
	// is adjacent to two parts, and we don't wanna double count
	scanned := map[Coordinate]struct{}{}
	for _, candidate := range partNumberCandidates {
		if _, ok := scanned[candidate]; ok {
			continue
		}

		scannedNumber, scannedDigits, err := scanPartNumber(inputLines, candidate)
		if err != nil {
			panic(fmt.Sprintf("candidate was invalid: %s", err))
		}

		for _, scannedDigit := range scannedDigits {
			scanned[scannedDigit] = struct{}{}
		}

		total += scannedNumber
	}

	return total
}

func part2(inputLines []string) int {
	gearPositions := findParts(inputLines, isGear)
	totalRatio := 0
	for _, gearPos := range gearPositions {
		partNumbers, err := scanForGearPartNumbers(inputLines, gearPos)
		if err != nil {
			panic(fmt.Sprintf("gear scan failed: %s", err))
		}

		if len(partNumbers) == 2 {
			totalRatio += partNumbers[0] * partNumbers[1]
		}
	}

	return totalRatio
}

// findParts will find all parts in the input, using the given isPart function to tell
// if a character is a valid part.
func findParts(inputLines []string, isPart func(rune) bool) []Coordinate {
	coords := []Coordinate{}
	for row, line := range inputLines {
		for col, char := range line {
			if isPart(char) {
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

// findNumbersAdjacentTo will find all the digit characters adjacent to a given coordinate
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

// scanPartNumber takes an initial digit character and attempts to complete it by scanning left and right from that
// position. Returns the found part number and the positions scanned.
func scanPartNumber(inputLines []string, knownDigitPosition Coordinate) (int, []Coordinate, error) {
	if !unicode.IsDigit(rune(inputLines[knownDigitPosition.row][knownDigitPosition.col])) {
		return 0, nil, errors.New("position was not a numeric char")
	}

	line := inputLines[knownDigitPosition.row]

	scanned := []Coordinate{}
	backwardsBuf := strings.Builder{}
	// Scan backward
	for i := knownDigitPosition.col - 1; i >= 0; i-- {
		char := line[i]
		if !unicode.IsDigit(rune(char)) {
			break
		}

		scanned = append(scanned, Coordinate{row: knownDigitPosition.row, col: i})
		backwardsBuf.WriteByte(char)
	}

	forwardsBuf := strings.Builder{}
	// Scan forward
	for i := knownDigitPosition.col; i < len(line); i++ {
		char := line[i]
		if !unicode.IsDigit(rune(char)) {
			break
		}

		scanned = append(scanned, Coordinate{row: knownDigitPosition.row, col: i})
		forwardsBuf.WriteByte(char)
	}

	fullNumber := reverseString(backwardsBuf.String()) + forwardsBuf.String()
	scannedNumber, err := strconv.Atoi(fullNumber)
	if err != nil {
		panic(fmt.Sprintf("%s was not a number", fullNumber))
	}

	return scannedNumber, scanned, nil
}

// scanForGearPartNumbers will find all part numbers adjacent to the given gear character position
func scanForGearPartNumbers(inputLines []string, gearPosition Coordinate) ([]int, error) {
	if !isGear(rune(inputLines[gearPosition.row][gearPosition.col])) {
		return nil, errors.New("position is not gear")
	}

	candidates := findNumbersAdjacentTo(inputLines, gearPosition)
	// Keep track of the digits scanned, as it is possible for a single part number to be adjacent
	// to the gear in multiple places
	scanned := map[Coordinate]struct{}{}

	adjacentPartNumbers := []int{}
	for _, candidate := range candidates {
		if _, ok := scanned[candidate]; ok {
			continue
		}

		scannedNumber, scannedDigits, err := scanPartNumber(inputLines, candidate)
		if err != nil {
			// Strictly a programmer error, as this would imply that findNumbersAdjacentTo
			// didn't find us a number
			panic(fmt.Sprintf("candidate was invalid: %s", err))
		}

		for _, scannedDigit := range scannedDigits {
			scanned[scannedDigit] = struct{}{}
		}

		adjacentPartNumbers = append(adjacentPartNumbers, scannedNumber)
	}

	return adjacentPartNumbers, nil
}

func isSymbol(r rune) bool {
	return r != '.' && !unicode.IsDigit(r)
}

func isGear(r rune) bool {
	return r == '*'
}

func reverseString(s string) string {
	buffer := strings.Builder{}
	for i := len(s) - 1; i >= 0; i-- {
		buffer.WriteByte(s[i])
	}

	return buffer.String()
}
