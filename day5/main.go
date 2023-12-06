package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
)

var conversionSteps = []ConvertsBetween{
	{from: "seed", to: "soil"},
	{from: "soil", to: "fertilizer"},
	{from: "fertilizer", to: "water"},
	{from: "water", to: "light"},
	{from: "light", to: "temperature"},
	{from: "temperature", to: "humidity"},
	{from: "humidity", to: "location"},
}

type Range struct {
	start int
	size  int
}

type ConversionMapEntry struct {
	destRange Range
	srcRange  Range
}

type ConversionMap []ConversionMapEntry

type ConvertsBetween struct {
	from string
	to   string
}

type WorkerData struct {
	startLocation int
	numToProcess  int
}

// Contains checks if the given value is contained in the range
func (r Range) Contains(n int) bool {
	return n >= r.start && n < r.start+r.size
}

// RangeDelta indicates how large the span of the range starts are for this entry
func (entry ConversionMapEntry) RangeDelta() int {
	return entry.destRange.start - entry.srcRange.start
}

// ConvertsTo executes the "conversion" of this step, as defined by the problem
func (conversionMap ConversionMap) ConvertsTo(n int) int {
	for _, entry := range conversionMap {
		if entry.srcRange.Contains(n) {
			return entry.RangeDelta() + n
		}
	}

	return n
}

// ReverseConversion is the onverse of ConvertsTo
func (conversionMap ConversionMap) ReverseConversion(n int) int {
	for _, entry := range conversionMap {
		if entry.destRange.Contains(n) {
			return n - entry.RangeDelta()
		}
	}

	return n
}

func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s inputfile [workSize]\n", os.Args[0])
		os.Exit(1)
	}

	workSize := 1000
	if len(os.Args) == 3 {
		var err error
		workSize, err = strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid work size %s", os.Args[3])
			os.Exit(1)
		}
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

	// I got lazy here
	fmt.Fprintln(os.Stderr, "Warning: Part 2 does not halt in the absence of a solution, so it taking a long time does not mean it will eventually find it")
	fmt.Printf("Part 1: %d\n", part1(seeds, conversions))
	fmt.Printf("Part 2: %d\n", part2(seeds, conversions, workSize))
}

func part1(seeds []int, conversions map[ConvertsBetween]ConversionMap) int {
	min := math.MaxInt
	for _, seed := range seeds {
		location := growPlant(seed, conversions)
		if location < min {
			min = location
		}
	}

	return min
}

func part2(seeds []int, conversions map[ConvertsBetween]ConversionMap, workSize int) int {
	seedRanges, err := makeSeedRanges(seeds)
	if err != nil {
		panic(fmt.Sprintf("failed to make seed ranges: %s", err))
	}

	answerChan := make(chan int)
	workChan := make(chan WorkerData)

	ctx, cancel := context.WithCancel(context.Background())
	numWorkers := runtime.NumCPU()
	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			part2Worker(ctx, seedRanges, conversions, workChan, answerChan)
			wg.Done()
		}()
	}

	dispatchCtx, cancelDispatch := context.WithCancel(ctx)
	go func() {
		dispatchPart2Work(dispatchCtx, workChan, workSize)
		// Wait for all workers to finish, and then close the answer chan so that we know they're gone
		wg.Wait()
		close(answerChan)
	}()

	bestAnswer := math.MaxInt
	for answer := range answerChan {
		// We may get multiple possible answers, but we only wanna take the min
		if answer < bestAnswer {
			bestAnswer = answer
			// Once we get an answer, stop dispatching work
			cancelDispatch()
		}
	}

	cancel()
	// This isn't really needed because we cancel the parent ctx but it satisfies the linter
	cancelDispatch()

	return bestAnswer
}

// growPlant will grow a plant from a seed through all the stages until all the conversions are complete
func growPlant(seed int, conversions map[ConvertsBetween]ConversionMap) int {
	item := seed
	for _, conversionStep := range conversionSteps {
		conversionMap, ok := conversions[conversionStep]
		if !ok {
			panic(fmt.Sprintf("Missing conversion for %s-to-%s", conversionStep.from, conversionStep.to))
		}

		item = conversionMap.ConvertsTo(item)
	}

	return item
}

// dispatchPart2Work will dispatch work to the given workChan. It will give workSize number of items
// for each piece of worker data, allowing fair distribution of work between goroutines
func dispatchPart2Work(ctx context.Context, workChan chan WorkerData, workSize int) {
	for i := 0; ; i += workSize {
		data := WorkerData{
			startLocation: i,
			numToProcess:  workSize,
		}

		select {
		case <-ctx.Done():
			close(workChan)
			return
		case workChan <- data:
		}
	}
}

// part2Worker solves part 2 for a given
func part2Worker(ctx context.Context, seedRanges []Range, conversions map[ConvertsBetween]ConversionMap, workChan <-chan WorkerData, answerChan chan<- int) {
	for workerData := range workChan {
		// fmt.Printf("Got work: %+v\n", workerData)
		for i := 0; i < workerData.numToProcess; i++ {
			// Stop working if context says we're done
			select {
			case <-ctx.Done():
				return
			default:
			}

			location := workerData.startLocation + i
			canReverse, err := canGrowLocation(seedRanges, conversions, location)
			if err != nil {
				// We can't do a ton here without a bunch of error plumbing that is too much work for an AoC problem.
				// An operator will see the error and handle it  :)
				fmt.Fprintf(os.Stderr, "Cannot reverse growth for location %d\n", location)
				continue
			}

			if canReverse {
				select {
				case <-ctx.Done():
					return
				case answerChan <- location:
				}
			}
		}
	}
}

// canGrowLocation indicates whether or not the location can be grown with the given seedRanges
func canGrowLocation(seedRanges []Range, conversions map[ConvertsBetween]ConversionMap, location int) (bool, error) {
	currentItem := location
	for i := len(conversionSteps) - 1; i >= 0; i-- {
		step := conversionSteps[i]
		conversionMap, ok := conversions[step]
		if !ok {
			return false, fmt.Errorf("no conversion available for %v", step)
		}

		currentItem = conversionMap.ReverseConversion(currentItem)
	}

	idx := slices.IndexFunc(seedRanges, func(r Range) bool {
		return r.Contains(currentItem)
	})

	return idx != -1, nil
}

// makeSeedRanges will convert a set of seed input values to ranges (only needed for part 2)
func makeSeedRanges(seeds []int) ([]Range, error) {
	if len(seeds)%2 != 0 {
		return nil, errors.New("number of seed entries must be even")
	}

	seedRanges := make([]Range, len(seeds)/2)
	for i := 0; i < len(seeds); i += 2 {
		rangeStart := seeds[i]
		rangeSize := seeds[i+1]

		seedRanges = append(seedRanges, Range{start: rangeStart, size: rangeSize})
	}

	return seedRanges, nil
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
			destRange: Range{start: dest, size: size},
			srcRange:  Range{start: src, size: size},
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
