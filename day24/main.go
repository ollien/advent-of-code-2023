package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Triplet struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Hailstone struct {
	Position Triplet `json:"position"`
	Velocity Triplet `json:"velocity"`
}

func (hailstone Hailstone) LineCoefficients() (float64, float64) {
	slope := hailstone.Velocity.Y / hailstone.Velocity.X
	intercept := hailstone.Position.Y - slope*hailstone.Position.X

	return slope, intercept
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
	hailstones, err := parseHailstones(inputLines)
	if err != nil {
		panic(fmt.Sprintf("invalid input: %s", err))
	}

	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(hailstones))
	fmt.Println("Part 2 can be solved using this JSON and the pysolve module")
	hailstoneJSON, err := json.MarshalIndent(hailstones, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("could not marshal hailstones: %s", err))
	}

	fmt.Println(string(hailstoneJSON))
}

func part1(hailstones []Hailstone) int {
	testMin := float64(200000000000000)
	testMax := float64(400000000000000)

	count := 0
	for i, h1 := range hailstones {
		for j := i + 1; j < len(hailstones); j++ {
			h2 := hailstones[j]

			x, y := intersectionPoint(h1, h2)

			x1DeltaSign := sign(x - h1.Position.X)
			y1DeltaSign := sign(y - h1.Position.Y)
			x2DeltaSign := sign(x - h2.Position.X)
			y2DeltaSign := sign(y - h2.Position.Y)

			h1InFuture := x1DeltaSign == sign(h1.Velocity.X) && y1DeltaSign == sign(h1.Velocity.Y)
			h2InFuture := x2DeltaSign == sign(h2.Velocity.X) && y2DeltaSign == sign(h2.Velocity.Y)
			inBounds := x <= testMax && x >= testMin && y <= testMax && y >= testMin
			finite := !math.IsInf(x, 0) && !math.IsInf(y, 0)

			if inBounds && finite && h1InFuture && h2InFuture {
				count++
			}
		}
	}

	return count
}

func intersectionPoint(h1, h2 Hailstone) (float64, float64) {
	h1Slope, h1Intercept := h1.LineCoefficients()
	h2Slope, h2Intercept := h2.LineCoefficients()

	// https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection#Given_two_line_equations
	x := (h2Intercept - h1Intercept) / (h1Slope - h2Slope)
	y := h1Slope*x + h1Intercept

	return x, y
}

func parseHailstones(inputLines []string) ([]Hailstone, error) {
	return tryParse(inputLines, parseHailstone)
}

func parseHailstone(inputLine string) (Hailstone, error) {
	linePattern := regexp.MustCompile(`^(-?\d+),\s*(-?\d+),\s*(-?\d+)\s*@\s*(-?\d+),\s*(-?\d+),\s*(-?\d+)$`)
	match := linePattern.FindStringSubmatch(inputLine)
	if match == nil {
		return Hailstone{}, errors.New("malformed hailstone")
	}

	matchedNumbers, err := tryParse(match[1:], strconv.Atoi)
	if err != nil {
		// Cannot happen, by the pattern
		panic(fmt.Sprintf("could not parse %v: %s", match[1:], err))
	}

	return Hailstone{
		Position: Triplet{
			X: float64(matchedNumbers[0]),
			Y: float64(matchedNumbers[1]),
			Z: float64(matchedNumbers[2]),
		},
		Velocity: Triplet{
			X: float64(matchedNumbers[3]),
			Y: float64(matchedNumbers[4]),
			Z: float64(matchedNumbers[5]),
		},
	}, nil
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

func sign(x float64) int {
	if x == 0 {
		return 0
	} else if x > 0 {
		return 1
	} else {
		return -1
	}
}
