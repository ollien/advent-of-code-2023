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

type Card struct {
	winningNumbers []int
	ourNumbers     []int
}

func (card Card) Score() int {
	score := 0
	winningNumbers := makeSet(card.winningNumbers)
	for _, ourNumber := range card.ourNumbers {
		if _, ok := winningNumbers[ourNumber]; !ok {
			continue
		}

		if score == 0 {
			score = 1
		} else {
			score *= 2
		}
	}

	return score
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

	input := string(inputBytes)
	inputLines := strings.Split(strings.TrimSpace(input), "\n")
	cards, err := parseCards(inputLines)
	if err != nil {
		panic(fmt.Sprintf("could not parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(cards))
}

func part1(cards []Card) int {
	score := 0
	for _, card := range cards {
		score += card.Score()
	}

	return score
}

func parseCards(inputLines []string) ([]Card, error) {
	return tryParse(inputLines, parseCard)
}

func parseCard(inputLine string) (Card, error) {
	pattern := regexp.MustCompile(`^Card\s+\d+: ((?:\s*\d+\s*?)+) \| ((?:\s*\d+\s*)+)$`)
	matches := pattern.FindStringSubmatch(inputLine)
	if matches == nil {
		return Card{}, errors.New("did not match line pattern")
	}

	winningNumbers, err := parseCardNumbers(matches[1])
	if err != nil {
		return Card{}, fmt.Errorf("parse winning numbers: %w", err)
	}

	ourNumbers, err := parseCardNumbers(matches[2])
	if err != nil {
		return Card{}, fmt.Errorf("parse winning numbers: %w", err)
	}

	return Card{winningNumbers: winningNumbers, ourNumbers: ourNumbers}, nil
}

func normalizeSeparatingSpaces(s string) string {
	pattern := regexp.MustCompile(`\s{2,}`)

	return pattern.ReplaceAllString(s, " ")
}

func parseCardNumbers(numbers string) ([]int, error) {
	normalizedNumbers := normalizeSeparatingSpaces(numbers)
	trimmedNumbers := strings.TrimSpace(normalizedNumbers)
	splitNumbers := strings.Split(trimmedNumbers, " ")

	return tryParse(splitNumbers, strconv.Atoi)
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

func makeSet[T comparable, S ~[]T](items S) map[T]struct{} {
	set := map[T]struct{}{}
	for _, item := range items {
		set[item] = struct{}{}
	}

	return set
}
