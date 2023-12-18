package main

import (
	"container/heap"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Direction int

const (
	DirectionNorth Direction = iota
	DirectionEast
	DirectionSouth
	DirectionWest
)

type Coordinate struct {
	Row                 int
	Col                 int
	FromDirection       Direction
	NumMovesInDirection int
}

type Heap[T comparable] struct {
	compare func(T, T) int
	data    []T
	// a cache of the item that are in data so that we can
	exists map[T]struct{}
}

var _ heap.Interface = &Heap[int]{}

func NewHeap[T comparable](compare func(T, T) int) *Heap[T] {
	h := &Heap[T]{
		compare: compare,
		data:    []T{},
		exists:  map[T]struct{}{},
	}
	heap.Init(h)

	return h
}

func (h *Heap[T]) Len() int {
	return len(h.data)
}

func (h *Heap[T]) Less(i, j int) bool {
	return h.compare(h.data[i], h.data[j]) < 0
}

func (h *Heap[T]) Swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

// Push pushes an item to the end of the underlying list. This is NOT a heap push. Check PushItem
func (h *Heap[T]) Push(x any) {
	h.exists[x.(T)] = struct{}{}
	h.data = append(h.data, x.(T))
}

// Pop will pop the last element from the end of the underlying list. This is NOT a heap pop. Check
// PopMin
func (h *Heap[T]) Pop() any {
	if len(h.data) == 0 {
		return nil
	}

	back := h.data[len(h.data)-1]
	delete(h.exists, back)
	h.data = h.data[:len(h.data)-1]

	return back
}

// PushItem is like heap.Push (required by the heap.Interface), but is type-safe. Will not panic if only this and
// PopItem are used.
func (h *Heap[T]) PushItem(item T) {
	heap.Push(h, item)
}

// PopMin is like heap.Pop (required by the heap.Interface), but is type-safe. Will not panic if only this and
// PUshItem are used.
func (h *Heap[T]) PopMin() T {
	return heap.Pop(h).(T)
}

func (h *Heap[T]) Contains(item T) bool {
	_, exists := h.exists[item]

	return exists
}

func (dir Direction) Opposite() Direction {
	switch dir {
	case DirectionNorth:
		return DirectionSouth
	case DirectionSouth:
		return DirectionNorth
	case DirectionEast:
		return DirectionWest
	case DirectionWest:
		return DirectionEast
	default:
		panic(fmt.Sprintf("invalid direction value %d", dir))
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
	grid, err := parseGrid(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(grid))

}

func part1(grid [][]int) int {
	heuristic := func(pos Coordinate) int {
		endRow := len(grid)
		endCol := len(grid[0])
		return abs(endRow-pos.Row) + abs(endCol-pos.Col)
	}

	startingLocation := Coordinate{
		Row:                 0,
		Col:                 0,
		FromDirection:       DirectionWest,
		NumMovesInDirection: 0,
	}
	estimatedDistances := map[Coordinate]int{
		startingLocation: heuristic(startingLocation),
	}

	shortestDistances := map[Coordinate]int{
		startingLocation: 0,
	}

	toVisit := NewHeap[Coordinate](func(c1, c2 Coordinate) int {
		estimatedC1Distance, haveC1Estimate := estimatedDistances[c1]
		estimatedC2Distance, haveC2Estimate := estimatedDistances[c2]
		if !haveC1Estimate && !haveC2Estimate {
			// both "infinity", but we will call them equal
			return 0
		} else if !haveC1Estimate {
			return 1
		} else if !haveC2Estimate {
			return -1
		}

		return estimatedC1Distance - estimatedC2Distance
	})

	toVisit.PushItem(startingLocation)
	parents := make(map[Coordinate]VisitedNode)

	firstVisit := true
	for toVisit.Len() > 0 {
		searchPos := toVisit.PopMin()
		if searchPos.Row == len(grid)-1 && searchPos.Col == len(grid[0])-1 {
			return shortestDistances[searchPos]
		}

		neighbors := neighbors(searchPos)
		for direction, neighborPos := range neighbors {
			if firstVisit {
				neighborPos.NumMovesInDirection = 0
			}

			if neighborPos.Row < 0 || neighborPos.Col < 0 || neighborPos.Row > len(grid)-1 || neighborPos.Col > len(grid[0])-1 {
				continue
			} else if searchPos.FromDirection == direction {
				// can't turn around
				continue
			} else if neighborPos.NumMovesInDirection > 2 {
				// can't go straight too much
				continue
			}

			toNeighbor := shortestDistances[searchPos] + grid[neighborPos.Row][neighborPos.Col]
			shortestToNeighbor, haveShortest := shortestDistances[neighborPos]
			if haveShortest && toNeighbor >= shortestToNeighbor {
				continue
			}

			shortestDistances[neighborPos] = toNeighbor
			estimatedDistances[neighborPos] = toNeighbor + heuristic(neighborPos)
			if !toVisit.Contains(neighborPos) {
				toVisit.PushItem(neighborPos)
			}
		}

		firstVisit = false
	}

	panic("search failed to find an element")
}

func neighbors(coordinate Coordinate) map[Direction]Coordinate {
	res := map[Direction]Coordinate{
		DirectionNorth: {Row: coordinate.Row - 1, Col: coordinate.Col, FromDirection: DirectionSouth},
		DirectionSouth: {Row: coordinate.Row + 1, Col: coordinate.Col, FromDirection: DirectionNorth},
		DirectionWest:  {Row: coordinate.Row, Col: coordinate.Col - 1, FromDirection: DirectionEast},
		DirectionEast:  {Row: coordinate.Row, Col: coordinate.Col + 1, FromDirection: DirectionWest},
	}

	inLastDirection := res[coordinate.FromDirection.Opposite()]
	inLastDirection.NumMovesInDirection = coordinate.NumMovesInDirection + 1
	res[coordinate.FromDirection.Opposite()] = inLastDirection

	return res
}

func parseGrid(inputLines []string) ([][]int, error) {
	grid := make([][]int, len(inputLines))
	for i, line := range inputLines {
		for _, char := range line {
			tileValue, err := strconv.Atoi(string(char))
			if err != nil {
				return nil, fmt.Errorf("invalid digit %c: %w", char, err)
			}

			grid[i] = append(grid[i], tileValue)
		}
	}

	return grid, nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}
