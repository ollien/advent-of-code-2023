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

type Pipe int

const (
	PipeUnknown Pipe = iota
	PipeVertical
	PipeHorizontal
	PipeL
	PipeJ
	Pipe7
	PipeF
)

type PipeMap map[Coordinate]Pipe

var ErrNoPipe = errors.New("no pipe at location")

func (coordinate Coordinate) North() Coordinate {
	return Coordinate{row: coordinate.row - 1, col: coordinate.col}
}

func (coordinate Coordinate) South() Coordinate {
	return Coordinate{row: coordinate.row + 1, col: coordinate.col}
}

func (coordinate Coordinate) East() Coordinate {
	return Coordinate{row: coordinate.row, col: coordinate.col + 1}
}

func (coordinate Coordinate) West() Coordinate {
	return Coordinate{row: coordinate.row, col: coordinate.col - 1}
}

// Neighbors gets all the neighbors of the given coordinate (cardinal directions)
func (coordinate Coordinate) Neighbors() []Coordinate {
	return []Coordinate{
		coordinate.North(),
		coordinate.South(),
		coordinate.East(),
		coordinate.West(),
	}
}

// ConnectsNorth will determine if the given pipe can connect to a pipe to its north
func (pipe Pipe) ConnectsNorth() bool {
	return pipe == PipeVertical || pipe == PipeL || pipe == PipeJ
}

// ConnectsSouth will determine if the given pipe can connect to a pipe to its south
func (pipe Pipe) ConnectsSouth() bool {
	return pipe == PipeVertical || pipe == Pipe7 || pipe == PipeF
}

// ConnectsEast will determine if the given pipe can connect to a pipe to its east
func (pipe Pipe) ConnectsEast() bool {
	return pipe == PipeHorizontal || pipe == PipeF || pipe == PipeL
}

// ConnectsWest will determine if the given pipe can connect to a pipe to its west
func (pipe Pipe) ConnectsWest() bool {
	return pipe == PipeHorizontal || pipe == Pipe7 || pipe == PipeJ
}

// Print will print the entire map in the form the puzzle presents it
func (pipeMap PipeMap) Print() {
	positions := mapKeys(pipeMap)
	compareRow := func(a, b Coordinate) int {
		return cmp.Compare(a.row, b.row)
	}

	compareCol := func(a, b Coordinate) int {
		return cmp.Compare(a.col, b.col)
	}

	minRow := slices.MinFunc(positions, compareRow).row
	maxRow := slices.MaxFunc(positions, compareRow).row
	minCol := slices.MinFunc(positions, compareCol).col
	maxCol := slices.MaxFunc(positions, compareCol).col

	for row := minRow; row <= maxRow; row++ {
		for col := minCol; col <= maxCol; col++ {
			location := Coordinate{row: row, col: col}
			pipe, ok := pipeMap[location]
			if !ok {
				fmt.Print(".")
			} else if pipe == PipeVertical {
				fmt.Print("|")
			} else if pipe == PipeHorizontal {
				fmt.Print("-")
			} else if pipe == PipeL {
				fmt.Print("L")
			} else if pipe == PipeJ {
				fmt.Print("J")
			} else if pipe == Pipe7 {
				fmt.Print("7")
			} else if pipe == PipeF {
				fmt.Print("F")
			}
		}
		fmt.Println()
	}
}

