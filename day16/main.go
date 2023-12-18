package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type Tile rune

const (
	TileMirrorRight        Tile = '/'
	TileMirrorLeft         Tile = '\\'
	TileSplitterVertical   Tile = '|'
	TileSplitterHorizontal Tile = '-'
)

type Direction int

const (
	DirectionNorth Direction = iota
	DirectionEast
	DirectionSouth
	DirectionWest
)

type Coordinate struct {
	row int
	col int
}

type Beam struct {
	position  Coordinate
	direction Direction
}

type TileGrid struct {
	height int
	width  int
	tiles  map[Coordinate]Tile
}

func (dir Direction) Horizontal() bool {
	return dir == DirectionEast || dir == DirectionWest
}

func (dir Direction) Vertical() bool {
	return dir == DirectionNorth || dir == DirectionSouth
}

func (grid TileGrid) Print() {
	grid.PrintWithBeams(nil)
}

func (grid TileGrid) PrintWithBeams(beams []Beam) {
	for row := 0; row < grid.height; row++ {
		for col := 0; col < grid.width; col++ {
			position := Coordinate{row: row, col: col}
			tile, ok := grid.tiles[position]
			beamIdx := slices.IndexFunc(beams, func(beam Beam) bool {
				return beam.position == position
			})

			if beamIdx != -1 && !ok {
				beamAtPosition := beams[beamIdx]
				switch beamAtPosition.direction {
				case DirectionNorth:
					fmt.Print("^")
				case DirectionSouth:
					fmt.Print("v")
				case DirectionWest:
					fmt.Print("<")
				case DirectionEast:
					fmt.Print(">")
				default:
					panic(fmt.Sprintf("invalid direction %d", beamAtPosition.direction))
				}
			} else if !ok {
				fmt.Print(".")
			} else {
				fmt.Printf("%c", tile)
			}
		}
		fmt.Println()
	}
}

func (grid TileGrid) InBounds(position Coordinate) bool {
	return position.row >= 0 && position.col >= 0 && position.row < grid.height && position.col < grid.width
}

func (beam Beam) MovedInDirection(dir Direction) Beam {
	updPosition := beam.position
	switch dir {
	case DirectionNorth:
		updPosition.row--
	case DirectionSouth:
		updPosition.row++
	case DirectionEast:
		updPosition.col++
	case DirectionWest:
		updPosition.col--
	default:
		panic(fmt.Sprintf("invalid direction value %d", dir))
	}

	return Beam{
		position:  updPosition,
		direction: dir,
	}
}

