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

type SequenceStatus int

const (
	SequenceStatusDoesntMatch SequenceStatus = iota
	SequenceStatusMatches
	SequencesStatusCouldMatch
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
	return r.generateStates(r.states, 0, 0)
}

func (r Record) generateStates(states SpringStates, stateIdx int, sequenceIdx int) []SpringStates {
	if slices.Index(states, SpringStateUnknown) == -1 && matchesSequence(states, r.sequences) == SequenceStatusMatches {
		return []SpringStates{states}
	} else if stateIdx > len(r.states)-1 {
		return []SpringStates{}
	} else if sequenceIdx == len(r.sequences) {
		// Try the same thing, but with all the remaining items as operational (as this can still be a match)
		allOperational := slices.Clone(states)
		for i, state := range allOperational {
			if state == SpringStateUnknown {
				allOperational[i] = SpringStateOperational
			}
		}

		if matchesSequence(allOperational, r.sequences) == SequenceStatusMatches {
			return []SpringStates{allOperational}
		} else {
			return []SpringStates{}
		}
	} else if matchesSequence(states[:stateIdx+1], r.sequences[:sequenceIdx+1]) == SequenceStatusMatches {
		return r.generateStates(states, stateIdx+1, sequenceIdx+1)
	} else if r.states[stateIdx] != SpringStateUnknown {
		return r.generateStates(states, stateIdx+1, sequenceIdx)
	}

	generatedStates := []SpringStates{}
	ifOperational := slices.Clone(states)
	ifOperational[stateIdx] = SpringStateOperational
	operationalMatches := matchesSequence(ifOperational[:stateIdx+1], r.sequences[:sequenceIdx+1])
	if operationalMatches == SequenceStatusMatches {
		generatedStates = append(generatedStates, r.generateStates(ifOperational, stateIdx+1, sequenceIdx+1)...)
	} else if operationalMatches == SequencesStatusCouldMatch {
		generatedStates = append(generatedStates, r.generateStates(ifOperational, stateIdx+1, sequenceIdx)...)
	}

	ifDamaged := slices.Clone(states)
	ifDamaged[stateIdx] = SpringStateDamaged
	damagedMatches := matchesSequence(ifDamaged[:stateIdx+1], r.sequences[:sequenceIdx+1])
	if damagedMatches == SequenceStatusMatches {
		generatedStates = append(generatedStates, r.generateStates(ifDamaged, stateIdx+1, sequenceIdx+1)...)
	} else if damagedMatches == SequencesStatusCouldMatch {
		generatedStates = append(generatedStates, r.generateStates(ifDamaged, stateIdx+1, sequenceIdx)...)
	}

	return generatedStates
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

func matchesSequence(states SpringStates, sequences []int) SequenceStatus {
	sequenceCountCursor := 0
	damageCount := 0
	needSpace := false
	for _, state := range states {
		if sequenceCountCursor > len(sequences)-1 && state == SpringStateDamaged {
			return SequenceStatusDoesntMatch
		} else if sequenceCountCursor > len(sequences)-1 {
			continue
		}

		sequenceCount := sequences[sequenceCountCursor]
		if needSpace && state != SpringStateOperational {
			return SequenceStatusDoesntMatch
		} else if needSpace {
			needSpace = false
			continue
		} else if state == SpringStateOperational && damageCount > 0 && damageCount < sequenceCount {
			return SequenceStatusDoesntMatch
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

	if sequenceCountCursor < len(sequences) && damageCount < sequences[sequenceCountCursor] {
		return SequencesStatusCouldMatch
	} else if sequenceCountCursor == len(sequences) {
		return SequenceStatusMatches
	} else {
		return SequenceStatusDoesntMatch
	}
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
