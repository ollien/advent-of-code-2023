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

func filterChars(line string, criteria func(rune) bool) string {
	validChars := []rune{}
	for _, char := range line {
		if criteria(char) {
			validChars = append(validChars, char)
		}
	}

	return string(validChars)
}

func getCoordinate(line string) (int, error) {
	candidates := filterChars(line, unicode.IsDigit)
	if len(candidates) == 0 {
		return 0, errors.New("invalid calibration value")
	}

	if len(candidates) == 1 {
		value, err := strconv.Atoi(candidates)
		if err != nil {
			// Programmer error, given the above filter
			panic("invalid numeric string " + candidates)
		}

		return value*10 + value, nil
	}

	value1 := int(candidates[0] - '0')
	value2 := int(candidates[len(candidates)-1] - '0')
	return value1*10 + value2, nil
}

func part1(input []string) int {
	total := 0
	for _, line := range input {
		coordinate, err := getCoordinate(line)
		if err != nil {
			panic(fmt.Sprintf("No valid coordinate on line '%s': %s\n", line, err))
		}

		total += coordinate
	}

	return total
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

	defer inputFile.Close()

	inputFileBytes, err := io.ReadAll(inputFile)
	if err != nil {
		panic(fmt.Sprintf("failed to read input file: %s", err))
	}

	input := string(inputFileBytes)
	inputLines := strings.Split(strings.TrimSpace(input), "\n")

	fmt.Printf("Part 1: %d\n", part1(inputLines))
}
