package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s inputfile\n", os.Args[0])
		os.Exit(1)
	}

	inputFilename := os.Args[1]
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		panic(fmt.Sprintf("could not open input file: %s", err))
	}

	defer inputFile.Close()

	inputBytes, err := io.ReadAll(inputFile)
	if err != nil {
		panic(fmt.Sprintf("could not read input file: %s", err))
	}

	input := strings.TrimSpace(string(inputBytes))
	inputLines := strings.Split(input, "\n")
	histories, err := parseHistories(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse histories: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(histories))
}

func part1(histories [][]int) int {
	total := 0
	for _, history := range histories {
		total += predictNextValue(history)
	}

	return total
}

// predictNextValue calculates the next item in the sequence by using the nth differences
func predictNextValue(history []int) int {
	if allEqual(history, 0) {
		return 0
	}

	differences := make([]int, len(history)-1)
	for i := range history[1:] {
		differences[i] = history[i+1] - history[i]
	}

	return history[len(history)-1] + predictNextValue(differences)
}

func allEqual[T comparable, S ~[]T](values S, shouldEqual T) bool {
	for _, value := range values {
		if value != shouldEqual {
			return false
		}
	}

	return true
}

func parseHistories(lines []string) ([][]int, error) {
	histories := make([][]int, len(lines))
	for i, line := range lines {
		history, err := parseHistory(line)
		if err != nil {
			return nil, fmt.Errorf("parse history %d: %s", i, err)
		}

		histories[i] = history
	}

	return histories, nil
}

func parseHistory(line string) ([]int, error) {
	rawNums := strings.Split(line, " ")
	history := make([]int, len(rawNums))
	for i, rawNum := range rawNums {
		n, err := strconv.Atoi(rawNum)
		if err != nil {
			return nil, fmt.Errorf("parse item %d: %w", i, err)
		}

		history[i] = n
	}

	return history, nil
}
