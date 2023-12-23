package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Coordinate struct {
	X int
	Y int
	Z int
}

type Brick []Coordinate

// BrickGraph is a map from indexes in a brick array to the dependent indexes
type BrickGraph map[int][]int

func (graph BrickGraph) ReachableFrom(idx int) map[int]struct{} {
	return graph.ReachableFromExcluding(idx, nil)
}

// ReachableFromExcluding will finds all nodes reachable from the given node index, but will not explore neighbors
// in the "excluding" set.
func (graph BrickGraph) ReachableFromExcluding(idx int, excluding map[int]struct{}) map[int]struct{} {
	visited := map[int]struct{}{}
	toVisit := []int{idx}
	for len(toVisit) > 0 {
		visiting := toVisit[0]
		toVisit = toVisit[1:]
		for _, neighbor := range graph[visiting] {
			if _, ok := visited[neighbor]; ok {
				continue
			} else if _, ok := excluding[neighbor]; ok {
				continue
			}

			visited[neighbor] = struct{}{}
			toVisit = append(toVisit, neighbor)
		}
	}

	return visited
}

func (b Brick) LowestPoint() Coordinate {
	minZFunc := func(a, b Coordinate) int {
		return cmp.Compare(a.Z, b.Z)
	}

	return slices.MinFunc(b, minZFunc)
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
	bricks, err := parseBricks(inputLines)
	if err != nil {
		panic(fmt.Sprintf("invalid input: %s", err))
	}

	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(bricks))
	fmt.Printf("Part 2: %d\n", part2(bricks))
}

func part1(inputBricks []Brick) int {
	slammedBricks := settleBricks(inputBricks)
	incoming, outgoing := buildBrickGraph(slammedBricks)
	removable := removableBricks(slammedBricks, incoming, outgoing)

	return len(removable)
}

func part2(inputBricks []Brick) int {
	slammedBricks := settleBricks(inputBricks)
	incoming, outgoing := buildBrickGraph(slammedBricks)
	total := 0
	for i := range slammedBricks {
		stableNodes := map[int]struct{}{}
		lastFalling := map[int]struct{}{}
		for {
			reachableNodes := outgoing.ReachableFromExcluding(i, stableNodes)
			for reachable := range reachableNodes {
				for _, parentOfReachable := range incoming[reachable] {
					if _, ok := reachableNodes[parentOfReachable]; !ok && parentOfReachable != i {
						// If any reachable node is accessible from another subgraph, it is "stable"
						stableNodes[reachable] = struct{}{}
					}
				}
			}

			falling := outgoing.ReachableFromExcluding(i, stableNodes)
			shouldBreak := mapKeysEqual(falling, lastFalling)
			lastFalling = falling
			if shouldBreak {
				break
			}

			// We must repeat this process until we reach a state where no more stable nodes are found
			// There are some cases where a node might be stable, but the children of said stable node
			// must also be considered invalidated (think of it as second-order stability)
		}

		total += len(lastFalling)
	}

	return total
}

func mapKeysEqual[T comparable, U any](m1, m2 map[T]U) bool {
	if len(m1) != len(m2) {
		return false
	}

	for key := range m1 {
		if _, ok := m2[key]; !ok {
			return false
		}
	}

	return true
}

func settleBricks(bricks []Brick) []Brick {
	sorted := slices.Clone(bricks)
	sortByHeight(sorted)

	slammedBricks := slices.Clone(sorted)
	for i := range sorted {
		brick, err := moveBrickDown(slammedBricks, i)
		if err != nil {
			// can't happen with our bounds
			panic(err)
		}

		slammedBricks[i] = brick
	}

	return slammedBricks
}

func moveBrickDown(bricks []Brick, brickIdx int) (Brick, error) {
	if brickIdx < 0 || brickIdx >= len(bricks) {
		return nil, fmt.Errorf("invalid brick index %d", brickIdx)
	}

	occupied := occupiedPositions(bricks)

	brick := slices.Clone(bricks[brickIdx])
	for brick.LowestPoint().Z > 1 {
		nextBrick := slices.Clone(brick)
		for i, pos := range nextBrick {
			nextBrick[i] = Coordinate{X: pos.X, Y: pos.Y, Z: pos.Z - 1}
			if idx, ok := occupied[nextBrick[i]]; ok && idx != brickIdx {
				return brick, nil
			}
		}

		brick = nextBrick
	}

	return brick, nil
}

