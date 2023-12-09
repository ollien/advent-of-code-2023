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

type Race struct {
	time           int
	recordDistance int
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
	races, err := parseRaces(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse races: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(races))
	fmt.Printf("Part 2: %d\n", part2(races))
}

func part1(races []Race) int {
	res := 1
	for _, race := range races {
		res *= numberOfWaysToWinRace(race)
	}

	return res
}

func part2(races []Race) int {
	bigRace, err := combineRaces(races)
	if err != nil {
		panic(err)
	}

	return numberOfWaysToWinRace(bigRace)
}

func numberOfWaysToWinRace(race Race) int {
	possibleWins := 0
	for timeHeld := 0; timeHeld <= race.time; timeHeld++ {
		distance := distanceForTimeHeld(timeHeld, race.time)
		if distance > race.recordDistance {
			possibleWins++
		}
	}

	return possibleWins
}

func distanceForTimeHeld(buttonHeld int, raceTime int) int {
	timeLeft := raceTime - buttonHeld
	speed := buttonHeld

	return speed * timeLeft
}

func parseRaces(lines []string) ([]Race, error) {
	if len(lines) != 2 {
		return nil, errors.New("input must be exactly two lines")
	}

	timeLine := removeRowPrefix(lines[0], "Time")
	distanceLine := removeRowPrefix(lines[1], "Distance")

	if timeLine == lines[0] {
		return nil, errors.New("malformed 'time' line")
	} else if distanceLine == lines[1] {
		return nil, errors.New("malformed 'distance' line")
	}

	spacePattern := regexp.MustCompile(`\s+`)
	rawTimeLineComponents := spacePattern.Split(timeLine, -1)
	rawDistanceLineComponents := spacePattern.Split(distanceLine, -1)
	if len(rawTimeLineComponents) != len(rawDistanceLineComponents) {
		return nil, errors.New("'time' and 'distance' lines have a different number of elements")
	}

	timeLineComponents, err := tryParse(rawTimeLineComponents, strconv.Atoi)
	if err != nil {
		return nil, fmt.Errorf("invalid element in 'time' line: %w", err)
	}

	distanceLineComponents, err := tryParse(rawDistanceLineComponents, strconv.Atoi)
	if err != nil {
		return nil, fmt.Errorf("invalid element in 'distance' line: %w", err)
	}

	races := make([]Race, len(timeLineComponents))
	for i := 0; i < len(timeLineComponents); i++ {
		timeComponent := timeLineComponents[i]
		distanceComponent := distanceLineComponents[i]
		races[i] = Race{
			time:           timeComponent,
			recordDistance: distanceComponent,
		}
	}

	return races, nil
}

func removeRowPrefix(s string, prefix string) string {
	pattern := regexp.MustCompile("^" + regexp.QuoteMeta(prefix) + `:\s*`)

	return pattern.ReplaceAllLiteralString(s, "")
}

func combineRaces(races []Race) (Race, error) {
	if len(races) == 0 {
		return Race{}, errors.New("cannot combine zero races into one")
	}

	raceTimes := make([]int, len(races))
	raceRecords := make([]int, len(races))
	for i, race := range races {
		raceTimes[i] = race.time
		raceRecords[i] = race.recordDistance
	}

	bigRaceTime := smashNumbers(raceTimes)
	bigRaceRecord := smashNumbers(raceRecords)

	return Race{
		time:           bigRaceTime,
		recordDistance: bigRaceRecord,
	}, nil
}

func smashNumbers(nums []int) int {
	if len(nums) == 0 {
		// programmer error
		panic("cannot combine zero numbers into a big one")
	}

	s := ""
	for _, n := range nums {
		s += strconv.Itoa(n)
	}

	bigNum, err := strconv.Atoi(s)
	if err != nil {
		// should never fail, given we only use numbers as is
		panic(fmt.Sprintf("converting %s to a number failed: %w", s, err))
	}

	return bigNum
}

func tryParse[T any](items []string, doParse func(s string) (T, error)) ([]T, error) {
	res := []T{}
	for i, line := range items {
		parsedItem, err := doParse(line)
		if err != nil {
			return nil, fmt.Errorf("malformed item at index %d: %w", i, err)
		}

		res = append(res, parsedItem)
	}

	return res, nil
}
