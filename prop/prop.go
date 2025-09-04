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
		r := rand.New(rand.NewSource(seed))
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
