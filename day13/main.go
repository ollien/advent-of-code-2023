package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

// Coordinate represents a position in our pattern
type Coordinate struct {
	row int
	col int
}

// SmudgeComparer is a stateful comparator between two slices. The idea is that for a given chain of comparisons,
// the two arrays will be allowed to differ by exactly one item, exactly once. All other times, they must be
// exactly identical.
type SmudgeComparer struct {
	haveDoneModification bool
}

// ComparePossiblySmudgedLine performs the smudged comparison. See struct doc for more details
func (comparer *SmudgeComparer) ComparePossiblySmudgedLine(a []int, b []int) bool {
	if comparer.haveDoneModification {
		return slices.Equal(a, b)
	}

	_, _, err := findExtraItem(a, b)
	if errors.Is(err, errSlicesEqual) {
		return true
	} else if errors.Is(err, errSlicesDiffer) {
		return false
	} else {
		comparer.haveDoneModification = true
		return true
	}
}

// HaveDoneModification indicates whether or not a modification has been performed during previous runs of
// / ComparePossiblySmudgedLine
func (comparer *SmudgeComparer) HaveDoneModification() bool {
	return comparer.haveDoneModification
}

var errSlicesEqual = errors.New("slices are equal")
var errSlicesDiffer = errors.New("slices differ by more than one element")

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
	rawPatternSections := strings.Split(input, "\n\n")
	patternSections, err := tryParse(rawPatternSections, parsePatternSection)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(patternSections))
	fmt.Printf("Part 2: %d\n", part2(patternSections))
}

func part1(sections []map[Coordinate]struct{}) int {
	total := 0
	for i, section := range sections {
		sectionResults, err := evaluateSection(section)
		if err != nil {
			panic(fmt.Sprintf("invalid section %d: %s", i, err))
		}

		total += sectionResults
	}

	return total
}

func part2(sections []map[Coordinate]struct{}) int {
	total := 0
	for i, section := range sections {
		sectionResults, err := evaluateSmudgedSection(section)
		if err != nil {
			panic(fmt.Sprintf("invalid section %d: %s", i, err))
		}

		total += sectionResults
	}

	return total
}

// evaluateSection will summarize the given section according to the mirror rules for part 1
func evaluateSection(section map[Coordinate]struct{}) (int, error) {
	maxRow := maxAlongAxis(Coordinate{row: 1, col: 0}, section)
	byRow := getAllAlongAxis(Coordinate{row: 1, col: 0}, section)
	for row := 0; row < maxRow; row++ {
		if isMirroredAcrossAxis(row, byRow, slices.Equal) {
			return (row + 1) * 100, nil
		}
	}

	maxCol := maxAlongAxis(Coordinate{row: 0, col: 1}, section)
	byCol := getAllAlongAxis(Coordinate{row: 0, col: 1}, section)
	for col := 0; col < maxCol; col++ {
		if isMirroredAcrossAxis(col, byCol, slices.Equal) {
			return col + 1, nil
		}
	}

	return 0, errors.New("section is not mirrored")
}

// evaluateSection will summarize the given section according to the mirror rules for part 2
func evaluateSmudgedSection(section map[Coordinate]struct{}) (int, error) {
	maxRow := maxAlongAxis(Coordinate{row: 1, col: 0}, section)
	byRow := getAllAlongAxis(Coordinate{row: 1, col: 0}, section)
	for row := 0; row < maxRow; row++ {
		comparer := SmudgeComparer{}
		if isMirroredAcrossAxis(row, byRow, comparer.ComparePossiblySmudgedLine) && comparer.haveDoneModification {
			return (row + 1) * 100, nil
		}
	}

	maxCol := maxAlongAxis(Coordinate{row: 0, col: 1}, section)
	byCol := getAllAlongAxis(Coordinate{row: 0, col: 1}, section)
	for col := 0; col < maxCol; col++ {
		comparer := SmudgeComparer{}
		if isMirroredAcrossAxis(col, byCol, comparer.ComparePossiblySmudgedLine) && comparer.haveDoneModification {
			return col + 1, nil
		}
	}

	return 0, errors.New("section is not mirrored")
}

