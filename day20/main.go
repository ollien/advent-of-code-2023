package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
)

const BroadcasterName = "broadcaster"

type WorkQueue []PendingPulse

type Pulse int

const (
	PulseLow Pulse = iota
	PulseHigh
)

type ParsedModuleKind int

type ParsedModule struct {
	Name     string
	Kind     ParsedModuleKind
	Children []string
}

const (
	ParsedModuleBroadcaster ParsedModuleKind = iota
	ParsedModuleFlipFlop
	ParsedModuleConjunction
)

type PulseHandler interface {
	ID() string
	HandlePulse(source PulseHandler, pulse Pulse)
}

var _ PulseHandler = &FlipFlop{}
var _ PulseHandler = &Conjunction{}
var _ PulseHandler = &Broadcaster{}
var _ PulseHandler = &ModuleShim{}

type ModuleInfo struct {
	name     string
	children []PulseHandler
}

type FlipFlop struct {
	ModuleInfo
	state bool
}

type Conjunction struct {
	ModuleInfo
	lastPulseFrom map[string]Pulse
}

type Broadcaster struct {
	ModuleInfo
}

type ModuleShim struct {
	name         string
	knownModules map[string]PulseHandler
	pulseQueue   *WorkQueue
}

type PendingPulse struct {
	pulse     Pulse
	sendPulse func()
}

func NewFlipFlop(name string, children []PulseHandler) *FlipFlop {
	return &FlipFlop{
		ModuleInfo: ModuleInfo{
			name:     name,
			children: children,
		},
		state: false,
	}
}

func NewBroadcaster(name string, children []PulseHandler) *Broadcaster {
	return &Broadcaster{
		ModuleInfo: ModuleInfo{
			name:     name,
			children: children,
		},
	}
}

func NewConjunction(name string, sources []PulseHandler, children []PulseHandler) *Conjunction {
	receivedHighPulseFrom := make(map[string]Pulse)
	for _, source := range sources {
		receivedHighPulseFrom[source.ID()] = PulseLow
	}

	return &Conjunction{
		lastPulseFrom: receivedHighPulseFrom,
		ModuleInfo: ModuleInfo{
			name:     name,
			children: children,
		},
	}
}

func NewShimmedModule(name string, allModules map[string]PulseHandler, pulseQueue *WorkQueue) ModuleShim {
	return ModuleShim{
		name:         name,
		knownModules: allModules,
		pulseQueue:   pulseQueue,
	}
}

func (handler ModuleShim) ID() string {
	return handler.name
}

func (handler ModuleShim) HandlePulse(source PulseHandler, pulse Pulse) {
	module, ok := handler.knownModules[handler.name]
	pulseFunc := func() {
		module.HandlePulse(source, pulse)
	}
	if !ok {
		// nop, nothing to send it to
		pulseFunc = func() {}
	}

	pendingPulse := PendingPulse{
		pulse:     pulse,
		sendPulse: pulseFunc,
	}

	handler.pulseQueue.Push(pendingPulse)
}

func (moduleInfo *ModuleInfo) ID() string {
	return moduleInfo.name
}

func (flipFlop *FlipFlop) HandlePulse(_ PulseHandler, pulse Pulse) {
	if pulse == PulseHigh {
		return
	} else if pulse != PulseLow {
		panic(fmt.Sprintf("invalid pulse value %d", pulse))
	}

	nextPulse := PulseHigh
	if flipFlop.state {
		nextPulse = PulseLow
	}

	broadcastPulse(flipFlop, flipFlop.children, nextPulse)

	flipFlop.state = !flipFlop.state
}

func (broadcaster *Broadcaster) HandlePulse(_ PulseHandler, pulse Pulse) {
	broadcastPulse(broadcaster, broadcaster.children, pulse)
}

func (conjunction *Conjunction) HandlePulse(source PulseHandler, pulse Pulse) {
	conjunction.lastPulseFrom[source.ID()] = pulse
	if conjunction.allInputsHigh() {
		broadcastPulse(conjunction, conjunction.children, PulseLow)
	} else {
		broadcastPulse(conjunction, conjunction.children, PulseHigh)
	}
}

func (conjunction *Conjunction) allInputsHigh() bool {
	for _, pulse := range conjunction.lastPulseFrom {
		if pulse == PulseLow {
			return false
		}
	}

	return true
}

func (pulse Pulse) String() string {
	switch pulse {
	case PulseHigh:
		return "PulseHigh"
	case PulseLow:
		return "PulseLow"
	default:
		panic(fmt.Sprintf("invalid pulse value %d", pulse))
	}
}

