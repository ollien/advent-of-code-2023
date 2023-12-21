package main

import (
	"slices"
	"testing"
)

type FakePulseHandler struct {
	pulses []Pulse
}

func (handler *FakePulseHandler) ID() string {
	return "fake"
}

func (handler *FakePulseHandler) HandlePulse(_ PulseHandler, pulse Pulse) {
	handler.pulses = append(handler.pulses, pulse)
}

func TestFlipFlopDoesNothingInDefaultStateWithHighPulse(t *testing.T) {
	source := NewFlipFlop("1337", []PulseHandler{})
	receiver := FakePulseHandler{}

	flipFlop := NewFlipFlop("asdf", []PulseHandler{&receiver})

	flipFlop.HandlePulse(source, PulseHigh)
	if len(receiver.pulses) != 0 {
		t.Fatalf("Got %d pulses, not 0", len(receiver.pulses))
	}
}

func TestFlipFlopEmitsHighWhenTurningOn(t *testing.T) {
	source := NewFlipFlop("1337", []PulseHandler{})
	receiver := FakePulseHandler{}

	flipFlop := NewFlipFlop("asdf", []PulseHandler{&receiver})

	flipFlop.HandlePulse(source, PulseLow)
	if !slices.Equal(receiver.pulses, []Pulse{PulseHigh}) {
		t.Fatalf("Got %d pulses, not 1 high", len(receiver.pulses))
	}
}

func TestFlipFlopEmitsLowWhenTurningOff(t *testing.T) {
	source := NewFlipFlop("1337", []PulseHandler{})
	receiver := FakePulseHandler{}

	flipFlop := NewFlipFlop("asdf", []PulseHandler{&receiver})

	flipFlop.HandlePulse(source, PulseLow)
	flipFlop.HandlePulse(source, PulseLow)
	if !slices.Equal(receiver.pulses, []Pulse{PulseHigh, PulseLow}) {
		t.Fatalf("Got %d pulses, not high then low", len(receiver.pulses))
	}
}

func TestBroadcasterSendsPulseToAllChildren(t *testing.T) {
	source := NewFlipFlop("1337", []PulseHandler{})
	receiver1 := FakePulseHandler{}
	receiver2 := FakePulseHandler{}
	receiver3 := FakePulseHandler{}

	broadcaster := NewBroadcaster("asdf", []PulseHandler{&receiver1, &receiver2, &receiver3})
	broadcaster.HandlePulse(source, PulseHigh)
	broadcaster.HandlePulse(source, PulseLow)

	if !slices.Equal(receiver1.pulses, []Pulse{PulseHigh, PulseLow}) {
		t.Fatalf("Got %d pulses, not high then low for receiver1", len(receiver1.pulses))
	}

	if !slices.Equal(receiver2.pulses, []Pulse{PulseHigh, PulseLow}) {
		t.Fatalf("Got %d pulses, not high then low for receiver2", len(receiver2.pulses))
	}

	if !slices.Equal(receiver2.pulses, []Pulse{PulseHigh, PulseLow}) {
		t.Fatalf("Got %d pulses, not high then low for receiver3", len(receiver2.pulses))
	}
}

func TestConjunctionSendsLowPulseForEveryIncompleteSource(t *testing.T) {
	source1 := NewFlipFlop("1337", []PulseHandler{})
	source2 := NewFlipFlop("1338", []PulseHandler{})
	source3 := NewFlipFlop("1339", []PulseHandler{})
	receiver := FakePulseHandler{}

	conjunction := NewConjunction("asdf", []PulseHandler{source1, source2, source3}, []PulseHandler{&receiver})

	conjunction.HandlePulse(source1, PulseHigh)
	conjunction.HandlePulse(source2, PulseHigh)

	if !slices.Equal(receiver.pulses, []Pulse{PulseHigh, PulseHigh}) {
		t.Fatalf("Got %d pulses, not high x2", len(receiver.pulses))
	}
}

func TestConjunctionSendsHighPulseOnceAllSourcesHaveEmitted(t *testing.T) {
	source1 := NewFlipFlop("1337", []PulseHandler{})
	source2 := NewFlipFlop("1338", []PulseHandler{})
	source3 := NewFlipFlop("1339", []PulseHandler{})
	receiver := FakePulseHandler{}

	conjunction := NewConjunction("asdf", []PulseHandler{source1, source2, source3}, []PulseHandler{&receiver})

	conjunction.HandlePulse(source1, PulseHigh)
	conjunction.HandlePulse(source2, PulseHigh)
	conjunction.HandlePulse(source3, PulseHigh)
	conjunction.HandlePulse(source1, PulseHigh)
	conjunction.HandlePulse(source2, PulseHigh)
	conjunction.HandlePulse(source3, PulseHigh)

	if !slices.Equal(receiver.pulses, []Pulse{PulseHigh, PulseHigh, PulseLow, PulseLow, PulseLow, PulseLow}) {
		t.Fatalf("Got %d pulses, not high x2 + low x4", len(receiver.pulses))
	}
}
