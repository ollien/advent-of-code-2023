package main

import (
	"container/list"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type HashMap[V any] []*list.List

type HashMapItem[V any] struct {
	key   string
	value V
}

// Operation is some sort of operation we can perform on a hashmap
type Operation[V any] func(HashMap[V])

// makeSetOperation makes an Operation that will set the given key to the given value in the hashmap
func makeSetOperation[V any](key string, value V) Operation[V] {
	return func(hm HashMap[V]) {
		keyHash := hash(key)

		element := findInList(hm[keyHash], func(item HashMapItem[V]) bool {
			return item.key == key
		})

		if element == nil {
			hm[keyHash].PushBack(HashMapItem[V]{key: key, value: value})
		} else {
			element.Value = HashMapItem[V]{value: value, key: key}
		}
	}
}

// makeDeleteOperation makes an Operation that will delete the given key from the hashmap
func makeDeleteOperation[V any](key string) Operation[V] {
	return func(hm HashMap[V]) {
		keyHash := hash(key)
		element := findInList(hm[keyHash], func(item HashMapItem[V]) bool {
			return item.key == key
		})

		if element != nil {
			hm[keyHash].Remove(element)
		}
	}
}

// NewHashMap makes a new hash map with the given value type
func NewHashMap[V any]() HashMap[V] {
	m := make(HashMap[V], 256)
	for i := range m {
		m[i] = list.New().Init()
	}

	return m
}

// EntriesInBox gets all the entries in the hash map box with the given index
func (hm HashMap[V]) EntriesInBox(idx int) ([]HashMapItem[V], error) {
	if idx > len(hm) {
		return nil, errors.New("out of bounds box id")
	}

	entryList := hm[idx]
	res := make([]HashMapItem[V], 0, entryList.Len())
	for cursor := entryList.Front(); cursor != nil; cursor = cursor.Next() {
		value := cursor.Value.(HashMapItem[V])
		res = append(res, value)
	}

	return res, nil
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
	inputElements := strings.Split(input, ",")

	fmt.Printf("Part 1: %d\n", part1(inputElements))
	fmt.Printf("Part 2: %d\n", part2(inputElements))
}

func part1(inputElements []string) int {
	sum := 0
	for _, element := range inputElements {
		sum += hash(element)
	}

	return sum
}

func part2(inputElements []string) int {
	operations := make([]Operation[int], 0, len(inputElements))
	for _, element := range inputElements {
		operation, err := parseOperation(element)
		if err != nil {
			panic(fmt.Sprintf("could not parse operation %q: %s", element, err))
		}

		operations = append(operations, operation)
	}

	hm := NewHashMap[int]()
	for _, operation := range operations {
		operation(hm)
	}

	power := 0
	for i := range hm {
		boxEntries, err := hm.EntriesInBox(i)
		if err != nil {
			// Can't happen since we're iterating over the known array
			panic(err)
		}

		boxNumber := i + 1
		for j, item := range boxEntries {
			power += boxNumber * (j + 1) * item.value
		}
	}

	return power
}

func hash(s string) int {
	res := 0
	for _, char := range s {
		res += int(char)
		res *= 17
		res %= 256
	}

	return res
}

// findInList finds the given element in the linked list. The element *MUST* be of the type V, or this will panic
func findInList[V any](l *list.List, isEqual func(V) bool) *list.Element {
	for cursor := l.Front(); cursor != nil; cursor = cursor.Next() {
		if isEqual(cursor.Value.(V)) {
			return cursor
		}
	}

	return nil
}

func parseOperation(s string) (Operation[int], error) {
	setPattern := regexp.MustCompile(`^([a-z]+)=(\d+)`)
	deletePattern := regexp.MustCompile(`^([a-z]+)-$`)

	setMatches := setPattern.FindStringSubmatch(s)
	deleteMatches := deletePattern.FindStringSubmatch(s)
	if setMatches != nil {
		key := setMatches[1]
		value, err := strconv.Atoi(setMatches[2])
		if err != nil {
			// cannot happen, since the pattern says this is a number
			panic(err)
		}

		return makeSetOperation(key, value), nil
	} else if deleteMatches != nil {
		key := deleteMatches[1]
		return makeDeleteOperation[int](key), nil
	} else {
		return nil, errors.New("invalid operation string")
	}
}
