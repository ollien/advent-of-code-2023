package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type Tile int

const (
	TileEmpty Tile = iota
	TileRoundRock
	TileCubeRock
)

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
}

func part1(inputGrid [][]Tile) int {
	rollNorth(inputGrid)
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

func rollNorth(inputGrid [][]Tile) {
	for row, line := range inputGrid {
		if row == 0 {
			continue
		}

		for col, tile := range line {
			if tile != TileRoundRock {
				continue
			}

			lastEmptyRow := row
			for candidateRow := row - 1; candidateRow >= 0; candidateRow-- {
				if inputGrid[candidateRow][col] == TileEmpty {
					lastEmptyRow = candidateRow
				} else {
					break
				}
			}

			inputGrid[row][col] = TileEmpty
			inputGrid[lastEmptyRow][col] = TileRoundRock
		}
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
