package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Direction int

const (
	DirectionUp = iota
	DirectionDown
	DirectionLeft
	DirectionRight
)

type Plan struct {
	Direction Direction
	Count     int
	ColorCode string
}

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
	plans, err := parsePlans(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(plans))
}

func part1(plans []Plan) int {
	if len(plans) == 0 {
		panic("cannot draw no plans!")
	}

	drawn := drawPlans(plans)
	emptySpaces := emptySpacesInDrawing(drawn)
	for len(emptySpaces) > 0 {
		empty := popMapKey(emptySpaces)
		seen, outside := flood(drawn, empty)
		if outside {
			for _, pos := range seen {
				delete(emptySpaces, pos)
			}
		} else {
			return len(seen) + len(drawn)
		}
	}

	panic("found no interior items")
}

func drawPlans(plans []Plan) map[Coordinate]string {
	cursor := Coordinate{Row: 0, Col: 0}
	drawn := map[Coordinate]string{}
	for _, plan := range plans {
		for i := 0; i < plan.Count; i++ {
			cursor = inDirection(cursor, plan.Direction)
			drawn[cursor] = plan.ColorCode
		}
	}

	return drawn
}

func drawingBounds(drawn map[Coordinate]string) (minRow, maxRow, minCol, maxCol int) {
	if len(drawn) == 0 {
		panic("cannot get bounds of empty drawing")
	}

	minRow = math.MaxInt
	minCol = math.MaxInt
	for position := range drawn {
		minRow = min(position.Row, minRow)
		minCol = min(position.Col, minCol)
		maxRow = max(position.Row, maxRow)
		maxCol = max(position.Col, maxCol)
	}

	return
}

func emptySpacesInDrawing(drawn map[Coordinate]string) map[Coordinate]struct{} {
	minRow, maxRow, minCol, maxCol := drawingBounds(drawn)
	emptySpaces := map[Coordinate]struct{}{}
	for row := minRow; row <= maxRow; row++ {
		for col := minCol; col <= maxCol; col++ {
			position := Coordinate{Row: row, Col: col}
			_, ok := drawn[position]
			if !ok {
				emptySpaces[position] = struct{}{}
			}
		}
	}

	return emptySpaces
}

// flood is a floodFill algorithm that will return the empty tiles flooded, and whether or not the exterior
// border was hit.
func flood(drawn map[Coordinate]string, seed Coordinate) (flooded []Coordinate, outside bool) {
	if _, ok := drawn[seed]; ok {
		// We can't flood a fille space
		return []Coordinate{}, true
	}

	outsideHit := false
	minRow, maxRow, minCol, maxCol := drawingBounds(drawn)
	visited := map[Coordinate]struct{}{}
	emptyVisited := []Coordinate{}
	toVisit := []Coordinate{seed}
	for len(toVisit) > 0 {
		visiting := toVisit[0]
		toVisit = toVisit[1:]
		if _, ok := visited[visiting]; ok {
			continue
		}

		visited[visiting] = struct{}{}
		emptyVisited = append(emptyVisited, visiting)

		for _, neighbor := range neighbors(visiting) {
			if _, ok := drawn[neighbor]; ok {
				continue
			} else if neighbor.Row < minRow || neighbor.Row > maxRow || neighbor.Col < minCol || neighbor.Col > maxCol {
				outsideHit = true
				continue
			}

			toVisit = append(toVisit, neighbor)
		}
	}

	return emptyVisited, outsideHit
}

func neighbors(coordinate Coordinate) []Coordinate {
	return []Coordinate{
		{Row: coordinate.Row + 1, Col: coordinate.Col},
		{Row: coordinate.Row - 1, Col: coordinate.Col},
		{Row: coordinate.Row, Col: coordinate.Col + 1},
		{Row: coordinate.Row, Col: coordinate.Col - 1},
	}
}

func inDirection(coordinate Coordinate, direction Direction) Coordinate {
	switch direction {
	case DirectionUp:
		return Coordinate{Row: coordinate.Row - 1, Col: coordinate.Col}
	case DirectionDown:
		return Coordinate{Row: coordinate.Row + 1, Col: coordinate.Col}
	case DirectionLeft:
		return Coordinate{Row: coordinate.Row, Col: coordinate.Col - 1}
	case DirectionRight:
		return Coordinate{Row: coordinate.Row, Col: coordinate.Col + 1}
	default:
		panic(fmt.Sprintf("invalid direction %d", direction))
	}
}

func parsePlans(inputLines []string) ([]Plan, error) {
	return tryParse(inputLines, parsePlan)
}

func parsePlan(rawPlan string) (Plan, error) {
	pattern := regexp.MustCompile(`^([RUDL]) (\d+) \(#([a-z0-f]+)\)`)
	matches := pattern.FindStringSubmatch(rawPlan)
	if matches == nil {
		return Plan{}, fmt.Errorf("malformed pattern")
	}

	direction, err := directionFromAcronym(matches[1])
	if err != nil {
		return Plan{}, fmt.Errorf("invalid direction %s: %w", matches[1], err)
	}

	count, err := strconv.Atoi(matches[2])
	if err != nil {
		// Can't happen by the expression pattern
		panic(fmt.Sprintf("failed to parse %s as number: %w", matches[2], err))
	}

	return Plan{
		Direction: direction,
		Count:     count,
		ColorCode: matches[3],
	}, nil
}

func directionFromAcronym(n string) (Direction, error) {
	switch n {
	case "R":
		return DirectionRight, nil
	case "U":
		return DirectionUp, nil
	case "D":
		return DirectionDown, nil
	case "L":
		return DirectionLeft, nil
	default:
		return DirectionUp, errors.New("invalid direction acronym")
	}
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

func mapKeys[T comparable, U any](m map[T]U) []T {
	res := make([]T, 0, len(m))
	for key := range m {
		res = append(res, key)
	}

	return res
}

func popMapKey[T comparable, U any](m map[T]U) T {
	for key := range m {
		delete(m, key)
		return key
	}

	panic("cannot pop empty map")
}
