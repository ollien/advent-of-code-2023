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

func (states SpringStates) String() string {
	res := ""
	for _, state := range states {
		switch state {
		case SpringStateDamaged:
			res += "#"
		case SpringStateOperational:
			res += "."
		case SpringStateUnknown:
			res += "?"
		default:
			panic("invalid state")
		}
	}

	return res
}

func (r Record) EvaluateUnknownStates() []SpringStates {
	return r.generateStates(r.states, 0, 0)
}

func (r Record) generateStates(states SpringStates, stateIdx, sequenceIdx int) []SpringStates {
	if stateIdx > len(states)-1 && sequenceIdx > len(r.sequences)-1 {
		return []SpringStates{states}
	} else if stateIdx > len(states)-1 {
		return []SpringStates{}
	} else if sequenceIdx > len(r.sequences)-1 {
		if !anyEqual(states[stateIdx:], SpringStateDamaged) {
			return []SpringStates{states}
		} else {
			return []SpringStates{}
		}
	}

	currentSequenceCount := r.sequences[sequenceIdx]
	startIdx := stateIdx - currentSequenceCount + 1

	if states[stateIdx] == SpringStateUnknown {
		ifDamaged := r.exploreWithState(states, stateIdx, sequenceIdx, SpringStateDamaged)
		ifOperational := r.exploreWithState(states, stateIdx, sequenceIdx, SpringStateOperational)

		return append(ifDamaged, ifOperational...)
	} else if states[stateIdx] == SpringStateOperational {
		// If this is damaged, keep going so we can find the rest of the group
		return r.generateStates(states, stateIdx+1, sequenceIdx)
	} else if states[stateIdx] != SpringStateDamaged {
		panic(fmt.Sprintf("invalid state %d", states[stateIdx]))
	}

	if startIdx < 0 && stateIdx < len(states)-1 && states[stateIdx+1] == SpringStateDamaged {
		// Could be a match, we don't know yet
		return r.generateStates(states, stateIdx+1, sequenceIdx)
	} else if startIdx < 0 && stateIdx < len(states)-1 && states[stateIdx+1] == SpringStateUnknown {
		return r.exploreWithState(states, stateIdx+1, sequenceIdx, SpringStateDamaged)
	} else if startIdx < 0 {
		// Can't be a match anymore
		return []SpringStates{}
	}

	haveRightDamagedCount := allEqual(states[startIdx:stateIdx+1], SpringStateDamaged)
	if haveRightDamagedCount && !(startIdx == 0 || states[startIdx-1] == SpringStateOperational) {
		// If all the items in the span are damaged, but the span before that is damaged, this is a bad search
		return []SpringStates{}
	} else if haveRightDamagedCount && stateIdx == len(states)-1 && sequenceIdx == len(r.sequences)-1 {
		// If we have the right damage count, and we've hit the end, then we're done with our search
		return []SpringStates{states}
	} else if haveRightDamagedCount && stateIdx == len(states)-1 {
		// If we have the right damage count, and we've hit the end, but we still have sequences left,
		// we can't match
		return []SpringStates{}
	} else if haveRightDamagedCount && states[stateIdx+1] == SpringStateDamaged {
		// If the next one is damaged, we have to end our search here
		return []SpringStates{}
	} else if haveRightDamagedCount && states[stateIdx+1] == SpringStateUnknown {
		// If we hit an unknown, try to finish this having ended the sequence
		return r.exploreWithState(states, stateIdx+1, sequenceIdx+1, SpringStateOperational)
	} else if haveRightDamagedCount {
		// We've finished a match successfully, the next is known to be operational
		return r.generateStates(states, stateIdx+1, sequenceIdx+1)
	} else if stateIdx < len(states)-1 && states[stateIdx+1] == SpringStateOperational {
		// This doesn't match, and we've hit the end, so nothing else we can try
		return []SpringStates{}
	} else if stateIdx < len(states)-1 && states[stateIdx+1] == SpringStateUnknown {
		return r.exploreWithState(states, stateIdx+1, sequenceIdx, SpringStateDamaged)
	}

	return r.generateStates(states, stateIdx+1, sequenceIdx)

}

func (r Record) exploreWithState(states SpringStates, stateIdx, sequenceIdx int, withState SpringState) []SpringStates {
	updStates := slices.Clone(states)
	updStates[stateIdx] = withState
	return r.generateStates(updStates, stateIdx, sequenceIdx)
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
 
func allEqual[T comparable, S ~[]T](slice S, val T) bool {
	for _, item := range slice {
		if item != val {
			return false
		}
	}

	return true
}

func anyEqual[T comparable, S ~[]T](slice S, val T) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}

	return false
}
