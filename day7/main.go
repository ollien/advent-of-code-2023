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
	Two Card = iota + 2
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

func (hand Hand) Kind() (HandKind, error) {
	handCounts := map[HandKind][]int{
		FiveOfAKind:  {5},
		FourOfAKind:  {1, 4},
		FullHouse:    {2, 3},
		ThreeOfAKind: {1, 1, 3},
		TwoPair:      {1, 2, 2},
		OnePair:      {1, 1, 1, 2},
		HighCard:     {1, 1, 1, 1, 1},
	}

	distinctCounts := hand.CountDistinct()
	cardCounts := mapValues(distinctCounts)
	slices.Sort(cardCounts)
	for kind, kindCounts := range handCounts {
		if slices.Equal(cardCounts, kindCounts) {
			return kind, nil
		}
	}

	return UnknownKind, errors.New("got it")
}

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

	for _, player := range players {
		kind, _ := player.hand.Kind()
		fmt.Printf("%+v %+v\n", player.hand, kind)
	}

	fmt.Printf("Part 1: %d\n", part1(players))
}

func part1(players []Player) int {
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
