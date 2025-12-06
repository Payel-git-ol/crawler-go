package markov

import (
	"fmt"
	"math/rand"
	"time"
)

// MarkovChain represents a Markov chain for URL traversal
type MarkovChain struct {
	transitions map[string][]string
	seed        int64
	rng         *rand.Rand
}

// NewMarkovChain creates a new Markov chain
func NewMarkovChain(seed int64) *MarkovChain {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &MarkovChain{
		transitions: make(map[string][]string),
		seed:        seed,
		rng:         rand.New(rand.NewSource(seed)),
	}
}

// AddTransition adds a transition from one state to another
func (mc *MarkovChain) AddTransition(from, to string) {
	if mc.transitions[from] == nil {
		mc.transitions[from] = []string{}
	}
	mc.transitions[from] = append(mc.transitions[from], to)
}

// GetNextState returns the next state based on Markov chain
func (mc *MarkovChain) GetNextState(current string) (string, error) {
	nextStates, exists := mc.transitions[current]
	if !exists || len(nextStates) == 0 {
		return "", fmt.Errorf("no transitions from state: %s", current)
	}

	// Randomly select one of the next states
	idx := mc.rng.Intn(len(nextStates))
	return nextStates[idx], nil
}

// GetAllTransitions returns all transitions
func (mc *MarkovChain) GetAllTransitions() map[string][]string {
	return mc.transitions
}

// GetState returns transitions from a specific state
func (mc *MarkovChain) GetState(state string) []string {
	return mc.transitions[state]
}

// RemoveTransition removes a specific transition
func (mc *MarkovChain) RemoveTransition(from, to string) {
	states, exists := mc.transitions[from]
	if !exists {
		return
	}

	for i, s := range states {
		if s == to {
			mc.transitions[from] = append(states[:i], states[i+1:]...)
			break
		}
	}

	if len(mc.transitions[from]) == 0 {
		delete(mc.transitions, from)
	}
}

// ClearTransitions clears all transitions
func (mc *MarkovChain) ClearTransitions() {
	mc.transitions = make(map[string][]string)
}

// GetStateCount returns the number of unique states
func (mc *MarkovChain) GetStateCount() int {
	return len(mc.transitions)
}

// GetTransitionCount returns the total number of transitions
func (mc *MarkovChain) GetTransitionCount() int {
	count := 0
	for _, states := range mc.transitions {
		count += len(states)
	}
	return count
}
