// I definitely overcomplicated this problem, but it took me a very long time to visualize things properly

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

type ScanDirection int

const (
	ScanDirectionHorizontal ScanDirection = iota
	ScanDirectionVertical
)

type PipeMap map[Coordinate]Pipe

var ErrMissingPipe = errors.New("no pipe at location")

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

// CardinalNeighbors gets all the neighbors of the given coordinate (cardinal directions)
func (coordinate Coordinate) CardinalNeighbors() []Coordinate {
	return []Coordinate{
		coordinate.North(),
		coordinate.South(),
		coordinate.East(),
		coordinate.West(),
	}
}

// AllNeighbors gets all the neighbors of the given coordinate in all directions
func (coordinate Coordinate) AllNeighbors() []Coordinate {
	diagonalNeighbors := []Coordinate{
		{row: coordinate.row - 1, col: coordinate.col - 1},
		{row: coordinate.row - 1, col: coordinate.col + 1},
		{row: coordinate.row + 1, col: coordinate.col - 1},
		{row: coordinate.row + 1, col: coordinate.col + 1},
	}

	return append(coordinate.CardinalNeighbors(), diagonalNeighbors...)
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

// IsCorner indicates whether or not a pipe is a corner
func (pipe Pipe) IsCorner() bool {
	return pipe == PipeJ || pipe == Pipe7 || pipe == PipeF || pipe == PipeL
}

// IsStraight indicates whether or not a pipe is straight
func (pipe Pipe) IsStraight() bool {
	return pipe == PipeHorizontal || pipe == PipeVertical
}

// IsParallelToScan indicates whether or not a pipe moves only parallel to the scan direction
func (pipe Pipe) IsParallelToScan(scanDir ScanDirection) bool {
	if scanDir == ScanDirectionHorizontal {
		return pipe == PipeHorizontal
	} else {
		return pipe == PipeVertical
	}
}

// ConnectedNeighbors gets only the connected neighbors to a pipe at a position
func (pipeMap PipeMap) ConnectedNeighbors(position Coordinate) []Coordinate {
	_, ok := pipeMap[position]
	if !ok {
		return []Coordinate{}
	}

	connectedNeighbors := []Coordinate{}
	for _, neighbor := range position.CardinalNeighbors() {
		if pipeMap.PipesConnect(position, neighbor) {
			connectedNeighbors = append(connectedNeighbors, neighbor)
		}
	}

	return connectedNeighbors
}

// PipeBounds finds the bounds of the pieps on the map
func (pipeMap PipeMap) PipeBounds() (Coordinate, Coordinate) {
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

	return Coordinate{row: minRow, col: minCol}, Coordinate{row: maxRow, col: maxCol}
}

// Print will print the entire map in the form the puzzle presents it
func (pipeMap PipeMap) Print() {
	minCorner, maxCorner := pipeMap.PipeBounds()
	for row := minCorner.row; row <= maxCorner.row; row++ {
		for col := minCorner.col; col <= maxCorner.col; col++ {
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
	fmt.Printf("Part 2: %d\n", part2(pipeMap, startPosition))
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

		for _, neighbor := range pipeMap.ConnectedNeighbors(visiting.position) {
			if _, ok := visited[neighbor]; ok {
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

func part2(pipeMap PipeMap, startPosition Coordinate) int {
	mainLoopMap := traceMainLoop(pipeMap, startPosition)

	regions := findEmptyRegions(mainLoopMap, startPosition)
	area := 0
	for _, region := range regions {
		isAccessible := isRegionExternallyAccessible(mainLoopMap, region)
		if !isAccessible {
			area += len(region)
		}
	}

	return area
}

// traceMainLoop walks the pipes and finds the pipes relevant to the problem
func traceMainLoop(pipeMap PipeMap, startPosition Coordinate) PipeMap {
	visited := map[Coordinate]Pipe{}
	toVisit := []Coordinate{startPosition}
	for len(toVisit) > 0 {
		visitingPosition := toVisit[0]
		toVisit = toVisit[1:]

		visited[visitingPosition] = pipeMap[visitingPosition]

		for _, neighbor := range pipeMap.ConnectedNeighbors(visitingPosition) {
			if _, ok := visited[neighbor]; ok {
				continue
			}

			toVisit = append(toVisit, neighbor)
		}
	}

	return visited
}

// isRegionExternallyAccessible indicates whether or not all of the given coordinates are internal to the loop
func isRegionExternallyAccessible(pipeMap PipeMap, region []Coordinate) bool {
	for _, pos := range region {
		if !isInsideViaRay(pipeMap, pos, ScanDirectionHorizontal) || !isInsideViaRay(pipeMap, pos, ScanDirectionVertical) {
			return true
		}
	}

	return false
}

// isInsideViaRay casts a ray in the given direction, counting the number of edge crossings to determine
// if a tile is inside
func isInsideViaRay(pipeMap PipeMap, target Coordinate, direction ScanDirection) bool {
	cursor := target
	if direction == ScanDirectionHorizontal {
		cursor.col = 0
	} else {
		cursor.row = 0
	}

	pipeBuffer := []Pipe{}
	crossings := 0
	for cursor != target {
		if pipeMap[cursor] != PipeUnknown {
			pipeBuffer = append(pipeBuffer, pipeMap[cursor])
		}

		if direction == ScanDirectionHorizontal {
			cursor.col++
		} else {
			cursor.row++
		}
	}
	crossings += numRayCrossings(direction, pipeBuffer)

	return crossings%2 == 1
}

// numRayCrossings counts the number of times a ray crosses a pipe
func numRayCrossings(scanDirection ScanDirection, scannedPipes []Pipe) int {
	crossings := 0
	for _, pipe := range scannedPipes {
		if pipe.IsStraight() && !pipe.IsParallelToScan(scanDirection) {
			crossings++
		} else if pipe.IsCorner() {
			if scanDirection == ScanDirectionHorizontal && (pipe == PipeL || pipe == PipeJ) {
				// If we are scanning horizontally, and we cross one of these chars, we are internal if we cross something of this variety only once
				crossings++
			} else if scanDirection == ScanDirectionVertical && (pipe == PipeF || pipe == PipeL) {
				// ditto for vertically
				crossings++
			}
		}
	}

	return crossings
}

// findEmptyRegions finds all of the locations where there are empty positions on the graph
func findEmptyRegions(pipeMap PipeMap, startPosition Coordinate) [][]Coordinate {
	emptyPositions := findEmptyPositions(pipeMap)
	regions := [][]Coordinate{}
	visited := map[Coordinate]struct{}{}
	for _, position := range emptyPositions {
		if _, ok := visited[position]; ok {
			continue
		}

		floodedPositions := flood(pipeMap, position)
		regions = append(regions, floodedPositions)
		for _, flooded := range floodedPositions {
			visited[flooded] = struct{}{}
		}
	}

	return regions
}

// findEmptyPositions finds all empty positions on the graph
func findEmptyPositions(pipeMap PipeMap) []Coordinate {
	minCorner, maxCorner := pipeMap.PipeBounds()
	empty := []Coordinate{}
	for row := minCorner.row; row < maxCorner.row; row++ {
		for col := minCorner.col; col < maxCorner.col; col++ {
			pos := Coordinate{row: row, col: col}
			if pipeMap[pos] == PipeUnknown {
				empty = append(empty, pos)
			}
		}
	}

	return empty
}

// flood performs a flood fill to locate neighboring empty spots
func flood(pipeMap PipeMap, start Coordinate) []Coordinate {
	minCorner, maxCorner := pipeMap.PipeBounds()
	toVisit := []Coordinate{start}
	visited := map[Coordinate]struct{}{}
	for len(toVisit) > 0 {
		visiting := toVisit[0]
		toVisit = toVisit[1:]
		if _, ok := pipeMap[visiting]; ok {
			// If we've hit a pipe on the bounding box, don't keep filling
			continue
		} else if visiting.row < minCorner.row || visiting.col < minCorner.col || visiting.row > maxCorner.row || visiting.col > maxCorner.col {
			// if we've moved out of bounds, don't continue either
			continue
		} else if _, ok := visited[visiting]; ok {
			// If we've already visited this, we don't need to try again
			continue
		}

		visited[visiting] = struct{}{}
		toVisit = append(toVisit, visiting.CardinalNeighbors()...)
	}

	return mapKeys(visited)
}

func parsePipeMap(inputLines []string) (PipeMap, Coordinate, error) {
	pipeMap := PipeMap{}
	startPosition := (*Coordinate)(nil)
	for row, line := range inputLines {
		for col, char := range line {
			pipe, err := parsePipeChar(char)
			if errors.Is(err, ErrMissingPipe) {
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
		return PipeUnknown, ErrMissingPipe
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
