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

type MapRange struct {
	start int
	size  int
}

type ConversionMapEntry struct {
	destRange MapRange
	srcRange  MapRange
}

type ConversionMap []ConversionMapEntry

type ConvertsBetween struct {
	from string
	to   string
}

func (mapRange MapRange) Contains(n int) bool {
	return n >= mapRange.start && n < mapRange.start+mapRange.size
}

func (conversionMap ConversionMap) ConvertsTo(n int) int {
	for _, entry := range conversionMap {
		if entry.srcRange.Contains(n) {
			delta := n - entry.srcRange.start
			return entry.destRange.start + delta
		}
	}

	return n
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
	seeds, conversions, err := parseAlmanac(input)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(seeds, conversions))
}

func part1(seeds []int, conversions map[ConvertsBetween]ConversionMap) int {
	conversionSteps := []ConvertsBetween{
		{from: "seed", to: "soil"},
		{from: "soil", to: "fertilizer"},
		{from: "fertilizer", to: "water"},
		{from: "water", to: "light"},
		{from: "light", to: "temperature"},
		{from: "temperature", to: "humidity"},
		{from: "humidity", to: "location"},
	}

	min := math.MaxInt
	for _, seed := range seeds {
		item := seed
		for _, conversionStep := range conversionSteps {
			conversionMap, ok := conversions[conversionStep]
			if !ok {
				panic(fmt.Sprintf("Missing conversion for %s-to-%s", conversionStep.from, conversionStep.to))
			}

			item = conversionMap.ConvertsTo(item)
		}

		if item < min {
			min = item
		}
	}

	return min
}

func parseAlmanac(input string) ([]int, map[ConvertsBetween]ConversionMap, error) {
	sections := strings.Split(input, "\n\n")
	seeds, err := parseSeeds(sections[0])
	if err != nil {
		return nil, nil, fmt.Errorf("parse seeds: %w", err)
	}

	conversions := map[ConvertsBetween]ConversionMap{}
	for i, section := range sections[1:] {
		convertsBetween, conversionMap, err := parseConversionSection(section)
		if err != nil {
			return nil, nil, fmt.Errorf("parse section %d: %w", i, err)
		}

		conversions[convertsBetween] = conversionMap
	}

	return seeds, conversions, nil
}

func parseSeeds(seedSection string) ([]int, error) {
	stripped := strings.TrimPrefix(seedSection, "seeds: ")
	if stripped == seedSection {
		return nil, errors.New("missing seeds prefix")
	}

	rawSeedNumbers := strings.Split(stripped, " ")
	seedNumbers, err := tryParse(rawSeedNumbers, strconv.Atoi)
	if err != nil {
		return nil, fmt.Errorf("invalid seed numbers: %w", err)
	}

	return seedNumbers, nil
}

func parseConversionSection(section string) (ConvertsBetween, ConversionMap, error) {
	sectionLines := strings.Split(section, "\n")
	if len(sectionLines) < 2 {
		return ConvertsBetween{}, nil, errors.New("not enough information in section")
	}

	convertsBetween, err := parseSectionHeading(sectionLines[0])
	if err != nil {
		return ConvertsBetween{}, nil, fmt.Errorf("invalid section heading: %w", err)
	}

	conversionMap, err := parseConversionMap(sectionLines[1:])
	if err != nil {
		return ConvertsBetween{}, nil, fmt.Errorf("invalid conversion: %w", err)
	}

	return convertsBetween, conversionMap, nil
}

func parseSectionHeading(heading string) (ConvertsBetween, error) {
	pattern := regexp.MustCompile(`^(\w+)-to-(\w+) map:`)
	matches := pattern.FindStringSubmatch(heading)
	if matches == nil {
		return ConvertsBetween{}, errors.New("malformed heading")
	}

	return ConvertsBetween{
		from: matches[1],
		to:   matches[2],
	}, nil
}

func parseConversionMap(lines []string) (ConversionMap, error) {
	conversionMap := ConversionMap{}
	for _, line := range lines {
		rawEntryNumbers := strings.Split(line, " ")
		entryNumbers, err := tryParse(rawEntryNumbers, strconv.Atoi)
		if err != nil {
			return nil, fmt.Errorf("conversion map entry numbers: %w", err)
		} else if len(entryNumbers) != 3 {
			return nil, fmt.Errorf("expected 3 numbers in a conversion section, got %d", len(rawEntryNumbers))
		}

		dest := entryNumbers[0]
		src := entryNumbers[1]
		size := entryNumbers[2]

		entry := ConversionMapEntry{
			destRange: MapRange{start: dest, size: size},
			srcRange:  MapRange{start: src, size: size},
		}

		conversionMap = append(conversionMap, entry)
	}

	return conversionMap, nil
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
