// Package prop provides property-based testing functionality for Go.
// It allows you to test properties of your code by generating random test cases
// and automatically shrinking counterexamples when failures are found.
package prop

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/lucaskalb/rapidx/gen"
)

// Config holds the configuration for property-based testing.
type Config struct {
	// Seed is the random seed used for test case generation.
	// If zero, a random seed will be generated based on the current time.
	Seed int64

	// Examples is the number of test cases to generate and run.
	Examples int

	// MaxShrink is the maximum number of shrinking steps to perform
	// when a counterexample is found.
	MaxShrink int

	// ShrinkStrat specifies the shrinking strategy to use.
	// Supported strategies: "bfs" (breadth-first), "dfs" (depth-first).
	ShrinkStrat string

	// StopOnFirstFailure determines whether to stop testing
	// after the first failing test case is found.
	StopOnFirstFailure bool

	// Parallelism specifies the number of parallel workers to use
	// for running test cases. Must be at least 1.
	Parallelism int
}

var (
	// flagSeed sets the random seed for test case generation.
	// Default: 0 (random seed based on current time).
	flagSeed = flag.Int64("rapidx.seed", 0, "Random seed for test case generation")

	// flagExamples sets the number of test cases to generate.
	// Default: 100.
	flagExamples = flag.Int("rapidx.examples", 100, "Number of test cases to generate")

	// flagMaxShrink sets the maximum number of shrinking steps.
	// Default: 400.
	flagMaxShrink = flag.Int("rapidx.maxshrink", 400, "Maximum number of shrinking steps")

	// flagShrinkStrat sets the shrinking strategy.
	// Default: "bfs" (breadth-first search).
	flagShrinkStrat = flag.String("rapidx.shrink.strategy", "bfs", "Shrinking strategy (bfs or dfs)")

	// flagParallelism sets the number of parallel workers.
	// Default: 1.
	flagParallelism = flag.Int("rapidx.shrink.parallel", 1, "Number of parallel workers")
)

// Default returns a Config with default values based on command-line flags.
// This is the recommended way to create a configuration for property-based testing.
func Default() Config {
	return Config{
		Seed:               *flagSeed,
		Examples:           *flagExamples,
		MaxShrink:          *flagMaxShrink,
		ShrinkStrat:        *flagShrinkStrat,
		StopOnFirstFailure: true,
		Parallelism:        *flagParallelism,
	}
}

// effectiveSeed returns the effective seed to use for random number generation.
// If the configured seed is zero, it returns a random seed based on the current time.
func (c Config) effectiveSeed() int64 {
	if c.Seed != 0 {
		return c.Seed
	}
	return time.Now().UnixNano()
}

// ForAll creates a property-based test that generates test cases using the provided generator
// and runs them against the given test function. It returns a function that takes the test
// body as a parameter.
//
// The test will generate cfg.Examples number of test cases, and if any fail, it will attempt
// to shrink the counterexample to find a minimal failing case.
//
// Example usage:
//
//	ForAll(t, prop.Default(), gen.Int())(func(t *testing.T, x int) {
//	    // Test property: x + 0 == x
//	    if x+0 != x {
//	        t.Errorf("addition identity failed for %d", x)
//	    }
//	})
func ForAll[T any](t *testing.T, cfg Config, g gen.Generator[T]) func(func(*testing.T, T)) {
	return func(body func(*testing.T, T)) {
		seed := cfg.effectiveSeed()
		r := rand.New(rand.NewSource(seed)) // #nosec G404 -- Using math/rand for deterministic property-based testing
		gen.SetShrinkStrategy(cfg.ShrinkStrat)

		t.Logf("[rapidx] seed=%d examples=%d maxshrink=%d strategy=%s parallelism=%d",
			seed, cfg.Examples, cfg.MaxShrink, cfg.ShrinkStrat, cfg.Parallelism)

		if cfg.Parallelism <= 1 {
			runSequential(t, cfg, g, body, seed, r)
		} else {
			runParallel(t, cfg, g, body, seed, r)
		}
	}
}