func (beam Beam) RotatedInDirection(dir Direction) Beam {
	return Beam{
		position:  beam.position,
		direction: dir,
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
	inputLines := strings.Split(input, "\n")
	grid, err := parseTileGrid(inputLines)
	if err != nil {
		panic(fmt.Sprintf("failed to parse input: %s", err))
	}

	// grid.Print()

	fmt.Printf("Part 1: %d\n", part1(grid))
}

func part1(grid TileGrid) int {
	startingBeam := Beam{position: Coordinate{row: 0, col: 0}, direction: DirectionEast}
	beams := []Beam{startingBeam}
	nextBeams := []Beam{}
	beamHistory := map[Beam]struct{}{
		startingBeam: {},
	}

	for len(beams) > 0 {
		for _, beam := range beams {
			updBeam := beam.MovedInDirection(beam.direction)
			tile, ok := grid.tiles[updBeam.position]
			if ok {
				resultingBeams := beamsFromCollision(tile, updBeam)
				nextBeams = append(nextBeams, resultingBeams...)
			} else {
				nextBeams = append(nextBeams, updBeam)
			}
		}

		beams = []Beam{}
		for _, beam := range nextBeams {
			_, seenBeam := beamHistory[beam]
			if grid.InBounds(beam.position) && !seenBeam {
				beams = append(beams, beam)
				beamHistory[beam] = struct{}{}
			}
		}

		nextBeams = []Beam{}
	}

	visitedPositions := map[Coordinate]struct{}{}
	for beam := range beamHistory {
		visitedPositions[beam.position] = struct{}{}
	}

	return len(visitedPositions)
}

func beamsFromCollision(tile Tile, beam Beam) []Beam {
	switch tile {
	case TileMirrorLeft, TileMirrorRight:
		return beamsFromMirror(tile, beam)
	case TileSplitterHorizontal, TileSplitterVertical:
		return beamsFromSplitter(tile, beam)
	default:
		panic(fmt.Sprintf("invalid tile %c", tile))
	}
}

func beamsFromSplitter(tile Tile, beam Beam) []Beam {
	if tile != TileSplitterHorizontal && tile != TileSplitterVertical {
		panic("cannot treat a non-splitter tile as a splitter")
	}

	if beam.direction.Horizontal() && tile == TileSplitterHorizontal {
		return []Beam{beam}
	} else if beam.direction.Vertical() && tile == TileSplitterVertical {
		return []Beam{beam}
	} else if beam.direction.Horizontal() && tile == TileSplitterVertical {
		return []Beam{
			beam.RotatedInDirection(DirectionNorth),
			beam.RotatedInDirection(DirectionSouth),
		}
	} else if beam.direction.Vertical() && tile == TileSplitterHorizontal {
		return []Beam{
			beam.RotatedInDirection(DirectionWest),
			beam.RotatedInDirection(DirectionEast),
		}
	} else {
		// This can't happen, we've hit all four permutations already
		panic("invalid configuration of beam")
	}
}

func beamsFromMirror(tile Tile, beam Beam) []Beam {
	switch tile {
	case TileMirrorRight:
		return beamsFromRightMirror(tile, beam)
	case TileMirrorLeft:
		return beamsFromLeftMirror(tile, beam)
	default:
		panic("cannot treat non-mirror as a mirror")
	}
}

func beamsFromRightMirror(tile Tile, beam Beam) []Beam {
	if tile != TileMirrorRight {
		panic("cannot treat non-right mirror as right mirror")
	}

	switch beam.direction {
	case DirectionEast:
		return []Beam{beam.RotatedInDirection(DirectionNorth)}
	case DirectionWest:
		return []Beam{beam.RotatedInDirection(DirectionSouth)}
	case DirectionNorth:
		return []Beam{beam.RotatedInDirection(DirectionEast)}
	case DirectionSouth:
		return []Beam{beam.RotatedInDirection(DirectionWest)}
	default:
		panic(fmt.Sprintf("invalid direction value %d", beam.direction))
	}
}

func beamsFromLeftMirror(tile Tile, beam Beam) []Beam {
	if tile != TileMirrorLeft {
		panic("cannot treat non-right mirror as right mirror")
	}

	switch beam.direction {
	case DirectionEast:
		return []Beam{beam.RotatedInDirection(DirectionSouth)}
	case DirectionWest:
		return []Beam{beam.RotatedInDirection(DirectionNorth)}
	case DirectionNorth:
		return []Beam{beam.RotatedInDirection(DirectionWest)}
	case DirectionSouth:
		return []Beam{beam.RotatedInDirection(DirectionEast)}
	default:
		panic(fmt.Sprintf("invalid direction value %d", beam.direction))
	}
}

func parseTileGrid(inputLines []string) (TileGrid, error) {
	if len(inputLines) == 0 {
		return TileGrid{}, fmt.Errorf("no tiles in grid")
	}

	width := len(inputLines[0])
	height := len(inputLines)
	tiles := make(map[Coordinate]Tile)
	for row, line := range inputLines {
		if len(line) != width {
			return TileGrid{}, errors.New("lines have unequal lengths")
		}

		for col, char := range line {
			switch char {
			case '.':
				continue
			case rune(TileMirrorLeft), rune(TileMirrorRight), rune(TileSplitterHorizontal), rune(TileSplitterVertical):
				pos := Coordinate{row: row, col: col}
				tiles[pos] = Tile(char)
			default:
				return TileGrid{}, fmt.Errorf("invalid tile character %c", char)
			}
		}
	}

	return TileGrid{
		height: height,
		width:  width,
		tiles:  tiles,
	}, nil
}
