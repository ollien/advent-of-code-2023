package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type CubeCounts = map[string]int

type Game struct {
	id     int
	rounds []CubeCounts
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s inputfile\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	inputFile, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("failed to open input file: %s", err))
	}

	inputBytes, err := io.ReadAll(inputFile)
	if err != nil {
		panic(fmt.Sprintf("failed to read input file: %s", err))
	}

	input := string(inputBytes)
	inputLines := strings.Split(strings.TrimSpace(input), "\n")
	games, err := parseGames(inputLines)
	if err != nil {
		panic(fmt.Sprintf("invalid input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(games))
	fmt.Printf("Part 2: %d\n", part2(games))
}

func part1(games []Game) int {
	invalidIDTotal := 0
	for _, game := range games {
		if isGameValid(game) {
			invalidIDTotal += game.id
		}
	}

	return invalidIDTotal
}

func part2(games []Game) int {
	totalPower := 0
	for _, game := range games {
		minPossibleCubes := maxCubesByColor(game)
		totalPower += cubePower(minPossibleCubes)
	}

	return totalPower
}

// isGameValid will check if the given game is valid by the number of cubes in the bag
func isGameValid(game Game) bool {
	cubeCounts := map[string]int{
		"red":   12,
		"green": 13,
		"blue":  14,
	}

	for _, round := range game.rounds {
		for color, count := range round {
			if count > cubeCounts[color] {
				return false
			}
		}
	}

	return true
}

// maxCubesByColor will get the maximum quantity of cubes for each color in the game's rounds
func maxCubesByColor(game Game) CubeCounts {
	maxCounts := CubeCounts{}
	for _, round := range game.rounds {
		for color, count := range round {
			currentMax := maxCounts[color]
			if count > currentMax {
				maxCounts[color] = count
			}
		}
	}

	return maxCounts
}

// cubePower calculates the "power" of the cubes selected in a round
func cubePower(cubes CubeCounts) int {
	return cubes["red"] * cubes["green"] * cubes["blue"]
}

func parseGames(inputLines []string) ([]Game, error) {
	return tryParse[Game](inputLines, parseGame)
}

func parseGame(line string) (Game, error) {
	gameID, chosenCubes, err := splitGameLine(line)
	if err != nil {
		return Game{}, fmt.Errorf("invalid game %q: %w", line, err)
	}

	rounds, err := parseRounds(chosenCubes)
	if err != nil {
		return Game{}, fmt.Errorf("invalid rounds %q: %w", chosenCubes, err)
	}

	return Game{
		id:     gameID,
		rounds: rounds,
	}, nil
}

func splitGameLine(line string) (int, string, error) {
	gamesPattern := regexp.MustCompile(`^Game (\d+): (.*)$`)
	matches := gamesPattern.FindStringSubmatch(line)
	if matches == nil {
		return 0, "", errors.New("malformed game line")
	}

	rawGameID := matches[1]
	gameID, err := strconv.Atoi(rawGameID)
	if err != nil {
		// we already know this isn't going to happen from the regexp above
		panic("game id was non-numeric")
	}

	return gameID, matches[2], nil
}

func parseRounds(roundsSpec string) ([]CubeCounts, error) {
	roundSpecs := strings.Split(roundsSpec, ";")
	return tryParse[CubeCounts](roundSpecs, parseRound)
}

func parseRound(round string) (CubeCounts, error) {
	roundPattern := regexp.MustCompile(`(\d+) (\w+)`)
	matches := roundPattern.FindAllStringSubmatch(round, -1)
	if matches == nil {
		return nil, errors.New("malformed round spec")
	}

	cubes := CubeCounts{}
	for _, match := range matches {
		color := match[2]
		rawCubeCount := match[1]
		count, err := strconv.Atoi(rawCubeCount)
		if err != nil {
			// we already know this isn't going to happen from the regexp above
			panic("cube count was non-numeric")
		}

		cubes[color] = count
	}

	return cubes, nil
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