func buildBrickGraph(bricks []Brick) (incoming, outgoing BrickGraph) {
	occupied := occupiedPositions(bricks)
	outgoing = make(BrickGraph)
	incoming = make(BrickGraph)

	for i, brick := range bricks {
		neighboring := map[int]struct{}{}
		for _, block := range brick {
			above := Coordinate{X: block.X, Y: block.Y, Z: block.Z + 1}
			if occupiedBy, ok := occupied[above]; ok && occupiedBy != i {
				neighboring[occupiedBy] = struct{}{}
			}
		}

		for neighbor := range neighboring {
			outgoing[i] = append(outgoing[i], neighbor)
			incoming[neighbor] = append(incoming[neighbor], i)
		}
	}

	return
}

func removableBricks(allBricks []Brick, incoming BrickGraph, outgoing BrickGraph) []int {
	removable := []int{}
	for i := range allBricks {
		dependents := outgoing[i]
		allDependentsSafe := true
		for _, dependent := range dependents {
			// There is more than one item which has this dependent as a dependent, so removing i would
			// not allow this to fall
			if len(incoming[dependent]) <= 1 {
				allDependentsSafe = false
				break
			}
		}

		if allDependentsSafe {
			removable = append(removable, i)
		}

	}

	return removable
}

func sortByHeight(bricks []Brick) {
	slices.SortFunc(bricks, func(brick1, brick2 Brick) int {
		min1Z := brick1.LowestPoint()
		min2Z := brick2.LowestPoint()

		return cmp.Compare(min1Z.Z, min2Z.Z)
	})
}

func occupiedPositions(bricks []Brick) map[Coordinate]int {
	occupied := map[Coordinate]int{}
	for i, brick := range bricks {
		for _, pos := range brick {
			occupied[pos] = i
		}
	}

	return occupied
}

func parseBricks(inputLines []string) ([]Brick, error) {
	bricks, err := tryParse(inputLines, parseBrick)
	if err != nil {
		return nil, fmt.Errorf("parse brick: %s", err)
	}

	positions := map[Coordinate]struct{}{}
	for _, brick := range bricks {
		for _, pos := range brick {
			if _, ok := positions[pos]; ok {
				return nil, fmt.Errorf("bricks overlap at %+v", pos)
			}

			positions[pos] = struct{}{}
		}
	}

	return bricks, nil
}

func parseBrick(line string) (Brick, error) {
	pattern := regexp.MustCompile(`^(\d+),(\d+),(\d+)~(\d+),(\d+),(\d+)$`)
	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, errors.New("malformed brick spec")
	}

	coordSlice1, err := tryParse([]string{matches[1], matches[2], matches[3]}, strconv.Atoi)
	if err != nil {
		// Can't happen, by the pattern
		panic(fmt.Sprintf("could not convert coordinate to integers: %s", err))
	}

	coordSlice2, err := tryParse([]string{matches[4], matches[5], matches[6]}, strconv.Atoi)
	if err != nil {
		// Can't happen, by the pattern
		panic(fmt.Sprintf("could not convert coordinate to integers: %s", err))
	}

	numDifferent := countDifferent(coordSlice1, coordSlice2)
	if numDifferent > 1 {
		return nil, fmt.Errorf("only one axis may differ in coordinates, found %d", numDifferent)
	}

	brick := make(Brick, 0, 3)
	for x := coordSlice1[0]; x <= coordSlice2[0]; x++ {
		for y := coordSlice1[1]; y <= coordSlice2[1]; y++ {
			for z := coordSlice1[2]; z <= coordSlice2[2]; z++ {
				brick = append(brick, Coordinate{X: x, Y: y, Z: z})
			}
		}
	}

	return brick, nil
}

func countDifferent[T comparable, S ~[]T](s1, s2 S) int {
	if len(s1) != len(s2) {
		panic("cannot compare lists of different lengths")
	}

	count := 0
	for i, item := range s1 {
		if item != s2[i] {
			count++
		}
	}

	return count
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
