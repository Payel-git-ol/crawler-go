package markov

import (
	"testing"
)

func TestNewMarkovChain(t *testing.T) {
	mc := NewMarkovChain(42)
	if mc == nil {
		t.Fatal("NewMarkovChain returned nil")
	}
}

func TestAddTransition(t *testing.T) {
	mc := NewMarkovChain(42)

	mc.AddTransition("state1", "state2")
	mc.AddTransition("state1", "state3")

	states := mc.GetState("state1")
	if len(states) != 2 {
		t.Errorf("Expected 2 transitions, got %d", len(states))
	}
}

func TestGetNextState(t *testing.T) {
	mc := NewMarkovChain(42)
	mc.AddTransition("state1", "state2")

	next, err := mc.GetNextState("state1")
	if err != nil {
		t.Fatalf("GetNextState returned error: %v", err)
	}

	if next != "state2" {
		t.Errorf("Expected 'state2', got '%s'", next)
	}
}

func TestGetNextStateNoTransition(t *testing.T) {
	mc := NewMarkovChain(42)

	_, err := mc.GetNextState("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent state")
	}
}

func TestRemoveTransition(t *testing.T) {
	mc := NewMarkovChain(42)
	mc.AddTransition("state1", "state2")
	mc.AddTransition("state1", "state3")

	mc.RemoveTransition("state1", "state2")

	states := mc.GetState("state1")
	if len(states) != 1 {
		t.Errorf("Expected 1 transition, got %d", len(states))
	}

	if states[0] != "state3" {
		t.Errorf("Expected 'state3', got '%s'", states[0])
	}
}

func TestGetStateCount(t *testing.T) {
	mc := NewMarkovChain(42)

	mc.AddTransition("state1", "state2")
	mc.AddTransition("state2", "state3")
	mc.AddTransition("state3", "state1")

	count := mc.GetStateCount()
	if count != 3 {
		t.Errorf("Expected 3 states, got %d", count)
	}
}

func TestGetTransitionCount(t *testing.T) {
	mc := NewMarkovChain(42)

	mc.AddTransition("state1", "state2")
	mc.AddTransition("state1", "state3")
	mc.AddTransition("state2", "state3")

	count := mc.GetTransitionCount()
	if count != 3 {
		t.Errorf("Expected 3 transitions, got %d", count)
	}
}

func TestClearTransitions(t *testing.T) {
	mc := NewMarkovChain(42)

	mc.AddTransition("state1", "state2")
	mc.ClearTransitions()

	count := mc.GetStateCount()
	if count != 0 {
		t.Errorf("Expected 0 states after clear, got %d", count)
	}
}
