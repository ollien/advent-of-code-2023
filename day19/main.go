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

type Part struct {
	XtremelyCoolRating int
	MusicalRating      int
	AerodynamicRating  int
	ShinyRating        int
}

type RuleCondition struct {
	PartRatingFunc     func(Part) int
	OperatorFunc       func(int, int) bool
	Operand            int
	SuccessDestination string
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
	sections := strings.Split(input, "\n\n")
	if len(sections) != 2 {
		panic(fmt.Sprintf("input file did not have expected number of sections (got %d, expected 2)", len(sections)))
	}

	rawRules := strings.Split(strings.TrimSpace(sections[0]), "\n")
	rules, err := parseRules(rawRules)
	if err != nil {
		panic(fmt.Sprintf("could not parse rules: %s", err))
	}

	rawParts := strings.Split(strings.TrimSpace(sections[1]), "\n")
	parts, err := parseParts(rawParts)
	if err != nil {
		panic(fmt.Sprintf("could not parse parts: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(rules, parts))
}

func part1(rules map[string]func(Part) string, parts []Part) int {
	acceptedParts := []Part{}
	for _, part := range parts {
		accepted, err := isPartAccepted(rules, part)
		if err != nil {
			panic(fmt.Sprintf("could not process part %v: %s", part, err))
		}

		if accepted {
			acceptedParts = append(acceptedParts, part)
		}
	}

	acceptedRatings := 0
	for _, part := range acceptedParts {
		acceptedRatings += part.XtremelyCoolRating
		acceptedRatings += part.MusicalRating
		acceptedRatings += part.AerodynamicRating
		acceptedRatings += part.ShinyRating
	}

	return acceptedRatings
}

func isPartAccepted(rules map[string]func(Part) string, part Part) (bool, error) {
	startRule, ok := rules["in"]
	if !ok {
		return false, errors.New("could not locate start rule")
	}

	rule := startRule
	for {
		nextRuleName := rule(part)
		if nextRuleName == "A" {
			return true, nil
		} else if nextRuleName == "R" {
			return false, nil
		}

		rule, ok = rules[nextRuleName]
		if !ok {
			return false, fmt.Errorf("could not locate rule %q", nextRuleName)
		}
	}
}

func parseParts(inputLines []string) ([]Part, error) {
	return tryParse(inputLines, parsePart)
}

func parsePart(input string) (Part, error) {
	partPattern := regexp.MustCompile(`^\{x=(\d+),m=(\d+),a=(\d+),s=(\d+)\}$`)
	matches := partPattern.FindStringSubmatch(input)
	if matches == nil {
		return Part{}, errors.New("malformed part")
	}

	ratings, err := tryParse(matches[1:], strconv.Atoi)
	if err != nil {
		// can't happen because the pattern only has integers
		panic(fmt.Sprintf("could not parse ratings: %s", err))
	}

	return Part{
		XtremelyCoolRating: ratings[0],
		MusicalRating:      ratings[1],
		AerodynamicRating:  ratings[2],
		ShinyRating:        ratings[3],
	}, nil
}

func parseRules(inputLines []string) (map[string]func(Part) string, error) {
	rules := make(map[string]func(Part) string, len(inputLines))
	for i, rawRule := range inputLines {
		ruleName, findDest, err := parseRule(rawRule)
		if err != nil {
			return nil, fmt.Errorf("invalid rule #%d: %w", i, err)
		}

		rules[ruleName] = findDest
	}

	return rules, nil
}

func parseRule(rawRule string) (string, func(Part) string, error) {
	declarationsPattern := regexp.MustCompile(`^([a-z]+)\{((?:[xmas][<>]\d+:[a-zAR]+,)+)([a-zAR]+)\}$`)
	declarationMatches := declarationsPattern.FindStringSubmatch(rawRule)
	if declarationMatches == nil {
		fmt.Println(rawRule)
		return "", nil, errors.New("malformed declarations")
	}

	name := declarationMatches[1]
	rawConditions := declarationMatches[2]
	fallbackDestination := declarationMatches[3]

	conditions, err := parseRuleConditions(rawConditions)
	if err != nil {
		return "", nil, fmt.Errorf("parse conditions: %w", err)
	}

	destinationFunc := buildRuleFunc(conditions, fallbackDestination)

	return name, destinationFunc, nil
}

func buildRuleFunc(conditions []RuleCondition, fallbackDestination string) func(Part) string {
	baseFunc := func(Part) string {
		return fallbackDestination
	}

	// We must store all of the destination functions, otherwise we will
	// be binding to old names of functions when wrapping :(
	destFuncs := []func(Part) string{baseFunc}
	destFunc := func(part Part) string {
		return destFuncs[0](part)
	}

	for i := len(conditions) - 1; i >= 0; i-- {
		condition := conditions[i]
		lastFunc := destFuncs[len(destFuncs)-1]
		ruleDestFunc := func(part Part) string {
			value := condition.PartRatingFunc(part)
			if condition.OperatorFunc(value, condition.Operand) {
				return condition.SuccessDestination
			} else {
				return lastFunc(part)
			}
		}

		destFuncs = append(destFuncs, ruleDestFunc)
		destFunc = ruleDestFunc
	}

	return destFunc
}

func parseRuleConditions(rawConditions string) ([]RuleCondition, error) {
	conditionPattern := regexp.MustCompile(`^([xmas])([<>])(\d+):([a-zAR]+)$`)

	operatorFuncs := map[string]func(int, int) bool{
		"<": func(i1, i2 int) bool { return i1 < i2 },
		">": func(i1, i2 int) bool { return i1 > i2 },
	}

	variableFuncs := map[string]func(Part) int{
		"x": func(part Part) int { return part.XtremelyCoolRating },
		"m": func(part Part) int { return part.MusicalRating },
		"a": func(part Part) int { return part.AerodynamicRating },
		"s": func(part Part) int { return part.ShinyRating },
	}

	splitRawConditions := strings.Split(strings.TrimRight(rawConditions, ","), ",")
	conditions := make([]RuleCondition, len(splitRawConditions))
	for i, rawCondition := range splitRawConditions {
		conditionMatches := conditionPattern.FindStringSubmatch(rawCondition)
		if conditionMatches == nil {
			return nil, fmt.Errorf("malformed condition %q", rawCondition)
		}

		variable := conditionMatches[1]
		operator := conditionMatches[2]
		rawOperand := conditionMatches[3]
		destination := conditionMatches[4]

		operand, err := strconv.Atoi(rawOperand)
		if err != nil {
			// cannot happen, we guarantee this is an integer from the pattern
			panic(fmt.Sprintf("could not parse condition: %s", err))
		}

		operatorFunc, ok := operatorFuncs[operator]
		if !ok {
			panic(fmt.Sprintf("invalid operator %s", operator))
		}

		variableFunc, ok := variableFuncs[variable]
		if !ok {
			panic(fmt.Sprintf("invalid rating variable %s", variable))
		}

		conditions[i] = RuleCondition{
			PartRatingFunc:     variableFunc,
			OperatorFunc:       operatorFunc,
			Operand:            operand,
			SuccessDestination: destination,
		}
	}

	return conditions, nil
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