// isMirroredAcrossAxis checks, For the given item along an axis, and a mapping between current axis items and
// perpendicular axis items, whether or not the surrounding rows could be considered mirrored
func isMirroredAcrossAxis(axisIdx int, byAxis map[int][]int, eqFunc func(a []int, b []int) bool) bool {
	maxAxisIdx := maxKey(byAxis)
	// Radiate "outwards" from the axis item, and check the corresponding elements
	for offset := 1; axisIdx+offset <= maxAxisIdx && axisIdx-(offset-1) >= 0; offset++ {
		axisItems := byAxis[axisIdx-(offset-1)]
		oppositeItems := byAxis[axisIdx+offset]
		if !eqFunc(axisItems, oppositeItems) {
			return false
		}
	}

	return true
}

// findExtraItem finds the index of a standalone differing item in each list. If one of the two return values
// is not negative one, that is the index of the standout item (note that only one will be > -1 at a time).
// If the two slices are equal, errSlicesEqual is returned. If the two slices differ by more than one element,
// errSlicesDiffer is returned.
//
// NOTE: this function assumes that these lists are sorted
func findExtraItem(a, b []int) (int, int, error) {
	if abs(len(a)-len(b)) > 1 {
		return -1, -1, errSlicesDiffer
	}

	aCursor := 0
	bCursor := 0
	aCandidate := -1
	bCandidate := -1
	for aCursor < len(a) && bCursor < len(b) {
		aItem := a[aCursor]
		bItem := b[bCursor]

		if aItem == bItem {
			aCursor++
			bCursor++
		} else if aItem < bItem {
			if aCandidate != -1 || bCandidate != -1 {
				return -1, -1, errSlicesDiffer
			}

			aCandidate = aCursor
			aCursor++
		} else {
			if aCandidate != -1 || bCandidate != -1 {
				return -1, -1, errSlicesDiffer
			}

			bCandidate = bCursor
			bCursor++
		}
	}

	if aCursor < len(a) && bCursor >= len(b) {
		return aCursor, -1, nil
	} else if bCursor >= len(a) && bCursor < len(b) {
		return -1, bCursor, nil
	} else if aCandidate == -1 && bCandidate == -1 {
		return -1, -1, errSlicesEqual
	}

	return aCandidate, bCandidate, nil
}

// maxAlongAxis finds the max value along the given axis, using a coordinate set of {1, 0} or {0, 1}
// Panics if an invalid axis is given
func maxAlongAxis(axis Coordinate, section map[Coordinate]struct{}) int {
	if !(axis.row == 0 && axis.col == 1 || axis.row == 1 && axis.col == 0) {
		panic("invalid axis")
	}

	res := 0
	for coord := range section {
		alongAxis := coord.row*axis.row + coord.col*axis.col
		if alongAxis > res {
			res = alongAxis
		}
	}

	return res
}

// getAllAlongAxis gets all items along the given axis, with the resulting map holding each index along that axis
// and the values holding all elements that are perpendicular to that axis
//
// Panics if an invalid axis is given
func getAllAlongAxis(axis Coordinate, section map[Coordinate]struct{}) map[int][]int {
	if !(axis.row == 0 && axis.col == 1 || axis.row == 1 && axis.col == 0) {
		panic("invalid axis")
	}

	res := map[int][]int{}
	for coord := range section {
		alongAxis := coord.row*axis.row + coord.col*axis.col
		notAlongAxis := coord.row*(1-axis.row) + coord.col*(1-axis.col)
		res[alongAxis] = append(res[alongAxis], notAlongAxis)
	}

	for _, perpendicularItems := range res {
		slices.Sort(perpendicularItems)
	}

	return res
}

func parsePatternSection(section string) (map[Coordinate]struct{}, error) {
	sectionLines := strings.Split(section, "\n")
	coords := map[Coordinate]struct{}{}
	for row, line := range sectionLines {
		for col, char := range line {
			if char == '#' {
				coords[Coordinate{row: row, col: col}] = struct{}{}
			} else if char != '.' {
				return nil, fmt.Errorf("invalid char: %c", char)
			}
		}
	}

	return coords, nil
}

func maxKey[T cmp.Ordered, U any](m map[T]U) T {
	if len(m) == 0 {
		panic("cannot get max key of zero length map")
	} else if len(m) == 1 {
		return *new(T)
	}

	keys := make([]T, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	return slices.Max(keys)
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

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}
