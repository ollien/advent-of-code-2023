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

type Card struct {
	id             int
	winningNumbers []int
	ourNumbers     []int
}

func (card Card) NumMatchingNumbers() int {
	matchingNumbers := 0

	winningNumbers := makeSet(card.winningNumbers)
	for _, ourNumber := range card.ourNumbers {
		if _, ok := winningNumbers[ourNumber]; ok {
			matchingNumbers++
		}
	}

	return matchingNumbers
}

func (card Card) Score() int {
	numMatchingNumbers := card.NumMatchingNumbers()
	if numMatchingNumbers == 0 {
		return 0
	}

	return pow2(numMatchingNumbers)
}

func (card Card) WinsCardsWithIDs() []int {
	numMatchingNumbers := card.NumMatchingNumbers()

	wonCards := make([]int, numMatchingNumbers)
	for i := 0; i < numMatchingNumbers; i++ {
		wonCards[i] = card.id + i + 1
	}

	return wonCards
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
	fmt.Printf("Part 2: %d\n", part2(cards))
}

func part1(cards []Card) int {
	score := 0
	for _, card := range cards {
		score += card.Score()
	}

	return score
}

func part2(cards []Card) int {
	if len(cards) == 0 {
		return 0
	}

	cardsByID := map[int]Card{}
	for _, card := range cards {
		cardsByID[card.id] = card
	}

	// This solution is a bit naive; I didn't get particularly clever with
	// the number of cards we had and treated it like a tree problem (I could
	// have just made this faster by doing
	// visitedCardIds[wonCard] += visitedCardIds[card.id]
	// which would have been equivalent, and faster, but meh, I like this
	// solution even if it's slow)
	visitedCardIDs := map[int]int{}
	cardsInPlay := slices.Clone(cards)

	for len(cardsInPlay) > 0 {
		card := cardsInPlay[0]
		cardsInPlay = cardsInPlay[1:]

		visitedCardIDs[card.id]++

		wonCardIDs := card.WinsCardsWithIDs()
		for _, wonCardID := range wonCardIDs {
			cardsInPlay = append(cardsInPlay, cardsByID[wonCardID])
		}
	}

	totalCards := 0
	for _, numVisited := range visitedCardIDs {
		totalCards += numVisited
	}

	return totalCards
}

func parseCards(inputLines []string) ([]Card, error) {
	return tryParse(inputLines, parseCard)
}

func parseCard(inputLine string) (Card, error) {
	pattern := regexp.MustCompile(`^Card\s+(\d+): ((?:\s*\d+\s*?)+) \| ((?:\s*\d+\s*)+)$`)
	matches := pattern.FindStringSubmatch(inputLine)
	if matches == nil {
		return Card{}, errors.New("did not match line pattern")
	}

	id, err := strconv.Atoi(matches[1])
	if err != nil {
		return Card{}, fmt.Errorf("parse id: %w", err)
	}

	winningNumbers, err := parseCardNumbers(matches[2])
	if err != nil {
		return Card{}, fmt.Errorf("parse winning numbers: %w", err)
	}

	ourNumbers, err := parseCardNumbers(matches[3])
	if err != nil {
		return Card{}, fmt.Errorf("parse our numbers: %w", err)
	}

	return Card{
		id:             id,
		winningNumbers: winningNumbers,
		ourNumbers:     ourNumbers,
	}, nil
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

func pow2(exp int) int {
	if exp == 0 {
		return 1
	}

	res := 1
	for i := 1; i < exp; i++ {
		res *= 2
	}

	return res
}
