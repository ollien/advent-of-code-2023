// TOO HIGH  288057819955899

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

type PartRatingType rune

const (
	RatingTypeX PartRatingType = 'x'
	RatingTypeM PartRatingType = 'm'
	RatingTypeA PartRatingType = 'a'
	RatingTypeS PartRatingType = 's'
)

type ComparisonOperator rune

const (
	OperatorGreater ComparisonOperator = '>'
	OperatorLess    ComparisonOperator = '<'
)

type Part struct {
	XtremelyCoolRating int
	MusicalRating      int
	AerodynamicRating  int
	ShinyRating        int
}

type Rule struct {
	Conditions          []RuleCondition
	FallbackDestination string
}

type RuleCondition struct {
	PartRatingType     PartRatingType
	Operator           ComparisonOperator
	Operand            int
	SuccessDestination string
}

type Range struct {
	// both are inclusive
	min int
	max int
}

func (part Part) Rating(ratingType PartRatingType) int {
	switch ratingType {
	case RatingTypeX:
		return part.XtremelyCoolRating
	case RatingTypeM:
		return part.MusicalRating
	case RatingTypeA:
		return part.AerodynamicRating
	case RatingTypeS:
		return part.ShinyRating
	default:
		panic(fmt.Sprintf("invalid rating type %c", ratingType))
	}
}

func (operator ComparisonOperator) Compare(a, b int) bool {
	switch operator {
	case OperatorGreater:
		return a > b
	case OperatorLess:
		return a < b
	default:
		panic(fmt.Sprintf("invalid operator %c", operator))
	}
}

func (r Range) Min() int {
	return r.min
}

func (r Range) Max() int {
	return r.max
}

func (r Range) Empty() bool {
	return r.min > r.max
}

func (r Range) EliminateLessThan(n int) Range {
	if r.Empty() {
		return r
	}

	if n > r.min {
		r.min = n
	}

	return r
}

func (r Range) EliminateLessThanEq(n int) Range {
	return r.EliminateLessThan(n + 1)
}

func (r Range) EliminateGreaterThan(n int) Range {
	if r.Empty() {
		return r
	}

	if n < r.max {
		r.max = n
	}

	return r
}

func (r Range) EliminateGreaterThanEq(n int) Range {
	return r.EliminateGreaterThan(n - 1)
}

func (r Range) Spread() int {
	if r.Empty() {
		return 0
	}

	return r.max - r.min + 1
}

func (r Range) Intersection(other Range) Range {
	return Range{
		min: max(r.min, other.min),
		max: min(r.max, other.max),
	}
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
	fmt.Printf("Part 2: %d\n", part2(rules))
}

