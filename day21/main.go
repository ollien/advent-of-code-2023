package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strings"
)

type Tile rune

const (
	TileTypeGarden Tile = '.'
	TileTypeRock   Tile = '#'
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
	fmt.Printf("Part 2: %f\n", part2(grid, start))
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

func part2(tiles map[Coordinate]Tile, start Coordinate) float64 {
	cursors := []Coordinate{start}
	minRow, maxRow, minCol, maxCol := gridSize(tiles)

	coordIdx := 0
	x := [3]float64{}
	y := [3]float64{}

	for i := 1; coordIdx < 3; i++ {
		nextCursors := []Coordinate{}
		visited := map[Coordinate]struct{}{}
		for _, cursor := range cursors {
			for _, neighbor := range neighbors(cursor) {
				rowRange := maxRow - minRow + 1
				colRange := maxCol - minCol + 1
				wrappedNeighbor := Coordinate{
					Row: ((neighbor.Row-minRow)%rowRange+maxRow+1)%rowRange + minRow,
					Col: ((neighbor.Col-minCol)%colRange+maxCol+1)%colRange + minCol,
				}
				if tiles[wrappedNeighbor] == TileTypeRock {
					continue
				} else if _, ok := visited[neighbor]; ok {
					continue
				}

				visited[neighbor] = struct{}{}
				nextCursors = append(nextCursors, neighbor)
			}
		}
		if (i-65)%(131) == 0 {
			x[coordIdx] = float64(i)
			y[coordIdx] = float64(len(visited))
			coordIdx++
		}

		cursors = nextCursors
	}

	return fitQuadratic(x, y, 26501365)
}

func fitQuadratic(x [3]float64, y [3]float64, desired float64) float64 {
	a0 := y[0]
	a1 := (y[1] - y[0]) / (x[1] - x[0])

	a2Numerator := (y[2]-y[1])/(x[2]-x[1]) - a1
	a2Denominator := x[2] - x[0]
	a2 := a2Numerator / a2Denominator

	// https://pythonnumericalmethods.berkeley.edu/notebooks/chapter17.05-Newtons-Polynomial-Interpolation.html
	return a2*(desired-x[1])*(desired-x[0]) + a1*(desired-x[0]) + a0
}

// not used, left for debugging
func printGrid(tiles map[Coordinate]Tile, cursors []Coordinate) {
	scale := 3
	minRow, maxRow, minCol, maxCol := gridSize(tiles)
	maxRow++
	maxCol++
	for i := minRow - maxRow*scale; i < maxRow*scale; i++ {
		for j := minCol - maxCol*scale; j < maxCol*scale; j++ {
			rowRange := maxRow - minRow
			colRange := maxCol - minCol
			p := Coordinate{Row: i, Col: j}
			pos := Coordinate{
				Row: ((p.Row-minRow)%rowRange+maxRow)%rowRange + minRow,
				Col: ((p.Col-minCol)%colRange+maxCol)%colRange + minCol,
			}
			if slices.Index(cursors, p) != -1 {
				fmt.Printf("\033[0;31mO\033[0m")
			} else {
				fmt.Printf("%c", tiles[pos])
			}

		}
		fmt.Println()
	}
	fmt.Println()
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

func gridSize(tiles map[Coordinate]Tile) (minRow, maxRow, minCol, maxCol int) {
	minRow = math.MaxInt
	minCol = math.MaxInt

	for coord := range tiles {
		minRow = min(coord.Row, minRow)
		maxRow = max(coord.Row, maxRow)
		minCol = min(coord.Col, minCol)
		maxCol = max(coord.Col, maxCol)
	}

	return
}
