package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Card int

const (
	Joker Card = iota + 1
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

type HandKind int

const (
	UnknownKind HandKind = iota
	HighCard
	OnePair
	TwoPair
	ThreeOfAKind
	FullHouse
	FourOfAKind
	FiveOfAKind
)

type Hand []Card

type Player struct {
	bid  int
	hand Hand
}

// Kind gets the "Kind" of the hand, which determines its value.
func (hand Hand) Kind() (HandKind, error) {
	type kindMapping struct {
		kind           HandKind
		distinctCounts []int
	}

	// Must iterate in order so we try each kind first
	handCounts := []kindMapping{
		{
			kind:           FiveOfAKind,
			distinctCounts: []int{5},
		},
		{
			kind:           FourOfAKind,
			distinctCounts: []int{1, 4},
		},
		{
			kind:           FullHouse,
			distinctCounts: []int{2, 3},
		},
		{
			kind:           ThreeOfAKind,
			distinctCounts: []int{1, 1, 3},
		},
		{
			kind:           TwoPair,
			distinctCounts: []int{1, 2, 2},
		},
		{
			kind:           OnePair,
			distinctCounts: []int{1, 1, 1, 2},
		},
		{
			kind:           HighCard,
			distinctCounts: []int{1, 1, 1, 1, 1},
		},
	}

	handWithoutJokers, numJokers := hand.WithoutCard(Joker)
	distinctCounts := handWithoutJokers.CountDistinct()
	cardCounts := mapValues(distinctCounts)
	slices.Sort(cardCounts)
	for _, mapping := range handCounts {
		if canMakeKindWithCardCombo(cardCounts, mapping.distinctCounts, numJokers) {
			return mapping.kind, nil
		}
	}

	return UnknownKind, errors.New("no known kind for hand")
}

// WithoutCard will return a copy of the Hand without the given card, toRemove
func (hand Hand) WithoutCard(toRemove Card) (Hand, int) {
	newHand := Hand{}
	numRemoved := 0
	for _, card := range hand {
		if card == toRemove {
			numRemoved++
			continue
		}

		newHand = append(newHand, card)
	}

	return newHand, numRemoved
}

// CountDistinct returns a count of distinct cards by each card type
func (hand Hand) CountDistinct() map[Card]int {
	buckets := make(map[Card]int)
	for _, card := range hand {
		buckets[card]++
	}

	return buckets
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
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	inputLines := strings.Split(input, "\n")
	players, err := parsePlayers(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse races: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(players))
	fmt.Printf("Part 2: %d\n", part2(players))
}

func part1(players []Player) int {
	return findWinnings(players)
}

func part2(originalPlayers []Player) int {
	players := makePart2Players(originalPlayers)

	return findWinnings(players)
}

// findWinnings finds the winning for each game
func findWinnings(players []Player) int {
	sortedPlayers := slices.Clone(players)
	slices.SortFunc(sortedPlayers, func(a, b Player) int {
		aKind, err := a.hand.Kind()
		if err != nil {
			panic(fmt.Errorf("invalid hand %+v: %w", a, err))
		}

		bKind, err := b.hand.Kind()
		if err != nil {
			panic(fmt.Errorf("invalid hand %+v: %w", b, err))
		}

		compareHands := cmp.Compare(aKind, bKind)
		if compareHands == 0 {
			return slices.Compare(a.hand, b.hand)
		} else {
			return compareHands
		}
	})

	winnings := 0
	for i, player := range sortedPlayers {
		winnings += (i + 1) * player.bid
	}

	return winnings
}

// makePart2Players prepares the players for part 2 by replacing Jacks with Jokers
func makePart2Players(players []Player) []Player {
	newPlayers := slices.Clone(players)
	for i, player := range newPlayers {
		updPlayer := player
		updPlayer.hand = makePart2Hand(player.hand)
		newPlayers[i] = updPlayer
	}

	return newPlayers
}

// makePart2Hand makes replaces Jacks with Jokers in each hand for part 2
func makePart2Hand(hand Hand) Hand {
	newHand := slices.Clone(hand)
	for i, card := range newHand {
		if card == Jack {
			newHand[i] = Joker
		}
	}
	return newHand
}

// canMakeKindWithCardCombo checks that if we can make the kind given. This is done by taking a sorted list of
// each of the number of each type of card. For instance, 55766 would be [2, 2, 1]
func canMakeKindWithCardCombo(handDistinctCardCounts []int, kindCardCounts []int, numJokers int) bool {
	if numJokers == 0 {
		return slices.Equal(kindCardCounts, handDistinctCardCounts)
	} else if len(handDistinctCardCounts) == 0 && numJokers == 5 && slices.Equal(kindCardCounts, []int{5}) {
		// If we have five jokers, we must act as a five of a kind (most valuable)
		return true
	}
	permutations := permutationsOfJokers(handDistinctCardCounts, numJokers)

	for _, permutation := range permutations {
		if slices.Equal(permutation, kindCardCounts) {
			return true
		}
	}

	return false
}

// Find all the positions where our sets of jokers could be located in in the distinct card count set
func permutationsOfJokers(distinctCardCounts []int, numJokers int) [][]int {
	if numJokers == 0 {
		return [][]int{distinctCardCounts}
	}

	permutations := [][]int{}
	for i, count := range distinctCardCounts {
		newCounts := slices.Clone(distinctCardCounts)
		newCounts[i] = count + 1
		permutations = append(permutations, permutationsOfJokers(newCounts, numJokers-1)...)
	}

	return permutations
}

func parsePlayers(inputLines []string) ([]Player, error) {
	players := make([]Player, 0, len(inputLines))
	for i, line := range inputLines {
		player, err := parsePlayer(line)
		if err != nil {
			return nil, fmt.Errorf("parse line %d: %w", i, err)
		}

		players = append(players, player)
	}

	return players, nil
}

func parsePlayer(inputLine string) (Player, error) {
	lineComponents := strings.Split(inputLine, " ")
	if len(lineComponents) != 2 {
		return Player{}, errors.New("malformed player line")
	}

	hand, err := parseHand(lineComponents[0])
	if err != nil {
		return Player{}, fmt.Errorf("parse hand: %w", err)
	}

	bid, err := strconv.Atoi(lineComponents[1])
	if err != nil {
		return Player{}, fmt.Errorf("parse bid: %w", err)
	}

	return Player{bid: bid, hand: hand}, nil
}

func parseHand(handStr string) (Hand, error) {
	cardMap := map[byte]Card{
		'A': Ace,
		'K': King,
		'Q': Queen,
		'J': Jack,
		'T': Ten,
		'9': Nine,
		'8': Eight,
		'7': Seven,
		'6': Six,
		'5': Five,
		'4': Four,
		'3': Three,
		'2': Two,
	}

	hand := Hand{}
	for _, handChar := range handStr {
		card, ok := cardMap[byte(handChar)]
		if !ok {
			return nil, fmt.Errorf("invalid card character %c", handChar)
		}

		hand = append(hand, card)
	}

	return hand, nil
}

func mapValues[T comparable, U any](m map[T]U) []U {
	values := make([]U, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}

	return values
}