func part1(rules map[string]Rule, parts []Part) int {
	ruleFuncs := make(map[string]func(Part) string, len(rules))
	for ruleName, rule := range rules {
		ruleFuncs[ruleName] = buildRuleFunc(rule)
	}

	acceptedParts := []Part{}
	for _, part := range parts {
		accepted, err := isPartAccepted(ruleFuncs, part)
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

func part2(rules map[string]Rule) int {
	ranges := map[PartRatingType]Range{
		RatingTypeX: {min: 1, max: 4000},
		RatingTypeM: {min: 1, max: 4000},
		RatingTypeA: {min: 1, max: 4000},
		RatingTypeS: {min: 1, max: 4000},
	}

	return combinationsSatisfyingRules(rules, "in", ranges)
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

func combinationsSatisfyingRules(rules map[string]Rule, currentRule string, ranges map[PartRatingType]Range) int {
	if currentRule == "R" {
		return 0
	} else if currentRule == "A" {
		combos := 1
		for _, r := range ranges {
			combos *= r.Spread()
		}

		return combos
	}

	rule, ok := rules[currentRule]
	if !ok {
		panic(fmt.Sprintf("invalid rule %s", currentRule))
	}

	combos := 0
	culledRanges := cloneMap(ranges)
	for _, condition := range rule.Conditions {
		affectedRange, ok := culledRanges[condition.PartRatingType]
		if !ok {
			panic(fmt.Sprintf("invalid part rating type %d", condition.PartRatingType))
		}

		if condition.Operator == OperatorGreater {
			updRanges := cloneMap(culledRanges)
			updRanges[condition.PartRatingType] = affectedRange.EliminateLessThanEq(condition.Operand)
			combos += combinationsSatisfyingRules(rules, condition.SuccessDestination, updRanges)

			culledRanges[condition.PartRatingType] = affectedRange.EliminateGreaterThan(condition.Operand)
		} else if condition.Operator == OperatorLess {
			updRanges := cloneMap(culledRanges)
			updRanges[condition.PartRatingType] = affectedRange.EliminateGreaterThanEq(condition.Operand)
			combos += combinationsSatisfyingRules(rules, condition.SuccessDestination, updRanges)

			culledRanges[condition.PartRatingType] = affectedRange.EliminateLessThan(condition.Operand)
		} else {
			panic(fmt.Sprintf("invalid operator %c", condition.Operand))
		}
	}

	return combos + combinationsSatisfyingRules(rules, rule.FallbackDestination, culledRanges)
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

func parseRules(inputLines []string) (map[string]Rule, error) {
	rules := make(map[string]Rule, len(inputLines))
	for i, rawRule := range inputLines {
		ruleName, rule, err := parseRule(rawRule)
		if err != nil {
			return nil, fmt.Errorf("invalid rule #%d: %w", i, err)
		}

		rules[ruleName] = rule
	}

	return rules, nil
}

func parseRule(rawRule string) (string, Rule, error) {
	declarationsPattern := regexp.MustCompile(`^([a-z]+)\{((?:[xmas][<>]\d+:[a-zAR]+,)+)([a-zAR]+)\}$`)
	declarationMatches := declarationsPattern.FindStringSubmatch(rawRule)
	if declarationMatches == nil {
		fmt.Println(rawRule)
		return "", Rule{}, errors.New("malformed declarations")
	}

	name := declarationMatches[1]
	rawConditions := declarationMatches[2]
	fallbackDestination := declarationMatches[3]

	conditions, err := parseRuleConditions(rawConditions)
	if err != nil {
		return "", Rule{}, fmt.Errorf("parse conditions: %w", err)
	}

	rule := Rule{
		Conditions:          conditions,
		FallbackDestination: fallbackDestination,
	}

	return name, rule, nil
}

func buildRuleFunc(rule Rule) func(Part) string {
	baseFunc := func(Part) string {
		return rule.FallbackDestination
	}

	// We must store all of the destination functions, otherwise we will
	// be binding to old names of functions when wrapping :(
	destFuncs := []func(Part) string{baseFunc}
	destFunc := func(part Part) string {
		return destFuncs[0](part)
	}

	for i := len(rule.Conditions) - 1; i >= 0; i-- {
		condition := rule.Conditions[i]
		lastFunc := destFuncs[len(destFuncs)-1]
		ruleDestFunc := func(part Part) string {
			value := part.Rating(condition.PartRatingType)
			if condition.Operator.Compare(value, condition.Operand) {
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

	splitRawConditions := strings.Split(strings.TrimRight(rawConditions, ","), ",")
	conditions := make([]RuleCondition, len(splitRawConditions))
	for i, rawCondition := range splitRawConditions {
		conditionMatches := conditionPattern.FindStringSubmatch(rawCondition)
		if conditionMatches == nil {
			return nil, fmt.Errorf("malformed condition %q", rawCondition)
		}

		// These first two are definitely safe, because the pattern restricts
		// these values to single-chars that are available in their types
		ratingType := PartRatingType(conditionMatches[1][0])
		operator := ComparisonOperator(conditionMatches[2][0])
		rawOperand := conditionMatches[3]
		destination := conditionMatches[4]

		operand, err := strconv.Atoi(rawOperand)
		if err != nil {
			// cannot happen, we guarantee this is an integer from the pattern
			panic(fmt.Sprintf("could not parse condition: %s", err))
		}

		conditions[i] = RuleCondition{
			PartRatingType:     ratingType,
			Operator:           operator,
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

func cloneMap[T comparable, U any, M ~map[T]U](m M) M {
	copy := make(M, len(m))
	for key, value := range m {
		copy[key] = value
	}

	return copy
}
