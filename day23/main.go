// This code is horrible and repetitive, I'm sorry

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

type GraphNode struct {
	Position Coordinate
	Weight   int
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

	fmt.Printf("Part 1: %d\n", part1(grid))
	fmt.Printf("Part 2: %d\n", part2(grid))
}

func part1(grid [][]Tile) int {
	return solve(grid, true)
}

func part2(grid [][]Tile) int {
	return solve(grid, false)
}

func solve(grid [][]Tile, respectSlopes bool) int {
	startCol, err := findStartingTile(grid[0])
	if err != nil {
		panic(fmt.Sprintf("could not find starting tile: %s", err))
	}

	endCol, err := findStartingTile(grid[len(grid)-1])
	if err != nil {
		panic(fmt.Sprintf("could not find starting tile: %s", err))
	}

	graph := buildCondensedGraph(
		Coordinate{Row: 0, Col: startCol},
		Coordinate{Row: len(grid) - 1, Col: endCol},
		grid,
		respectSlopes,
	)

	return findLongestPath(
		Coordinate{Row: 0, Col: startCol},
		Coordinate{Row: len(grid) - 1, Col: endCol},
		grid,
		graph,
	)
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

func buildCondensedGraph(start, end Coordinate, grid [][]Tile, respectSlopes bool) map[Coordinate][]GraphNode {
	inBounds := func(pos Coordinate) bool {
		return pos.Row >= 0 && pos.Row < len(grid) && pos.Col >= 0 && pos.Col < len(grid[0])
	}

	toVisit := []Coordinate{start}
	visited := map[Coordinate]struct{}{}
	distances := map[Coordinate]int{
		start: 0,
	}

	intersections := map[Coordinate]struct{}{start: {}, end: {}}
	for len(toVisit) > 0 {
		visiting := toVisit[0]
		toVisit = toVisit[1:]
		visited[visiting] = struct{}{}

		validNeighbors := []Coordinate{}
		for _, neighbor := range findNeighbors(grid, visiting, respectSlopes) {
			if !inBounds(neighbor) {
				continue
			} else if grid[neighbor.Row][neighbor.Col] == TileWall {
				continue
			}

			distances[neighbor] = distances[visiting] + 1
			validNeighbors = append(validNeighbors, neighbor)
		}

		if len(validNeighbors) > 2 {
			intersections[visiting] = struct{}{}
		}

		for _, neighbor := range validNeighbors {
			if neighbor == end {
				intersections[neighbor] = struct{}{}
			}

			if _, ok := visited[neighbor]; !ok {
				toVisit = append(
					toVisit,
					neighbor,
				)
			}
		}
	}

	res := map[Coordinate][]GraphNode{}
	for intersection := range intersections {
		toVisit := []Coordinate{intersection}
		visited := map[Coordinate]struct{}{}
		distances := map[Coordinate]int{
			intersection: 0,
		}

		for len(toVisit) > 0 {
			visiting := toVisit[0]
			toVisit = toVisit[1:]
			visited[visiting] = struct{}{}
			for _, neighbor := range findNeighbors(grid, visiting, respectSlopes) {
				if !inBounds(neighbor) {
					continue
				} else if grid[neighbor.Row][neighbor.Col] == TileWall {
					continue
				}

				distances[neighbor] = distances[visiting] + 1
				if _, ok := intersections[neighbor]; ok && neighbor != intersection {
					res[intersection] = append(
						res[intersection],
						GraphNode{Position: neighbor, Weight: distances[neighbor]},
					)
				} else if _, ok := visited[neighbor]; !ok {
					toVisit = append(toVisit, neighbor)
				}
			}
		}
	}

	return res
}

func findLongestPath(start Coordinate, end Coordinate, grid [][]Tile, graph map[Coordinate][]GraphNode) int {
	var dfs func(Coordinate, []GraphNode) []GraphNode
	dfs = func(coordinate Coordinate, path []GraphNode) []GraphNode {
		children := graph[coordinate]
		longestPath := path
		for _, child := range children {
			if slices.ContainsFunc(path, func(node GraphNode) bool { return node.Position == child.Position }) {
				continue
			}

			nextPath := slices.Clone(path)
			nextPath = append(nextPath, child)
			fullPath := dfs(child.Position, nextPath)
			if sumWeights(fullPath) > sumWeights(longestPath) && fullPath[len(fullPath)-1].Position == end {
				longestPath = fullPath
			}
		}

		return longestPath
	}

	res := dfs(start, []GraphNode{})

	return sumWeights(res)
}

func sumWeights(nodes []GraphNode) int {
	total := 0
	for _, node := range nodes {
		total += node.Weight
	}

	return total
}

func findNeighbors(grid [][]Tile, position Coordinate, respectSlopes bool) []Coordinate {
	upNeighbor := Coordinate{Row: position.Row - 1, Col: position.Col}
	downNeighbor := Coordinate{Row: position.Row + 1, Col: position.Col}
	leftNeighbor := Coordinate{Row: position.Row, Col: position.Col - 1}
	rightNeighbor := Coordinate{Row: position.Row, Col: position.Col + 1}

	if respectSlopes && grid[position.Row][position.Col] == TileLeft {
		return []Coordinate{leftNeighbor}
	} else if respectSlopes && grid[position.Row][position.Col] == TileRight {
		return []Coordinate{rightNeighbor}
	} else if respectSlopes && grid[position.Row][position.Col] == TileUp {
		return []Coordinate{upNeighbor}
	} else if respectSlopes && grid[position.Row][position.Col] == TileDown {
		return []Coordinate{downNeighbor}
	}

	return []Coordinate{upNeighbor, downNeighbor, leftNeighbor, rightNeighbor}
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
