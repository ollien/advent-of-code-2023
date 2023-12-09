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
	fmt.Printf("Part 2: %d\n", part2(directions, nodeMap))
}

func part1(directions []Direction, nodeMap map[NodeAddress]NodeChoice) int {
	const (
		NodeAddressStart NodeAddress = "AAA"
		NodeAddressEnd   NodeAddress = "ZZZ"
	)

	directionCursor := 0
	currentNode := NodeAddressStart
	steps := 0

	for currentNode != NodeAddressEnd {
		direction := directions[directionCursor]
		currentNode = nodeMap[currentNode].TakeDirection(direction)
		directionCursor = (directionCursor + 1) % len(directions)
		steps++
	}

	return steps
}

func part2(directions []Direction, nodeMap map[NodeAddress]NodeChoice) int {
	directionCursor := 0
	nodes := findPart2StartingNodes(nodeMap)
	if len(nodes) == 0 {
		panic("no starting nodes")
	}

	steps := 0
	encounteredEnd := []int{}

	for len(encounteredEnd) != len(nodes) {
		direction := directions[directionCursor]
		for i, node := range nodes {
			nodes[i] = nodeMap[node].TakeDirection(direction)
			if nodeEndsIn(nodes[i], 'Z') {
				encounteredEnd = append(encounteredEnd, steps+1)
			}
		}

		directionCursor = (directionCursor + 1) % len(directions)
		steps++
	}

	// Once we have encountered all the steps to get to each ending, the LCM will find the first time they all match
	return sliceLCM(encounteredEnd)
}

// sliceLCM finds the LCM of the numbers in the given slice. Panics if the slice is of length zero
func sliceLCM(nums []int) int {
	if len(nums) == 0 {
		panic("cannot find lcm of zero numbers")
	}

	result := nums[0]
	for _, n := range nums[1:] {
		result = lcm(result, n)
	}

	return result
}

func lcm(a, b int) int {
	return b * (a / gcd(a, b))
}

func gcd(a, b int) int {
	// https://en.wikipedia.org/wiki/Euclidean_algorithm
	factor := a
	rem := b
	for rem != 0 {
		oldRem := rem
		rem = factor % rem
		factor = oldRem
	}

	return factor
}

func findPart2StartingNodes(nodeMap map[NodeAddress]NodeChoice) []NodeAddress {
	startNodes := []NodeAddress{}
	for addr := range nodeMap {
		if nodeEndsIn(addr, 'A') {
			startNodes = append(startNodes, addr)
		}
	}

	return startNodes
}

func nodeEndsIn(addr NodeAddress, c byte) bool {
	return addr[len(addr)-1] == c
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
	pattern := regexp.MustCompile(`^([0-9A-Z]{2}[A-Z]) = \(([0-9A-Z]{2}[A-Z]), ([0-9A-Z]{2}[A-Z])\)$`)
	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		return "", NodeChoice{}, errors.New("malformed line")
	}

	source := NodeAddress(matches[1])
	choice := NodeChoice{left: NodeAddress(matches[2]), right: NodeAddress(matches[3])}

	return source, choice, nil
}
