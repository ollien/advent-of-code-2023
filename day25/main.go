package main

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
)

type ParsedComponent struct {
	Name      string
	Connected []string
}

type ParsedCut struct {
	Node1 string
	Node2 string
}

func main() {
	if !((len(os.Args) == 3 && os.Args[2] == "dot") || (len(os.Args) > 3 && os.Args[2] == "cut")) {
		fmt.Fprintf(os.Stderr, "Usage: %s inputfile command\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "command must be 'dot' or 'cut'. 'cut' is followed by a comma separated list of edges to cut (e.g. \"abc,bcd cde,def\")")
		fmt.Fprintln(os.Stderr, "This solution uses graphviz (neato) + manual inspection, and then will return the answer with a given cutset")
		os.Exit(1)
	}

	command := os.Args[2]
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
	components, err := parseComponents(inputLines)
	if err != nil {
		panic(fmt.Sprintf("invalid input: %s", err))
	}

	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	if command == "dot" {
		printDOT(components)
	} else if command == "cut" {
		cuts, err := parseCuts(os.Args[3:])
		if err != nil {
			panic(fmt.Sprintf("failed to part cuts: %s", err))
		}

		fmt.Printf("Part 1: %d\n", part1(components, cuts))
	} else {
		panic("invalid command")
	}
}

func part1(allComponents map[string][]string, cuts []ParsedCut) int {
	deleteEdgeBetween := func(components map[string][]string, source, target string) {
		sourceEdges := components[source]
		components[source] = slices.DeleteFunc(sourceEdges, func(name string) bool {
			return name == target
		})
	}

	trimmedComponents := maps.Clone(allComponents)
	for _, cut := range cuts {
		deleteEdgeBetween(trimmedComponents, cut.Node1, cut.Node2)
		deleteEdgeBetween(trimmedComponents, cut.Node2, cut.Node1)
	}

	visitedCounts := map[int]struct{}{}
	fullGraph := buildFullGraph(trimmedComponents)

	for node := range trimmedComponents {
		visitedCount := numNodesReachableFrom(fullGraph, node)
		visitedCounts[visitedCount] = struct{}{}
	}

	// This makes a bit of an assumption, which is that each section will have a different set of reachable nodes
	// However, I don't think the puzzle is going to have them both doing that, so we make this assumption
	if len(visitedCounts) != 2 {
		panic("No two distinct sections were found")
	}

	total := 1
	for count := range visitedCounts {
		total *= count
	}

	return total
}

func buildFullGraph(components map[string][]string) map[string][]string {
	graphSets := make(map[string]map[string]struct{})
	for source, children := range components {
		for _, child := range children {
			if graphSets[source] == nil {
				graphSets[source] = make(map[string]struct{})
			}

			if graphSets[child] == nil {
				graphSets[child] = make(map[string]struct{})
			}
			graphSets[source][child] = struct{}{}
			graphSets[child][source] = struct{}{}
		}
	}

	fullGraph := make(map[string][]string)
	for source, children := range graphSets {
		for child := range children {
			fullGraph[source] = append(fullGraph[source], child)
		}
	}

	return fullGraph
}

func numNodesReachableFrom(components map[string][]string, source string) int {
	toVisit := []string{source}
	visited := map[string]struct{}{}
	for len(toVisit) > 0 {
		visiting := toVisit[0]
		toVisit = toVisit[1:]
		visited[visiting] = struct{}{}

		for _, neighbor := range components[visiting] {
			if _, ok := visited[neighbor]; ok {
				continue
			}

			toVisit = append(toVisit, neighbor)
		}
	}

	return len(visited)
}

func printDOT(components map[string][]string) {
	fmt.Println("graph {")
	for name, connectedChildren := range components {
		for _, child := range connectedChildren {
			fmt.Printf("  %s -- %s;\n", name, child)
		}
	}
	fmt.Println("}")
}

func parseComponents(lines []string) (map[string][]string, error) {
	parsedComponents, err := tryParse(lines, parseComponentLine)
	if err != nil {
		return nil, err
	}

	res := make(map[string][]string, len(parsedComponents))
	for _, component := range parsedComponents {
		res[component.Name] = component.Connected
	}

	return res, nil
}

func parseComponentLine(line string) (ParsedComponent, error) {
	pattern := regexp.MustCompile(`^([a-z]{3}): ((?:[a-z]{3}\s?)+)$`)
	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		return ParsedComponent{}, errors.New("malformed component line")
	}

	return ParsedComponent{
		Name:      matches[1],
		Connected: strings.Split(matches[2], " "),
	}, nil
}

func parseCuts(cuts []string) ([]ParsedCut, error) {
	return tryParse(cuts, parseCut)
}

func parseCut(cut string) (ParsedCut, error) {
	pattern := regexp.MustCompile(`^([a-z]{3}),([a-z]{3})$`)
	matches := pattern.FindStringSubmatch(cut)
	if matches == nil {
		return ParsedCut{}, nil
	}

	return ParsedCut{
		Node1: matches[1],
		Node2: matches[2],
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
