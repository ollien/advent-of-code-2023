package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type Tile rune

const (
	TileEmpty Tile = '.'
	TileWall  Tile = '#'
	TileRight Tile = '>'
	TileLeft  Tile = '<'
	TileUp    Tile = '^'
	TileDown  Tile = 'v'
)

type Coordinate struct {
	Row int
	Col int
}

func main() {
	if len(os.Args) != 2 {
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
	grid, err := parseGrid(inputLines)
	if err != nil {
		panic(fmt.Sprintf("could not parse input: %s", err))
	}

	for _, line := range grid {
		for _, char := range line {
			fmt.Printf("%c", char)
		}
		fmt.Println()
	}

	fmt.Printf("Part 1: %d\n", part1(grid))
}

func part1(grid [][]Tile) int {
	startCol, err := findStartingTile(grid[0])
	if err != nil {
		panic(fmt.Sprintf("could not find starting tile: %s", err))
	}

	path := findLongestPath(
		Coordinate{Row: 0, Col: startCol},
		grid,
	)

	return len(path) - 1
}

func findStartingTile(firstRow []Tile) (int, error) {
	candidate := (*int)(nil)
	for col, item := range firstRow {
		if item == TileEmpty && candidate == nil {
			savedCol := col
			candidate = &savedCol
		} else if item == TileEmpty && candidate != nil {
			return 0, errors.New("more than one open position")
		}
	}

	if candidate == nil {
		return 0, errors.New("no open positions")
	}

	return *candidate, nil
}

func findLongestPath(start Coordinate, grid [][]Tile) []Coordinate {
	// position := start
	inBounds := func(pos Coordinate) bool {
		return pos.Row >= 0 && pos.Row < len(grid) && pos.Col >= 0 && pos.Col < len(grid[0])
	}

	var dfs func(Coordinate, []Coordinate) []Coordinate
	// visited := map[Coordinate]struct{}{}
	dfs = func(cursor Coordinate, path []Coordinate) []Coordinate {
		// fmt.Println(path)
		// visited[cursor] = struct{}{}

		upNeighbor := Coordinate{Row: cursor.Row - 1, Col: cursor.Col}
		downNeighbor := Coordinate{Row: cursor.Row + 1, Col: cursor.Col}
		leftNeighbor := Coordinate{Row: cursor.Row, Col: cursor.Col - 1}
		rightNeighbor := Coordinate{Row: cursor.Row, Col: cursor.Col + 1}
		switch grid[cursor.Row][cursor.Col] {
		case TileUp:
			// _, ok := visited[upNeighbor]
			ok := slices.Contains(path, upNeighbor)
			if inBounds(upNeighbor) && !ok {
				nextPath := slices.Clone(path)
				return dfs(upNeighbor, append(nextPath, upNeighbor))
			}
		case TileDown:
			// _, ok := visited[downNeighbor]
			ok := slices.Contains(path, downNeighbor)
			if inBounds(downNeighbor) && !ok {
				nextPath := slices.Clone(path)
				return dfs(downNeighbor, append(nextPath, downNeighbor))
			}
		case TileLeft:
			ok := slices.Contains(path, leftNeighbor)
			// _, ok := visited[leftNeighbor]
			if inBounds(leftNeighbor) && !ok {
				nextPath := slices.Clone(path)
				return dfs(leftNeighbor, append(nextPath, leftNeighbor))
			}
		case TileRight:
			ok := slices.Contains(path, rightNeighbor)
			// _, ok := visited[rightNeighbor]
			if inBounds(rightNeighbor) && !ok {
				nextPath := slices.Clone(path)
				return dfs(rightNeighbor, append(nextPath, rightNeighbor))
			}
		case TileEmpty:
			neighbors := []Coordinate{upNeighbor, downNeighbor, leftNeighbor, rightNeighbor}
			longestPath := path
			for _, neighbor := range neighbors {
				if !inBounds(neighbor) {
					continue
				} else if slices.Contains(path, neighbor) {
					continue
				}
				//else if _, ok := visited[neighbor]; ok {
				//	continue
				//}

				nextPath := slices.Clone(path)
				fullPath := dfs(neighbor, append(nextPath, neighbor))
				if len(fullPath) > len(longestPath) {
					longestPath = fullPath
				}
			}

			return longestPath
		case TileWall:
			return path
		}

		return path
	}

	return dfs(start, []Coordinate{})
}

func parseGrid(inputLines []string) ([][]Tile, error) {
	if len(inputLines) == 0 {
		return nil, errors.New("cannot parse grid of no length")
	}

	height := len(inputLines)
	width := len(inputLines[0])
	grid := make([][]Tile, height)
	for i, line := range inputLines {
		if len(line) != width {
			return nil, errors.New("grid's rows are not of consistent length")
		}

		for _, char := range line {
			if Tile(char) != TileWall && Tile(char) != TileEmpty && Tile(char) != TileUp && Tile(char) != TileDown && Tile(char) != TileLeft && Tile(char) != TileRight {
				return nil, fmt.Errorf("invalid tile char %c", char)
			}

			grid[i] = append(grid[i], Tile(char))
		}
	}

	return grid, nil
}
