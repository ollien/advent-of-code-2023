package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Tile rune

const (
	TileTypeGarden Tile = '.'
	TileTypeRock        = '#'
)

type Coordinate struct {
	Row int
	Col int
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
	grid, start, err := parseGrid(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(grid, start))
}

func part1(tiles map[Coordinate]Tile, start Coordinate) int {
	cursors := []Coordinate{start}
	lastCount := 0
	for i := 0; i < 64; i++ {
		nextCursors := []Coordinate{}
		visited := map[Coordinate]struct{}{}
		for _, cursor := range cursors {
			for _, neighbor := range neighbors(cursor) {
				if tiles[neighbor] == TileTypeRock {
					continue
				} else if _, ok := visited[neighbor]; ok {
					continue
				}

				visited[neighbor] = struct{}{}
				nextCursors = append(nextCursors, neighbor)
			}
		}

		cursors = nextCursors
		lastCount = len(visited)
	}

	return lastCount
}

func neighbors(coord Coordinate) []Coordinate {
	return []Coordinate{
		{Row: coord.Row + 1, Col: coord.Col},
		{Row: coord.Row - 1, Col: coord.Col},
		{Row: coord.Row, Col: coord.Col + 1},
		{Row: coord.Row, Col: coord.Col - 1},
	}
}

func parseGrid(inputLines []string) (map[Coordinate]Tile, Coordinate, error) {
	tiles := map[Coordinate]Tile{}
	start := Coordinate{}
	startFound := false

	for row, line := range inputLines {
		for col, char := range line {
			coordinate := Coordinate{Row: row, Col: col}
			switch char {
			case 'S':
				startFound = true
				start = coordinate
				tiles[coordinate] = TileTypeGarden
			case '.', '#':
				tiles[coordinate] = Tile(char)
			default:
				return nil, Coordinate{}, fmt.Errorf("invalid tile char %c", char)
			}
		}
	}

	if !startFound {
		return nil, Coordinate{}, errors.New("no start tile found")
	}

	return tiles, start, nil
}