func (queue *WorkQueue) Push(pulse PendingPulse) {
	*queue = append(*queue, pulse)
}

func (queue *WorkQueue) Pop() PendingPulse {
	if len(*queue) == 0 {
		panic("cannot pop empty queue")
	}

	front := (*queue)[0]
	*queue = (*queue)[1:]

	return front
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

	workQueue := make(WorkQueue, 0)
	modules, err := buildModulesFromInput(inputLines, &workQueue)
	if err != nil {
		panic(fmt.Sprintf("could not parse modules: %s", err))
	}

	fmt.Printf("Part 1: %d\n", part1(modules, &workQueue))
}

func part1(modules map[string]PulseHandler, workQueue *WorkQueue) int {
	button := NewBroadcaster("button", []PulseHandler{
		NewShimmedModule("broadcaster", modules, workQueue),
	})

	pulseCounts := map[Pulse]int{}
	for i := 0; i < 1000; i++ {
		button.HandlePulse(nil, PulseLow)

		for len(*workQueue) > 0 {
			work := workQueue.Pop()
			pulseCounts[work.pulse]++
			work.sendPulse()
		}
		fmt.Println()
	}

	fmt.Println(pulseCounts)
	total := 1
	for _, n := range pulseCounts {
		total *= n
	}

	return total
}

func buildModulesFromInput(inputLines []string, pulseQueue *WorkQueue) (map[string]PulseHandler, error) {
	parsedModules, err := tryParse(inputLines, parseModule)
	if err != nil {
		return nil, fmt.Errorf("parse modules: %w", err)
	}

	return buildModules(parsedModules, pulseQueue), nil
}

func parseModule(inputLine string) (ParsedModule, error) {
	modulePattern := regexp.MustCompile(`^([%&]?)([a-z]+) -> ((?:[a-z]+(?:, )?)+)$`)
	matches := modulePattern.FindStringSubmatch(inputLine)
	if matches == nil {
		return ParsedModule{}, errors.New("malformed module")
	}

	name := matches[2]
	children := strings.Split(matches[3], ", ")
	moduleKind, err := moduleKindFromPrefix(matches[1])
	if err != nil {
		return ParsedModule{}, fmt.Errorf("module kind: %w", err)
	} else if moduleKind == ParsedModuleBroadcaster && name != BroadcasterName {
		return ParsedModule{}, errors.New("got unexpected broadcaster with different name")
	}

	return ParsedModule{
		Name:     name,
		Children: children,
		Kind:     moduleKind,
	}, nil
}

func buildModules(parsed []ParsedModule, pulseQueue *WorkQueue) map[string]PulseHandler {
	registry := make(map[string]PulseHandler, len(parsed))
	for _, module := range parsed {
		childModules := makeLookupModules(registry, module.Children, pulseQueue)
		switch module.Kind {
		case ParsedModuleBroadcaster:
			registry[module.Name] = NewBroadcaster(module.Name, childModules)
		case ParsedModuleFlipFlop:
			registry[module.Name] = NewFlipFlop(module.Name, childModules)
		case ParsedModuleConjunction:
			parentModuleNames := findParentModules(parsed, module.Name)
			parentModules := makeLookupModules(registry, parentModuleNames, pulseQueue)
			registry[module.Name] = NewConjunction(module.Name, parentModules, childModules)
		}
	}

	return registry
}

func findParentModules(parsed []ParsedModule, desiredModule string) []string {
	parents := []string{}
	for _, module := range parsed {
		moduleIdx := slices.Index(module.Children, desiredModule)
		if moduleIdx != -1 {
			parents = append(parents, module.Name)
		}
	}

	return parents
}

func makeLookupModules(registry map[string]PulseHandler, childNames []string, pulseQueue *WorkQueue) []PulseHandler {
	res := []PulseHandler{}
	for _, childName := range childNames {
		handler := NewShimmedModule(childName, registry, pulseQueue)
		res = append(res, handler)
	}

	return res
}

func moduleKindFromPrefix(prefix string) (ParsedModuleKind, error) {
	switch prefix {
	case "":
		return ParsedModuleBroadcaster, nil
	case "%":
		return ParsedModuleFlipFlop, nil
	case "&":
		return ParsedModuleConjunction, nil
	default:
		return ParsedModuleBroadcaster, fmt.Errorf("invalid module prefix %q", prefix)
	}
}

func broadcastPulse(source PulseHandler, recipients []PulseHandler, pulse Pulse) {
	for _, child := range recipients {
		fmt.Printf("%s --%s--> %s\n", source.ID(), pulse, child.ID())

		child.HandlePulse(source, pulse)
	}
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
