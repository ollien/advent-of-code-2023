package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
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

type Range struct {
	// start and end are inclusive
	start Coordinate
	end   Coordinate
}

type DrawnPlan []Range
type Matrix [2][2]int

func NewMatrix(a, b, c, d int) Matrix {
	return [2][2]int{
		{a, b},
		{c, d},
	}
}

func NewRange(start Coordinate, direction Direction, count int) Range {
	end := inDirection(start, direction, count)

	return Range{
		start: start,
		end:   end,
	}
}

func (mat Matrix) Det() int {
	return mat[0][0]*mat[1][1] - mat[0][1]*mat[1][0]
}

func (r Range) Start() Coordinate {
	return r.start
}

func (r Range) End() Coordinate {
	return r.end
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
	verts, err := findVerts(drawn)
	if err != nil {
		panic(err)
	}

	area := 0
	border := 0
	for i := 0; i < len(verts); i++ {
		point1 := verts[i]
		point2 := verts[(i+1)%len(verts)]

		mat := NewMatrix(point1.Col, point2.Col, point1.Row, point2.Row)
		area += mat.Det()
		border += abs(point2.Row-point1.Row) + abs(point2.Col-point1.Col)
	}

	// Why is this plus one necessary???
	return (area+border)/2 + 1
}

func drawPlans(plans []Plan) DrawnPlan {
	cursor := Coordinate{Row: 0, Col: 0}
	drawn := DrawnPlan{}
	for _, plan := range plans {
		r := NewRange(cursor, plan.Direction, plan.Count)
		drawn = append(drawn, r)
		cursor = r.End()
	}

	return drawn
}

func findVerts(drawn DrawnPlan) ([]Coordinate, error) {
	startingPoint := drawn[0].Start()
	cursor := drawn[0]
	verts := []Coordinate{}
	for {
		idx := slices.IndexFunc(drawn, func(r Range) bool {
			return r.Start() == cursor.End()
		})

		if idx == -1 {
			return nil, errors.New("input is not a loop")
		}

		cursor = drawn[idx]
		verts = append(verts, cursor.Start())

		if cursor.Start() == startingPoint {
			return verts, nil
		}
	}
}

func inDirection(coordinate Coordinate, direction Direction, n int) Coordinate {
	switch direction {
	case DirectionUp:
		return Coordinate{Row: coordinate.Row - n, Col: coordinate.Col}
	case DirectionDown:
		return Coordinate{Row: coordinate.Row + n, Col: coordinate.Col}
	case DirectionLeft:
		return Coordinate{Row: coordinate.Row, Col: coordinate.Col - n}
	case DirectionRight:
		return Coordinate{Row: coordinate.Row, Col: coordinate.Col + n}
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

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}
