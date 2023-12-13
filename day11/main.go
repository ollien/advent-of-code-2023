package main

import (
	"errors"
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
	expandedInputLines, err := expandInput(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to expand input: %s", err))
	}

	nodes := parseInputMatrix(expandedInputLines)
	fmt.Printf("Part 1: %d\n", part1(nodes))
}

func part1(nodes []Coordinate) int {
	type Pair struct {
		node1 Coordinate
		node2 Coordinate
	}

	pairs := []Pair{}
	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			pair := Pair{node1: nodes[i], node2: nodes[j]}
			pairs = append(pairs, pair)
		}
	}

	total := 0
	for _, pair := range pairs {
		distance := abs(pair.node2.col-pair.node1.col) + abs(pair.node2.row-pair.node1.row)
		total += abs(distance)
	}

	return total
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func expandInput(inputLines []string) ([][]rune, error) {
	if len(inputLines) == 0 {
		return nil, nil
	}

	runeMatrix := makeRuneMatrix(inputLines)
	expanded1 := expandInputVertically(runeMatrix)
	transposed1, err := transposeMatrix(expanded1)
	if err != nil {
		return nil, err
	}

	expanded2 := expandInputVertically(transposed1)
	transposed2, err := transposeMatrix(expanded2)
	if err != nil {
		return nil, err
	}

	return transposed2, nil
}

func makeRuneMatrix(strings []string) [][]rune {
	matrix := make([][]rune, len(strings))
	for i, line := range strings {
		matrix[i] = make([]rune, len(line))
		for j, char := range line {
			matrix[i][j] = char
		}
	}

	return matrix
}

func expandInputVertically(inputMatrix [][]rune) [][]rune {
	expanded := [][]rune{}
	for _, line := range inputMatrix {
		expanded = append(expanded, line)
		if count(line, '.') == len(line) {
			expanded = append(expanded, slices.Clone(line))
		}
	}

	return expanded
}

func transposeMatrix[T any, S ~[]T, M []S](matrix M) (M, error) {
	for _, row := range matrix {
		if len(row) != len(matrix[0]) {
			return nil, errors.New("not all rows are the same length")
		}
	}

	transposed := make(M, 0)
	for range matrix {
		transposed = append(transposed, make(S, len(matrix)))
	}

	for i, row := range matrix {
		for j, item := range row {
			transposed[j][i] = item
		}
	}

	return transposed, nil
}

func count[T comparable, S ~[]T](s S, target T) int {
	count := 0
	for _, item := range s {
		if item == target {
			count++
		}
	}

	return count
}

func parseInputMatrix(input [][]rune) []Coordinate {
	nodes := []Coordinate{}

	for row, line := range input {
		for col, char := range line {
			if char == '#' {
				nodes = append(nodes, Coordinate{row: row, col: col})
			}
		}
	}

	return nodes
}