// runSequential executes property-based tests sequentially (single-threaded).
// It generates test cases one by one and runs them against the test function.
// If a test fails, it attempts to shrink the counterexample.
func runSequential[T any](t *testing.T, cfg Config, g gen.Generator[T], body func(*testing.T, T), seed int64, r *rand.Rand) {
	for i := 0; i < cfg.Examples; i++ {
		val, shrink := g.Generate(r, gen.Size{})
		name := fmt.Sprintf("ex#%d", i+1)

		passed := t.Run(name, func(st *testing.T) { body(st, val) })
		if passed {
			continue
		}

		min := val
		steps := 0
		acceptedPrev := true

		for steps < cfg.MaxShrink {
			next, ok := shrink(acceptedPrev)
			if !ok {
				break
			}
			steps++
			sname := fmt.Sprintf("%s/shrink#%d", name, steps)

			stillFails := !t.Run(sname, func(st *testing.T) { body(st, next) })
			if stillFails {
				min = next
				acceptedPrev = true
			} else {
				acceptedPrev = false
			}
		}

		full := fmt.Sprintf("^%s$/%s(/|$)", t.Name(), name)
		t.Fatalf("[rapidx] property failed; seed=%d; examples_run=%d; shrunk_steps=%d\n"+
			"counterexample (min): %#v\nreplay: go test -run '%s' -rapidx.seed=%d",
			seed, i+1, steps, min, full, seed)

		if cfg.StopOnFirstFailure {
			return
		}
	}
}

// runParallel executes property-based tests in parallel using multiple goroutines.
// It distributes test cases across multiple workers and collects failure results.
// The random number generator is protected by a mutex to ensure thread safety.
func runParallel[T any](t *testing.T, cfg Config, g gen.Generator[T], body func(*testing.T, T), seed int64, r *rand.Rand) {
	// Create a channel to distribute test indices to workers
	testChan := make(chan int, cfg.Examples)

	// Send all test indices to the channel
	for i := 0; i < cfg.Examples; i++ {
		testChan <- i
	}
	close(testChan)

	// WaitGroup to coordinate worker goroutines
	var wg sync.WaitGroup

	// Mutex to protect the shared random number generator
	var randMutex sync.Mutex

	// Channel to collect failure results from workers
	failureChan := make(chan failureResult, cfg.Examples)

	// Start worker goroutines
	for i := 0; i < cfg.Parallelism; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Process test cases from the channel
			for testIndex := range testChan {
				// Generate test case (protected by mutex for thread safety)
				randMutex.Lock()
				val, shrink := g.Generate(r, gen.Size{})
				randMutex.Unlock()

				name := fmt.Sprintf("ex#%d", testIndex+1)

				// Run the test case
				passed := t.Run(name, func(st *testing.T) { body(st, val) })
				if passed {
					continue
				}

				// Test failed, attempt to shrink the counterexample
				min := val
				steps := 0
				acceptedPrev := true

				for steps < cfg.MaxShrink {
					next, ok := shrink(acceptedPrev)
					if !ok {
						break
					}
					steps++
					sname := fmt.Sprintf("%s/shrink#%d", name, steps)

					stillFails := !t.Run(sname, func(st *testing.T) { body(st, next) })
					if stillFails {
						min = next
						acceptedPrev = true
					} else {
						acceptedPrev = false
					}
				}

				// Send failure result to the channel
				failureChan <- failureResult{
					testIndex: testIndex,
					name:      name,
					min:       min,
					steps:     steps,
				}

				if cfg.StopOnFirstFailure {
					return
				}
			}
		}(i)
	}

	// Close the failure channel when all workers are done
	go func() {
		wg.Wait()
		close(failureChan)
	}()

	// Process failure results and report them
	for failure := range failureChan {
		full := fmt.Sprintf("^%s$/%s(/|$)", t.Name(), failure.name)
		t.Fatalf("[rapidx] property failed; seed=%d; examples_run=%d; shrunk_steps=%d\n"+
			"counterexample (min): %#v\nreplay: go test -run '%s' -rapidx.seed=%d",
			seed, failure.testIndex+1, failure.steps, failure.min, full, seed)

		if cfg.StopOnFirstFailure {
			return
		}
	}
}

