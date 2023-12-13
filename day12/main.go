package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type SpringState int

const (
	SpringStateOperational SpringState = iota
	SpringStateDamaged
	SpringStateUnknown
)

type SpringStates []SpringState

type Record struct {
	states    SpringStates
	sequences []int
}

func (states SpringStates) Print() {
	for _, state := range states {
		switch state {
		case SpringStateDamaged:
			fmt.Print("#")
		case SpringStateOperational:
			fmt.Print(".")
		case SpringStateUnknown:
			fmt.Print("?")
		default:
			panic("invalid state")
		}
	}
	fmt.Println()
}

func (r Record) EvaluateUnknownStates() []SpringStates {
	res := []SpringStates{}
	for _, possibleStates := range getAllPossibleStates(r.states, 0) {
		if r.matchesSequence(possibleStates) {
			res = append(res, possibleStates)
		}
	}

	return res
}

func (r Record) matchesSequence(states SpringStates) bool {
	sequenceCountCursor := 0
	// states.Print()
	damageCount := 0
	needSpace := false
	for _, state := range states {
		if sequenceCountCursor > len(r.sequences)-1 && state == SpringStateDamaged {
			return false
		} else if sequenceCountCursor > len(r.sequences)-1 {
			continue
		}

		sequenceCount := r.sequences[sequenceCountCursor]
		if needSpace && state != SpringStateOperational {
			return false
		} else if needSpace {
			needSpace = false
			continue
		} else if state == SpringStateOperational && damageCount > 0 && damageCount < sequenceCount {
			return false
		}

		if state == SpringStateDamaged {
			damageCount++
		} else if state == SpringStateOperational {
			damageCount = 0
		}

		if damageCount == sequenceCount {
			sequenceCountCursor++
			damageCount = 0
			needSpace = true
		}
	}

	return sequenceCountCursor == len(r.sequences)
}

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

	records, err := parseRecords(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(records))
}

func part1(records []Record) int {
	total := 0
	for _, record := range records {
		possibilities := record.EvaluateUnknownStates()
		total += len(possibilities)
	}

	return total
}

func getAllPossibleStates(states SpringStates, trackingIdx int) []SpringStates {
	if trackingIdx > len(states)-1 {
		return []SpringStates{}
	} else if states[trackingIdx] != SpringStateUnknown {
		return getAllPossibleStates(states, trackingIdx+1)
	}

	ifOperational := slices.Clone(states)
	ifOperational[trackingIdx] = SpringStateOperational
	ifDamaged := slices.Clone(states)
	ifDamaged[trackingIdx] = SpringStateDamaged

	if slices.Index(states[trackingIdx+1:], SpringStateUnknown) == -1 {
		return []SpringStates{ifOperational, ifDamaged}
	}

	possibilities := []SpringStates{}
	possibilities = append(possibilities, getAllPossibleStates(ifOperational, trackingIdx+1)...)
	possibilities = append(possibilities, getAllPossibleStates(ifDamaged, trackingIdx+1)...)

	return possibilities
}

func parseRecords(inputLines []string) ([]Record, error) {
	return tryParse(inputLines, parseRecord)
}

func parseRecord(inputLine string) (Record, error) {
	lineRegexp := regexp.MustCompile(`^([#.?]+) ((?:\d+,)*\d+)$`)
	matches := lineRegexp.FindStringSubmatch(inputLine)
	if matches == nil {
		return Record{}, errors.New("malformed input line")
	}

	states, err := parseSpringStates(matches[1])
	if err != nil {
		return Record{}, fmt.Errorf("spring state: %w", err)
	}

	sequences, err := tryParse(strings.Split(matches[2], ","), strconv.Atoi)
	if err != nil {
		return Record{}, fmt.Errorf("sequences: %w", err)
	}

	return Record{states: states, sequences: sequences}, nil
}

func parseSpringStates(inputStates string) ([]SpringState, error) {
	states := make([]SpringState, len(inputStates))
	for i, char := range inputStates {
		switch char {
		case '.':
			states[i] = SpringStateOperational
		case '#':
			states[i] = SpringStateDamaged
		case '?':
			states[i] = SpringStateUnknown
		default:
			return nil, fmt.Errorf("invalid spring state char, %c", char)
		}
	}

	return states, nil
}

func tryParse[T any](items []string, parse func(string) (T, error)) ([]T, error) {
	res := make([]T, 0, len(items))
	for i, item := range items {
		parsed, err := parse(item)
		if err != nil {
			return nil, fmt.Errorf("invalid item #%d: %w", i+1, err)
		}

		res = append(res, parsed)
	}

	return res, nil
}
