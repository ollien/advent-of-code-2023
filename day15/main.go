package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

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
}

func part1(inputElements []string) int {
	sum := 0
	for _, element := range inputElements {
		sum += hash(element)
	}

	return sum
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