// failureResult holds information about a failed test case after shrinking.
type failureResult struct {
	// testIndex is the index of the test case that failed.
	testIndex int

	// name is the name of the test case.
	name string

	// min is the minimal counterexample found through shrinking.
	min interface{}

	// steps is the number of shrinking steps performed.
	steps int
}

// StateMachine represents a state machine for property-based testing.
// S is the state type, C is the command type.
type StateMachine[S, C any] struct {
	// InitialState is the starting state of the state machine.
	InitialState S

	// Commands defines the available commands that can be executed on the state machine.
	Commands []Command[S, C]
}

// Command represents a single command that can be executed on a state machine.
type Command[S, C any] struct {
	// Name is a human-readable name for the command.
	Name string

	// Generator creates instances of the command.
	Generator gen.Generator[C]

	// Execute applies the command to the current state and returns the new state.
	// If an error is returned, the command execution is considered failed.
	Execute func(S, C) (S, error)

	// Precondition determines if a command can be executed in the given state.
	// Commands that don't meet their precondition are skipped during execution.
	Precondition func(S, C) bool

	// Postcondition validates that the command execution was correct.
	// It receives the original state, the command, and the resulting state.
	// If it returns false, the test fails.
	Postcondition func(S, C, S) bool
}

// CommandSequence represents a sequence of commands to be executed on a state machine.
type CommandSequence[C any] struct {
	Commands []C
}

// StateMachineResult holds the result of executing a command sequence on a state machine.
type StateMachineResult[S, C any] struct {
	// FinalState is the state after executing all commands.
	FinalState S

	// ExecutionHistory contains the history of state transitions.
	ExecutionHistory []StateTransition[S, C]

	// SkippedCommands contains commands that were skipped due to precondition failures.
	SkippedCommands []C
}

// StateTransition represents a single state transition in the execution history.
type StateTransition[S, C any] struct {
	// Command is the command that was executed.
	Command C

	// FromState is the state before command execution.
	FromState S

	// ToState is the state after command execution.
	ToState S

	// Error is any error that occurred during command execution.
	Error error
}

// commandSequenceGenerator creates a generator for command sequences.
type commandSequenceGenerator[S, C any] struct {
	stateMachine StateMachine[S, C]
	maxLength    int
}

// Generate implements the Generator interface for command sequences.
func (g commandSequenceGenerator[S, C]) Generate(r *rand.Rand, sz gen.Size) (CommandSequence[C], gen.Shrinker[CommandSequence[C]]) {
	// Determine sequence length based on size constraints
	maxLen := g.maxLength
	if maxLen <= 0 {
		maxLen = sz.Max
		if maxLen <= 0 {
			maxLen = 10 // Default maximum length
		}
	}

	// Generate a random length between 0 and maxLen
	length := r.Intn(maxLen + 1)

	commands := make([]C, length)
	shrinkers := make([]gen.Shrinker[C], length)

	// Generate each command in the sequence
	for i := 0; i < length; i++ {
		// Select a random command type
		if len(g.stateMachine.Commands) == 0 {
			// No commands available, skip
			continue
		}
		cmdIndex := r.Intn(len(g.stateMachine.Commands))
		cmd := g.stateMachine.Commands[cmdIndex]

		// Generate the command
		cmdVal, cmdShrinker := cmd.Generator.Generate(r, sz)
		commands[i] = cmdVal
		shrinkers[i] = cmdShrinker
	}

	// If no commands were generated (because no commands are available), create empty sequence
	if len(g.stateMachine.Commands) == 0 {
		commands = make([]C, 0)
		shrinkers = make([]gen.Shrinker[C], 0)
	}

	sequence := CommandSequence[C]{Commands: commands}

	// Create a shrinker for the sequence
	shrinker := func(accept bool) (CommandSequence[C], bool) {
		// Try different shrinking strategies
		if len(commands) > 0 {
			// Strategy 1: Remove commands from the end
			newCommands := make([]C, len(commands)-1)
			copy(newCommands, commands[:len(commands)-1])
			newSequence := CommandSequence[C]{Commands: newCommands}
			return newSequence, true
		}

		// Strategy 2: Shrink individual commands
		for i := len(commands) - 1; i >= 0; i-- {
			if newCmd, ok := shrinkers[i](accept); ok {
				newCommands := make([]C, len(commands))
				copy(newCommands, commands)
				newCommands[i] = newCmd
				newSequence := CommandSequence[C]{Commands: newCommands}
				return newSequence, true
			}
		}

		return sequence, false
	}

	return sequence, shrinker
}