// PipesConnect will check if the pipes at the given positions connect
func (pipeMap PipeMap) PipesConnect(position1, position2 Coordinate) bool {
	pipe1 := pipeMap[position1]
	pipe2 := pipeMap[position2]

	if pipe1.ConnectsNorth() && pipe2.ConnectsSouth() && position2 == position1.North() {
		return true
	} else if pipe1.ConnectsSouth() && pipe2.ConnectsNorth() && position2 == position1.South() {
		return true
	} else if pipe1.ConnectsEast() && pipe2.ConnectsWest() && position2 == position1.East() {
		return true
	} else if pipe1.ConnectsWest() && pipe2.ConnectsEast() && position2 == position1.West() {
		return true
	} else {
		return false
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
	pipeMap, startPosition, err := parsePipeMap(inputLines)
	if err != nil {
		panic(fmt.Sprintf("could not parse pipes: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(pipeMap, startPosition))
}

func part1(pipeMap PipeMap, startPosition Coordinate) int {
	type Visit struct {
		position Coordinate
		distance int
	}
	maxDistance := 0
	visited := map[Coordinate]struct{}{}
	toVisit := []Visit{{position: startPosition, distance: 0}}
	for len(toVisit) > 0 {
		visiting := toVisit[0]
		toVisit = toVisit[1:]

		visited[visiting.position] = struct{}{}
		if visiting.distance > maxDistance {
			maxDistance = visiting.distance
		}

		for _, neighbor := range visiting.position.Neighbors() {
			if pipeMap[neighbor] == PipeUnknown || !pipeMap.PipesConnect(visiting.position, neighbor) {
				continue
			} else if _, ok := visited[neighbor]; ok {
				continue
			}

			neighborVisit := Visit{
				position: neighbor,
				distance: visiting.distance + 1,
			}

			toVisit = append(toVisit, neighborVisit)
		}
	}

	return maxDistance
}

func parsePipeMap(inputLines []string) (PipeMap, Coordinate, error) {
	pipeMap := PipeMap{}
	startPosition := (*Coordinate)(nil)
	for row, line := range inputLines {
		for col, char := range line {
			pipe, err := parsePipeChar(char)
			if errors.Is(err, ErrNoPipe) {
				continue
			} else if err != nil {
				return nil, Coordinate{}, fmt.Errorf("malformed at (%d, %d): %w", row, col, err)
			}

			position := Coordinate{row: row, col: col}
			pipeMap[position] = pipe
			if pipe == PipeUnknown {
				startPosition = &position
			}
		}
	}

	if startPosition == nil {
		return nil, Coordinate{}, errors.New("no start position found")
	}

	startPipeType, err := inferPipeType(pipeMap, *startPosition)
	if err != nil {
		return nil, Coordinate{}, fmt.Errorf("infer start pipe type: %w", err)
	}

	pipeMap[*startPosition] = startPipeType

	return pipeMap, *startPosition, nil
}

func parsePipeChar(c rune) (Pipe, error) {
	switch c {
	case '.':
		return PipeUnknown, ErrNoPipe
	case '|':
		return PipeVertical, nil
	case '-':
		return PipeHorizontal, nil
	case 'L':
		return PipeL, nil
	case 'J':
		return PipeJ, nil
	case '7':
		return Pipe7, nil
	case 'F':
		return PipeF, nil
	case 'S':
		return PipeUnknown, nil
	default:
		return PipeUnknown, fmt.Errorf("invalid pipe char %c", c)
	}
}

// inferPipeType will use the PipeMap to infer what type of Pipe should exist at the given location
func inferPipeType(pipeMap PipeMap, position Coordinate) (Pipe, error) {
	north := pipeMap[Coordinate{row: position.row - 1, col: position.col}]
	south := pipeMap[Coordinate{row: position.row + 1, col: position.col}]
	west := pipeMap[Coordinate{row: position.row, col: position.col - 1}]
	east := pipeMap[Coordinate{row: position.row, col: position.col + 1}]

	if north.ConnectsSouth() && south.ConnectsNorth() {
		return PipeVertical, nil
	} else if west.ConnectsEast() && east.ConnectsWest() {
		return PipeHorizontal, nil
	} else if north.ConnectsSouth() && east.ConnectsWest() {
		return PipeL, nil
	} else if north.ConnectsSouth() && west.ConnectsEast() {
		return PipeJ, nil
	} else if south.ConnectsNorth() && west.ConnectsEast() {
		return Pipe7, nil
	} else if south.ConnectsNorth() && east.ConnectsWest() {
		return PipeF, nil
	} else {
		return PipeUnknown, fmt.Errorf("infer pipe type: north=%v south=%v west=%v east=%v", north, south, west, east)
	}
}

func mapKeys[T comparable, U any](m map[T]U) []T {
	res := make([]T, len(m))
	i := 0
	for key := range m {
		res[i] = key
		i++
	}

	return res
}
