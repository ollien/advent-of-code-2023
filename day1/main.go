package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

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
	fmt.Printf("Part 2: %d\n", part2(inputLines))
}

func part1(input []string) int {
	digits := []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	}

	result, err := solve(input, digits, strconv.Atoi)
	if err != nil {
		panic(err)
	}

	return result
}

func part2(input []string) int {
	digits := []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	}

	words := []string{
		"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
	}

	allPossible := make([]string, len(digits), len(digits)+len(words))
	copy(allPossible, digits)
	allPossible = append(allPossible, words...)

	result, err := solve(input, allPossible, func(s string) (int, error) {
		n, err := strconv.Atoi(s)
		if err == nil {
			return n, nil
		}

		idx := slices.Index(words, s)
		if idx != -1 {
			return idx, nil
		}

		return 0, errors.New("invalid digit")
	})

	if err != nil {
		panic(err)
	}

	return result
}

// filterToMatches find all overlapping matches of the stringset "valid" in "line"
func filterToMatches(line string, valid []string) []string {
	type match struct {
		idx   int
		value string
	}

	matches := []match{}
	start := 0
	// While we _could_ advance by the length of the string this
	// challenge allows for overlapping strings. We can be a bit
	// naive by just advancing one char and working that way
	for i := range line {
		for _, matchCandidate := range valid {
			if strings.HasPrefix(line[i:], matchCandidate) {
				matches = append(matches, match{idx: start, value: matchCandidate})
				break
			}
		}
	}

	slices.SortFunc(matches, func(a, b match) int {
		return a.idx - b.idx
	})

	output := []string{}
	for _, item := range matches {
		output = append(output, item.value)
	}

	return output
}

// smashToDigits will "smash" two digits together to form a two digit number
func smashDigits(digit1, digit2 int) int {
	return digit1*10 + digit2
}

// mapToDigits will convert a list of strings to an int, and return the aggregated results
// or return an error immediately if one is encountered. In other words, it is assumed
// every item in this list is a valid number according to the toDigit function
func mapToDigits(items []string, toDigit func(string) (int, error)) ([]int, error) {
	output := []int{}
	for _, item := range items {
		n, err := toDigit(item)
		if err != nil {
			return nil, fmt.Errorf("convert %s: %w", item, err)
		}

		output = append(output, n)
	}

	return output, nil
}

// getCoordinate gets a coordinate from the given calibration value
func getCoordinate(line string, validNumbers []string, convertToNumber func(string) (int, error)) (int, error) {
	candidates := filterToMatches(line, validNumbers)
	if len(candidates) == 0 {
		return 0, errors.New("invalid calibration value")
	}

	digits, err := mapToDigits(candidates, convertToNumber)
	if err != nil {
		// Programmer error, given the above filter
		panic(err)
	}

	if len(digits) == 1 {
		return smashDigits(digits[0], digits[0]), nil
	}

	value1 := digits[0]
	value2 := digits[len(candidates)-1]
	return smashDigits(value1, value2), nil
}

func solve(input []string, validNumbers []string, convert func(string) (int, error)) (int, error) {
	total := 0
	for _, line := range input {
		coordinate, err := getCoordinate(line, validNumbers, convert)
		if err != nil {
			return 0, fmt.Errorf("no valid coordinate on line '%s': %s", line, err)
		}

		total += coordinate
	}

	return total, nil
}
