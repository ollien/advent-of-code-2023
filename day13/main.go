package main

import (
	"cmp"
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
	rawPatternSections := strings.Split(input, "\n\n")
	patternSections, err := tryParse(rawPatternSections, parsePatternSection)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(patternSections))
}

func part1(sections []map[Coordinate]struct{}) int {
	total := 0
	for i, section := range sections {
		sectionResults, err := evaluateSection(section)
		if err != nil {
			panic(fmt.Sprintf("invalid section %d: %s", i, err))
		}

		total += sectionResults
	}

	return total
}

func evaluateSection(section map[Coordinate]struct{}) (int, error) {
	maxRow := maxAlongAxis(Coordinate{row: 1, col: 0}, section)
	byRow := getAllAlongAxis(Coordinate{row: 1, col: 0}, section)
	for row := 0; row < maxRow; row++ {
		if isMirroredAcrossAxis(row, byRow) {
			return (row + 1) * 100, nil
		}
	}

	maxCol := maxAlongAxis(Coordinate{row: 0, col: 1}, section)
	byCol := getAllAlongAxis(Coordinate{row: 0, col: 1}, section)
	for col := 0; col < maxCol; col++ {
		if isMirroredAcrossAxis(col, byCol) {
			return col + 1, nil
		}
	}

	return 0, errors.New("section is not mirrored")
}

func isMirroredAcrossAxis(axisIdx int, byAxis map[int][]int) bool {
	maxAxisIdx := maxKey(byAxis)
	// Radiate "outwards" from the axis item, and check the corresponding elements
	for offset := 1; axisIdx+offset <= maxAxisIdx && axisIdx-(offset-1) >= 0; offset++ {
		rowItems := byAxis[axisIdx-(offset-1)]
		oppositeItems := byAxis[axisIdx+offset]
		if !slices.Equal(rowItems, oppositeItems) {
			return false
		}
	}

	return true
}

func maxAlongAxis(axis Coordinate, section map[Coordinate]struct{}) int {
	if !(axis.row == 0 && axis.col == 1 || axis.row == 1 && axis.col == 0) {
		panic("invalid axis")
	}

	res := 0
	for coord := range section {
		alongAxis := coord.row*axis.row + coord.col*axis.col
		if alongAxis > res {
			res = alongAxis
		}
	}

	return res
}

func getAllAlongAxis(axis Coordinate, section map[Coordinate]struct{}) map[int][]int {
	if !(axis.row == 0 && axis.col == 1 || axis.row == 1 && axis.col == 0) {
		panic("invalid axis")
	}

	res := map[int][]int{}
	for coord := range section {
		alongAxis := coord.row*axis.row + coord.col*axis.col
		notAlongAxis := coord.row*(1-axis.row) + coord.col*(1-axis.col)
		res[alongAxis] = append(res[alongAxis], notAlongAxis)
	}

	for _, perpendicularItems := range res {
		slices.Sort(perpendicularItems)
	}

	return res
}

func parsePatternSection(section string) (map[Coordinate]struct{}, error) {
	sectionLines := strings.Split(section, "\n")
	coords := map[Coordinate]struct{}{}
	for row, line := range sectionLines {
		for col, char := range line {
			if char == '#' {
				coords[Coordinate{row: row, col: col}] = struct{}{}
			} else if char != '.' {
				return nil, fmt.Errorf("invalid char: %c", char)
			}
		}
	}

	return coords, nil
}

func maxKey[T cmp.Ordered, U any](m map[T]U) T {
	if len(m) == 0 {
		panic("cannot get max key of zero length map")
	} else if len(m) == 1 {
		return *new(T)
	}

	keys := make([]T, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	return slices.Max(keys)
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
