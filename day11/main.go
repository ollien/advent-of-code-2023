package main

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type Coordinate struct {
	row int
	col int
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

	nodes, err := parseInputMatrix(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(nodes))
	fmt.Printf("Part 2: %d\n", part2(nodes))
}

func part1(nodes []Coordinate) int {
	expanded := expandUniverse(nodes, 2)
	return computePairwiseDistanceTotal(expanded)
}

func part2(nodes []Coordinate) int {
	expanded := expandUniverse(nodes, 1_000_000)
	return computePairwiseDistanceTotal(expanded)
}

func computePairwiseDistanceTotal(nodes []Coordinate) int {
	type Pair struct {
		node1 Coordinate
		node2 Coordinate
	}

	total := 0
	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			pair := Pair{node1: nodes[i], node2: nodes[j]}
			distance := abs(pair.node2.col-pair.node1.col) + abs(pair.node2.row-pair.node1.row)

			total += distance
		}
	}

	return total
}

func expandUniverse(nodes []Coordinate, expansion int) []Coordinate {
	rowExpanded := expandAlongAxis(
		nodes,
		expansion,
		func(c Coordinate) int { return c.row },
		func(c Coordinate, n int) Coordinate { return Coordinate{row: n, col: c.col} },
	)

	expanded := expandAlongAxis(
		rowExpanded,
		expansion,
		func(c Coordinate) int { return c.col },
		func(c Coordinate, n int) Coordinate { return Coordinate{row: c.row, col: n} },
	)

	return expanded
}

// expandAlongAxis will expand by the given amount the universe only along a single axis.
// getAxis will allow the function to get the value of a single axis, given a coordinate,
// and setAxis must return a new coordinate with the same axis set to the given value.
func expandAlongAxis(
	nodes []Coordinate,
	expansion int,
	getAxis func(Coordinate) int,
	setAxisValue func(Coordinate, int) Coordinate,
) []Coordinate {
	expanded := slices.Clone(nodes)
	slices.SortFunc(expanded, func(a, b Coordinate) int { return cmp.Compare(getAxis(a), getAxis(b)) })

	blank := findBlankAxisValues(nodes, getAxis)
	totalExpansion := 0
	for _, blankIdx := range blank {
		for i := range expanded {
			axisValue := getAxis(expanded[i])
			if axisValue <= blankIdx+totalExpansion {
				continue
			}

			expanded[i] = setAxisValue(expanded[i], axisValue+(expansion-1))
		}
		totalExpansion += (expansion - 1)
	}

	return expanded
}

// findBlankAxisValues finds all the positions where all values along the given axis are blank
func findBlankAxisValues(nodes []Coordinate, getAxis func(Coordinate) int) []int {
	maxCoord := slices.MaxFunc(nodes, func(node1, node2 Coordinate) int {
		return cmp.Compare(getAxis(node1), getAxis(node2))
	})

	axisMax := getAxis(maxCoord)
	knownAlongAxis := mapToSet(nodes, getAxis)
	blank := []int{}
	for i := 0; i <= axisMax; i++ {
		if _, ok := knownAlongAxis[i]; !ok {
			blank = append(blank, i)
		}
	}

	return blank
}

// mapToSet maps all values of the given input and turns them into a set
func mapToSet[T any, U comparable, S ~[]T](items S, mapper func(T) U) map[U]struct{} {
	res := map[U]struct{}{}
	for _, item := range items {
		key := mapper(item)
		res[key] = struct{}{}
	}

	return res
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func parseInputMatrix(input []string) ([]Coordinate, error) {
	nodes := []Coordinate{}

	for row, line := range input {
		for col, char := range line {
			if char == '#' {
				nodes = append(nodes, Coordinate{row: row, col: col})
			} else if char != '.' {
				return nil, fmt.Errorf("invalid input char %c", char)
			}
		}
	}

	return nodes, nil
}
