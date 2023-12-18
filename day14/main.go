package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const Part2Cycles = 1000000000

type Tile int
type Direction int

const (
	TileEmpty Tile = iota
	TileRoundRock
	TileCubeRock
)

const (
	DirectionNorth Direction = iota
	DirectionWest
	DirectionSouth
	DirectionEast
)

type Coordinate struct {
	row int
	col int
}

func (t Tile) String() string {
	switch t {
	case TileEmpty:
		return "."
	case TileRoundRock:
		return "O"
	case TileCubeRock:
		return "#"
	default:
		panic(fmt.Sprintf("invalid tile value %d", t))
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
	grid, err := parseTileGrid(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(Clone2D(grid)))
	fmt.Printf("Part 2: %d\n", part2(Clone2D(grid)))
}

func part1(inputGrid [][]Tile) int {
	rollDirection(inputGrid, DirectionNorth)

	return calculateNorthernLoad(inputGrid)
}

func part2(inputGrid [][]Tile) int {
	period := -1
	previouslySeenStates := map[string]struct{}{}
	for i := 0; i < Part2Cycles; i++ {
		rollCycle(inputGrid)
		serialized := ""
		for _, row := range inputGrid {
			for _, tile := range row {
				serialized += tile.String()
			}
		}

		_, ok := previouslySeenStates[serialized]
		if ok {
			period = i
			break
		}

		previouslySeenStates[serialized] = struct{}{}
	}

	if period == -1 {
		// if SOMEHOW we did not find a period, I guess we just finished the simulation
		return calculateNorthernLoad(inputGrid)
	}

	nextCycleIter := Part2Cycles / period * period
	// finish off the rest
	for i := nextCycleIter - 1; i <= Part2Cycles; i++ {
		rollCycle(inputGrid)
	}

	return calculateNorthernLoad(inputGrid)
}

func calculateNorthernLoad(inputGrid [][]Tile) int {
	load := 0
	for row, rowItems := range inputGrid {
		for _, tile := range rowItems {
			if tile == TileRoundRock {
				load += len(inputGrid) - row
			}
		}
	}

	return load
}

// rollCycle will run through all the directions in a cycle and roll in each of them
func rollCycle(inputGrid [][]Tile) {
	for direction := DirectionNorth; direction <= DirectionEast; direction++ {
		rollDirection(inputGrid, direction)
	}
}

// rollDirection will roll each round rock to the maximum possible position in that direction
func rollDirection(inputGrid [][]Tile, direction Direction) {
	iterateAgainstDirection(inputGrid, direction, func(row, col int) {
		tile := inputGrid[row][col]
		if tile != TileRoundRock {
			return
		}

		lastEmpty := findLastEmptyInDirection(inputGrid, row, col, direction)

		inputGrid[row][col] = TileEmpty
		inputGrid[lastEmpty.row][lastEmpty.col] = TileRoundRock
	})
}

// findLastEmptyInDirection will find the next open position in the given direction.
// If none are available, the original coordinate is returned (which is sufficient for this puzzle)
func findLastEmptyInDirection(inputGrid [][]Tile, row, col int, direction Direction) Coordinate {
	dRow, dCol := makeRayForDirection(direction)
	lastEmpty := Coordinate{row: row, col: col}
	cursor := Coordinate{row: row + dRow, col: col + dCol}
	for cursor.row >= 0 && cursor.col >= 0 && cursor.row < len(inputGrid) && cursor.col < len(inputGrid[row]) {
		if inputGrid[cursor.row][cursor.col] == TileEmpty {
			lastEmpty = cursor
		} else {
			break
		}

		cursor = Coordinate{row: cursor.row + dRow, col: cursor.col + dCol}
	}

	return lastEmpty
}

// iterateAgainstDirection will iterate over the grid in such a way that the iteration order moves against the given
// direction. This is useful for rolling as we do not want to have to continually recompute our rolls; going
// backwards is helpful.
//
// The first item in any direction is skipped, as it can't be used for rolling
func iterateAgainstDirection(inputGrid [][]Tile, direction Direction, fun func(row, col int)) {
	switch direction {
	case DirectionNorth, DirectionWest:
		for row := 0; row < len(inputGrid); row++ {
			if row == 0 && direction == DirectionNorth {
				continue
			}

			for col := 0; col < len(inputGrid[row]); col++ {
				if col == 0 && direction == DirectionWest {
					continue
				}

				fun(row, col)
			}
		}
	case DirectionSouth, DirectionEast:
		for row := len(inputGrid) - 1; row >= 0; row-- {
			if row == len(inputGrid)-1 && direction == DirectionSouth {
				continue
			}
			for col := len(inputGrid[row]) - 1; col >= 0; col-- {
				if col == len(inputGrid[row])-1 && direction == DirectionEast {
					continue
				}
				fun(row, col)
			}
		}
	default:
		panic(fmt.Sprintf("invalid direction %d", direction))
	}
}

// makeRayForDirection gets the direction to scan for empty tiles (row, col) for the given direction
func makeRayForDirection(direction Direction) (int, int) {
	switch direction {
	case DirectionNorth:
		return -1, 0
	case DirectionSouth:
		return 1, 0
	case DirectionEast:
		return 0, 1
	case DirectionWest:
		return 0, -1
	default:
		panic(fmt.Sprintf("invalid direction %d", direction))
	}
}

func Clone2D[T any, S ~[][]T](grid S) S {
	clone := make(S, len(grid))
	for i, row := range grid {
		clone[i] = slices.Clone(row)
	}

	return clone
}

func parseTileGrid(inputLines []string) ([][]Tile, error) {
	if len(inputLines) == 0 {
		return [][]Tile{}, nil
	}

	grid := make([][]Tile, len(inputLines))
	for i, line := range inputLines {
		if len(line) != len(inputLines[0]) {
			return nil, errors.New("lines have uneven lengths")
		}

		for _, char := range line {
			tile, err := tileForRune(char)
			if err != nil {
				return nil, err
			}

			grid[i] = append(grid[i], tile)
		}
	}

	return grid, nil
}

func tileForRune(r rune) (Tile, error) {
	switch r {
	case '.':
		return TileEmpty, nil
	case 'O':
		return TileRoundRock, nil
	case '#':
		return TileCubeRock, nil
	default:
		return TileEmpty, fmt.Errorf("invalid tile char %c", r)
	}
}