// findMatchingCommand finds a command that can handle the given command.
// This is a simplified implementation - in practice you'd want proper command type discrimination.
func findMatchingCommand[S, C any](sm StateMachine[S, C], cmd C) *Command[S, C] {
	if len(sm.Commands) == 0 {
		return nil
	}
	// For simplicity, we'll use the first command that can handle the command
	// In a real implementation, you might want to add command type discrimination
	return &sm.Commands[0]
}

// executeStateMachine executes a command sequence on a state machine and returns the result.
func executeStateMachine[S, C any](sm StateMachine[S, C], sequence CommandSequence[C]) StateMachineResult[S, C] {
	state := sm.InitialState
	history := make([]StateTransition[S, C], 0, len(sequence.Commands))
	skipped := make([]C, 0)

	for _, cmd := range sequence.Commands {
		// Find a command that can handle this command type
		matchedCmd := findMatchingCommand(sm, cmd)

		if matchedCmd == nil {
			// No commands available, skip
			skipped = append(skipped, cmd)
			continue
		}

		// Check precondition
		if matchedCmd.Precondition != nil && !matchedCmd.Precondition(state, cmd) {
			skipped = append(skipped, cmd)
			continue
		}

		// Execute the command
		fromState := state
		newState, err := matchedCmd.Execute(state, cmd)

		// Record the transition
		transition := StateTransition[S, C]{
			Command:   cmd,
			FromState: fromState,
			ToState:   newState,
			Error:     err,
		}
		history = append(history, transition)

		// Update state if no error occurred
		if err == nil {
			state = newState
		}
	}

	return StateMachineResult[S, C]{
		FinalState:       state,
		ExecutionHistory: history,
		SkippedCommands:  skipped,
	}
}

// TestStateMachine tests a state machine using property-based testing.
// It generates command sequences and validates that the state machine behaves correctly.
func TestStateMachine[S, C any](t *testing.T, sm StateMachine[S, C], cfg Config) {
	// Create a generator for command sequences
	seqGen := commandSequenceGenerator[S, C]{
		stateMachine: sm,
		maxLength:    20, // Default maximum sequence length
	}

	// Use the existing ForAll function to test the state machine
	ForAll(t, cfg, seqGen)(func(t *testing.T, sequence CommandSequence[C]) {
		result := executeStateMachine(sm, sequence)

		// Validate the execution result
		for _, transition := range result.ExecutionHistory {
			// Find the command that was executed
			var executedCmd *Command[S, C]
			for i := range sm.Commands {
				// For simplicity, we'll assume the first command matches
				// In a real implementation, you might want to add proper command matching
				executedCmd = &sm.Commands[i]
				break
			}

			if executedCmd != nil && executedCmd.Postcondition != nil {
				if !executedCmd.Postcondition(transition.FromState, transition.Command, transition.ToState) {
					t.Errorf("postcondition failed for command %s: from %v, cmd %v, to %v",
						executedCmd.Name, transition.FromState, transition.Command, transition.ToState)
				}
			}

			// Check that no unexpected errors occurred
			if transition.Error != nil {
				t.Errorf("unexpected error executing command %v: %v", transition.Command, transition.Error)
			}
		}
	})
}
