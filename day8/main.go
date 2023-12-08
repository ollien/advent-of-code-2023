package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type Direction int
type NodeAddress string

const (
	DirectionLeft Direction = iota
	DirectionRight
)

const (
	NodeAddressStart NodeAddress = "AAA"
	NodeAddressEnd   NodeAddress = "ZZZ"
)

type NodeChoice struct {
	left  NodeAddress
	right NodeAddress
}

// TakeDirection will take the node in the given direction. Panics if an invalid direction is given
func (choice NodeChoice) TakeDirection(direction Direction) NodeAddress {
	switch direction {
	case DirectionLeft:
		return choice.left
	case DirectionRight:
		return choice.right
	default:
		panic("invalid direction value")
	}
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
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	inputLines := strings.Split(input, "\n")
	if len(inputLines) < 3 {
		panic("not enough data in input to parse")
	}

	directions, err := parseDirectionLine(inputLines[0])
	if err != nil {
		panic(fmt.Sprintf("failed to parse direction lien: %s", err))
	}

	nodeMap, err := parseMap(inputLines[2:])
	if err != nil {
		panic(fmt.Sprintf("failed to parse map: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(directions, nodeMap))
}

func part1(directions []Direction, nodeMap map[NodeAddress]NodeChoice) int {
	directionCursor := 0
	currentNode := NodeAddressStart
	steps := 0

	for currentNode != NodeAddressEnd {
		if currentNode == NodeAddressEnd {
			return steps
		}

		direction := directions[directionCursor]
		currentNode = nodeMap[currentNode].TakeDirection(direction)
		directionCursor = (directionCursor + 1) % len(directions)
		steps++
	}

	return steps
}

func parseDirectionLine(line string) ([]Direction, error) {
	directions := make([]Direction, len(line))
	for i, char := range line {
		if char == 'L' {
			directions[i] = DirectionLeft
		} else if char == 'R' {
			directions[i] = DirectionRight
		} else {
			return nil, fmt.Errorf("invalid direction char '%c'", char)
		}
	}

	return directions, nil
}

func parseMap(lines []string) (map[NodeAddress]NodeChoice, error) {
	mapNodes := make(map[NodeAddress]NodeChoice, len(lines))
	for _, line := range lines {
		source, choice, err := parseMapLine(line)
		if err != nil {
			return nil, fmt.Errorf("could not parse %q: %w", line, err)
		}

		mapNodes[source] = choice
	}

	return mapNodes, nil
}

func parseMapLine(line string) (NodeAddress, NodeChoice, error) {
	pattern := regexp.MustCompile(`^([A-Z]{3}) = \(([A-Z]{3}), ([A-Z]{3})\)$`)
	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		return "", NodeChoice{}, errors.New("malformed line")
	}

	source := NodeAddress(matches[1])
	choice := NodeChoice{left: NodeAddress(matches[2]), right: NodeAddress(matches[3])}

	return source, choice, nil
}
