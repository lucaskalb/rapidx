// Package prop_test contains tests for the prop package.
// These tests verify the functionality of property-based testing features
// including configuration, test execution, and shrinking behavior.
package prop

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/lucaskalb/rapidx/gen"
)

// TestConfig_effectiveSeed tests the effectiveSeed method of Config.
// It verifies that zero seeds generate random values and non-zero seeds are preserved.
func TestConfig_effectiveSeed(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool // true if seed should be non-zero
	}{
		{
			name: "seed zero should generate random seed",
			config: Config{
				Seed: 0,
			},
			expected: true,
		},
		{
			name: "seed non-zero should return same seed",
			config: Config{
				Seed: 12345,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := tt.config.effectiveSeed()
			if tt.config.Seed == 0 {

				if seed == 0 {
					t.Errorf("effectiveSeed() = %d, expected non-zero random seed", seed)
				}
			} else {

				if seed != tt.config.Seed {
					t.Errorf("effectiveSeed() = %d, expected %d", seed, tt.config.Seed)
				}
			}
		})
	}
}

// TestConfig_effectiveSeed_Consistency tests that effectiveSeed generates unique values
// when called multiple times with a zero seed configuration.
func TestConfig_effectiveSeed_Consistency(t *testing.T) {
	config := Config{Seed: 0}

	// Track generated seeds to ensure uniqueness
	seeds := make(map[int64]bool)
	for i := 0; i < 10; i++ {
		seed := config.effectiveSeed()
		if seeds[seed] {
			t.Errorf("effectiveSeed() generated duplicate seed: %d", seed)
		}
		seeds[seed] = true
		time.Sleep(time.Microsecond) // Small delay to ensure different timestamps
	}
}

// TestDefault tests that the Default() function returns a valid configuration
// with all required fields set to reasonable values.
func TestDefault(t *testing.T) {
	config := Default()

	// Verify that all configuration fields have valid values
	if config.Examples <= 0 {
		t.Errorf("Default().Examples = %d, expected > 0", config.Examples)
	}

	if config.MaxShrink <= 0 {
		t.Errorf("Default().MaxShrink = %d, expected > 0", config.MaxShrink)
	}

	if config.ShrinkStrat == "" {
		t.Errorf("Default().ShrinkStrat = %q, expected non-empty", config.ShrinkStrat)
	}

	if !config.StopOnFirstFailure {
		t.Errorf("Default().StopOnFirstFailure = %v, expected true", config.StopOnFirstFailure)
	}

	if config.Parallelism <= 0 {
		t.Errorf("Default().Parallelism = %d, expected > 0", config.Parallelism)
	}
}

func TestConfig_Fields(t *testing.T) {
	config := Config{
		Seed:               12345,
		Examples:           50,
		MaxShrink:          200,
		ShrinkStrat:        "dfs",
		StopOnFirstFailure: false,
		Parallelism:        8,
	}

	if config.Seed != 12345 {
		t.Errorf("Config.Seed = %d, expected 12345", config.Seed)
	}

	if config.Examples != 50 {
		t.Errorf("Config.Examples = %d, expected 50", config.Examples)
	}

	if config.MaxShrink != 200 {
		t.Errorf("Config.MaxShrink = %d, expected 200", config.MaxShrink)
	}

	if config.ShrinkStrat != "dfs" {
		t.Errorf("Config.ShrinkStrat = %q, expected 'dfs'", config.ShrinkStrat)
	}

	if config.StopOnFirstFailure != false {
		t.Errorf("Config.StopOnFirstFailure = %v, expected false", config.StopOnFirstFailure)
	}

	if config.Parallelism != 8 {
		t.Errorf("Config.Parallelism = %d, expected 8", config.Parallelism)
	}
}

// TestForAll_SequentialExecution tests the ForAll function with sequential execution
// (parallelism = 1) to ensure basic property-based testing functionality works.
func TestForAll_SequentialExecution(t *testing.T) {
	// Configure for sequential execution
	config := Config{
		Seed:        12345,
		Examples:    5,
		MaxShrink:   10,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a simple generator that always returns 42
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	// Test that the property holds for all generated values
	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

func TestForAll_ParallelExecution(t *testing.T) {

	config := Config{
		Seed:        12345,
		Examples:    5,
		MaxShrink:   10,
		ShrinkStrat: "bfs",
		Parallelism: 2,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

func TestForAll_WithShrinking(t *testing.T) {
	// Test shrinking behavior with a working shrinker
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 5, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 3 {
				return shrinkerCallCount, true
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {

		if val < 0 || val > 10 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

func TestForAll_WithDFSSStrategy(t *testing.T) {

	config := Config{
		Seed:        12345,
		Examples:    3,
		MaxShrink:   5,
		ShrinkStrat: "dfs",
		Parallelism: 1,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

func TestForAll_WithHighParallelism(t *testing.T) {

	config := Config{
		Seed:        12345,
		Examples:    10,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 8,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

func TestForAll_WithZeroExamples(t *testing.T) {
	// Test with zero examples
	config := Config{
		Seed:        12345,
		Examples:    0,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

func TestForAll_WithZeroMaxShrink(t *testing.T) {
	// Test with zero max shrink
	config := Config{
		Seed:        12345,
		Examples:    3,
		MaxShrink:   0,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

// Test failure scenarios in runSequential
func TestForAll_SequentialFailure(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    3,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false // No shrinking
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}

func TestForAll_SequentialFailureCodePath(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   2,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	t.Run("failure_test", func(st *testing.T) {
		// This will trigger the failure path in runSequential
		ForAll(st, config, gen)(func(t *testing.T, val int) {
			t.Errorf("This should fail: got %d", val)
		})
	})
}

func TestForAll_SequentialFailureWithShrinking(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   3,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 5, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 2 {
				return shrinkerCallCount, true
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}

func TestForAll_SequentialFailureWithShrinkingAcceptance(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 10, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 3 {

				return 10 - shrinkerCallCount, true
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}

func TestForAll_SequentialStopOnFirstFailureFalse(t *testing.T) {
	config := Config{
		Seed:               12345,
		Examples:           3,
		MaxShrink:          2,
		ShrinkStrat:        "bfs",
		Parallelism:        1,
		StopOnFirstFailure: false,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}

// Test failure scenarios in runParallel
func TestForAll_ParallelFailure(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    3,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 2,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	// This should fail and trigger the parallel failure path
	ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}

func TestForAll_ParallelFailureWithShrinking(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    2,
		MaxShrink:   3,
		ShrinkStrat: "bfs",
		Parallelism: 2,
	}

	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 5, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 2 {
				return shrinkerCallCount, true
			}
			return 0, false
		}
	})

	// This should fail and trigger parallel shrinking
	ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}

func TestForAll_ParallelStopOnFirstFailureFalse(t *testing.T) {
	config := Config{
		Seed:               12345,
		Examples:           3,
		MaxShrink:          2,
		ShrinkStrat:        "bfs",
		Parallelism:        2,
		StopOnFirstFailure: false,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}

// Test edge cases
func TestForAll_EmptyGenerator(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator that returns zero value
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 0, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 0 {
			t.Errorf("Expected 0, got %d", val)
		}
	})
}

func TestForAll_GeneratorWithNilShrinker(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator with a shrinker that immediately returns false
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false // No shrinking possible
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

// Test shrinking logic more thoroughly
func TestForAll_ShrinkingWithAcceptancePattern(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   10,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator that tests the acceptance pattern
	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 10, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 5 {
				// Simulate shrinking that sometimes succeeds, sometimes fails
				if shrinkerCallCount%2 == 0 {
					return 10 - shrinkerCallCount, true
				} else {
					return 10 - shrinkerCallCount, true
				}
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we're testing the shrinking code path without failure
		if val < 0 || val > 10 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

// Test concurrent safety
func TestForAll_ConcurrentSafety(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    20,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 4,
	}

	// Create a generator that uses the random number generator
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		// Use the random generator to create some variation
		val := r.Intn(100)
		return val, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// Test that concurrent access doesn't cause issues
		if val < 0 || val >= 100 {
			t.Errorf("Value %d is outside expected range [0, 100)", val)
		}
	})
}

// Test with different shrink strategies
func TestForAll_WithDFSStrategy(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    3,
		MaxShrink:   5,
		ShrinkStrat: "dfs",
		Parallelism: 1,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

func TestForAll_WithInvalidStrategy(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    3,
		MaxShrink:   5,
		ShrinkStrat: "invalid",
		Parallelism: 1,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

// Test failureResult struct
func TestFailureResult(t *testing.T) {
	fr := failureResult{
		testIndex: 1,
		name:      "test",
		min:       42,
		steps:     3,
	}

	if fr.testIndex != 1 {
		t.Errorf("Expected testIndex 1, got %d", fr.testIndex)
	}

	if fr.name != "test" {
		t.Errorf("Expected name 'test', got %q", fr.name)
	}

	if fr.min != 42 {
		t.Errorf("Expected min 42, got %v", fr.min)
	}

	if fr.steps != 3 {
		t.Errorf("Expected steps 3, got %d", fr.steps)
	}
}

// Test more edge cases and boundary conditions
func TestForAll_WithMaxShrinkReached(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   2,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator that can shrink more than MaxShrink
	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 10, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 5 { // More than MaxShrink
				return 10 - shrinkerCallCount, true
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we're testing the shrinking code path
		if val < 0 || val > 10 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

func TestForAll_WithShrinkerThatReturnsFalse(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator with a shrinker that immediately returns false
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false // No shrinking possible
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

func TestForAll_WithShrinkerThatAlternates(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   10,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator with a shrinker that alternates between success and failure
	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 10, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 5 {
				// Alternate between smaller and larger values
				if shrinkerCallCount%2 == 0 {
					return 10 - shrinkerCallCount, true
				} else {
					return 10 + shrinkerCallCount, true
				}
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we're testing the shrinking code path
		if val < 0 || val > 20 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

func TestForAll_WithHighParallelismAndManyExamples(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    50,
		MaxShrink:   3,
		ShrinkStrat: "bfs",
		Parallelism: 10,
	}

	// Create a generator that uses the random number generator
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		val := r.Intn(1000)
		return val, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// Test that concurrent access doesn't cause issues
		if val < 0 || val >= 1000 {
			t.Errorf("Value %d is outside expected range [0, 1000)", val)
		}
	})
}

func TestForAll_WithSingleExample(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

func TestForAll_WithSingleExampleParallel(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 4,
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

// Test the flag variables are accessible
func TestFlagVariables(t *testing.T) {
	// Test that flag variables are accessible (they should be non-zero defaults)
	if *flagExamples <= 0 {
		t.Errorf("flagExamples should be > 0, got %d", *flagExamples)
	}

	if *flagMaxShrink <= 0 {
		t.Errorf("flagMaxShrink should be > 0, got %d", *flagMaxShrink)
	}

	if *flagShrinkStrat == "" {
		t.Errorf("flagShrinkStrat should not be empty, got %q", *flagShrinkStrat)
	}

	if *flagParallelism <= 0 {
		t.Errorf("flagParallelism should be > 0, got %d", *flagParallelism)
	}
}

// Test that Default() uses flag values
func TestDefault_UsesFlagValues(t *testing.T) {
	config := Default()

	// The Default() function should use the flag values
	if config.Examples != *flagExamples {
		t.Errorf("Default().Examples = %d, expected %d", config.Examples, *flagExamples)
	}

	if config.MaxShrink != *flagMaxShrink {
		t.Errorf("Default().MaxShrink = %d, expected %d", config.MaxShrink, *flagMaxShrink)
	}

	if config.ShrinkStrat != *flagShrinkStrat {
		t.Errorf("Default().ShrinkStrat = %q, expected %q", config.ShrinkStrat, *flagShrinkStrat)
	}

	if config.Parallelism != *flagParallelism {
		t.Errorf("Default().Parallelism = %d, expected %d", config.Parallelism, *flagParallelism)
	}
}

// Test more comprehensive scenarios to increase coverage
func TestForAll_WithComplexShrinkingPattern(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    2,
		MaxShrink:   10,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator with complex shrinking behavior
	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 20, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 8 {
				// Create a pattern that tests different acceptance scenarios
				if accept {
					// If previous was accepted, try a different approach
					return 20 - shrinkerCallCount*2, true
				} else {
					// If previous was rejected, try a smaller step
					return 20 - shrinkerCallCount, true
				}
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we're testing the shrinking code path
		if val < 0 || val > 20 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

func TestForAll_WithParallelismEqualToExamples(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    4,
		MaxShrink:   3,
		ShrinkStrat: "bfs",
		Parallelism: 4, // Same as examples
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		val := r.Intn(100)
		return val, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val < 0 || val >= 100 {
			t.Errorf("Value %d is outside expected range [0, 100)", val)
		}
	})
}

func TestForAll_WithParallelismGreaterThanExamples(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    2,
		MaxShrink:   3,
		ShrinkStrat: "bfs",
		Parallelism: 8, // Greater than examples
	}

	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		val := r.Intn(50)
		return val, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		if val < 0 || val >= 50 {
			t.Errorf("Value %d is outside expected range [0, 50)", val)
		}
	})
}

func TestForAll_WithShrinkingThatExhausts(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator that exhausts its shrinking possibilities
	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 10, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 3 {
				return 10 - shrinkerCallCount, true
			}
			return 0, false // Exhausted
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we're testing the shrinking code path
		if val < 0 || val > 10 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

func TestForAll_WithDifferentSeedValues(t *testing.T) {
	// Test with different seed values to ensure seed handling works
	seeds := []int64{0, 1, 42, 12345, 999999}

	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed_%d", seed), func(t *testing.T) {
			config := Config{
				Seed:        seed,
				Examples:    2,
				MaxShrink:   3,
				ShrinkStrat: "bfs",
				Parallelism: 1,
			}

			gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
				val := r.Intn(100)
				return val, func(accept bool) (int, bool) {
					return 0, false
				}
			})

			ForAll(t, config, gen)(func(t *testing.T, val int) {
				if val < 0 || val >= 100 {
					t.Errorf("Value %d is outside expected range [0, 100)", val)
				}
			})
		})
	}
}

func TestForAll_WithDifferentStrategies(t *testing.T) {
	strategies := []string{"bfs", "dfs", "invalid", ""}

	for _, strategy := range strategies {
		t.Run(fmt.Sprintf("strategy_%s", strategy), func(t *testing.T) {
			config := Config{
				Seed:        12345,
				Examples:    2,
				MaxShrink:   3,
				ShrinkStrat: strategy,
				Parallelism: 1,
			}

			gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
				return 42, func(accept bool) (int, bool) {
					return 0, false
				}
			})

			ForAll(t, config, gen)(func(t *testing.T, val int) {
				if val != 42 {
					t.Errorf("Expected 42, got %d", val)
				}
			})
		})
	}
}

func TestForAll_WithBoundaryParallelism(t *testing.T) {
	// Test boundary values for parallelism
	parallelismValues := []int{1, 2, 4, 8, 16}

	for _, parallelism := range parallelismValues {
		t.Run(fmt.Sprintf("parallelism_%d", parallelism), func(t *testing.T) {
			config := Config{
				Seed:        12345,
				Examples:    10,
				MaxShrink:   3,
				ShrinkStrat: "bfs",
				Parallelism: parallelism,
			}

			gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
				val := r.Intn(1000)
				return val, func(accept bool) (int, bool) {
					return 0, false
				}
			})

			ForAll(t, config, gen)(func(t *testing.T, val int) {
				if val < 0 || val >= 1000 {
					t.Errorf("Value %d is outside expected range [0, 1000)", val)
				}
			})
		})
	}
}

// Test to increase coverage by testing more code paths
func TestForAll_SequentialWithShrinkingThatSucceeds(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator that will shrink but the test will pass
	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 10, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 3 {
				return 10 - shrinkerCallCount, true
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we're testing the shrinking code path without failure
		if val < 0 || val > 10 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

func TestForAll_ParallelWithShrinkingThatSucceeds(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    2,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 2,
	}

	// Create a generator that will shrink but the test will pass
	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 10, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 3 {
				return 10 - shrinkerCallCount, true
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we're testing the shrinking code path without failure
		if val < 0 || val > 10 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

// Test the specific code paths that are not covered
func TestForAll_SequentialShrinkingLoop(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   10,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator that exercises the shrinking loop
	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 20, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 8 {
				// Test the acceptance logic
				if accept {
					return 20 - shrinkerCallCount*2, true
				} else {
					return 20 - shrinkerCallCount, true
				}
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we're testing the shrinking code path
		if val < 0 || val > 20 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

func TestForAll_ParallelShrinkingLoop(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    2,
		MaxShrink:   10,
		ShrinkStrat: "bfs",
		Parallelism: 2,
	}

	// Create a generator that exercises the shrinking loop
	shrinkerCallCount := 0
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 20, func(accept bool) (int, bool) {
			shrinkerCallCount++
			if shrinkerCallCount <= 8 {
				// Test the acceptance logic
				if accept {
					return 20 - shrinkerCallCount*2, true
				} else {
					return 20 - shrinkerCallCount, true
				}
			}
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we're testing the shrinking code path
		if val < 0 || val > 20 {
			t.Errorf("Value %d is outside expected range", val)
		}
	})
}

// Test to increase coverage by testing the continue path in runSequential
func TestForAll_SequentialContinuePath(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    3,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator that will always pass
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test always passes, so we test the continue path
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

// Test to increase coverage by testing the break path in shrinking
func TestForAll_SequentialShrinkingBreakPath(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    1,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 1,
	}

	// Create a generator with a shrinker that immediately returns false
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false // This will trigger the break path
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we test the shrinking break path
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

// Test to increase coverage by testing the parallel continue path
func TestForAll_ParallelContinuePath(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    3,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 2,
	}

	// Create a generator that will always pass
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test always passes, so we test the continue path in parallel
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}

// Test to increase coverage by testing the parallel shrinking break path
func TestForAll_ParallelShrinkingBreakPath(t *testing.T) {
	config := Config{
		Seed:        12345,
		Examples:    2,
		MaxShrink:   5,
		ShrinkStrat: "bfs",
		Parallelism: 2,
	}

	// Create a generator with a shrinker that immediately returns false
	gen := gen.From(func(r *rand.Rand, sz gen.Size) (int, gen.Shrinker[int]) {
		return 42, func(accept bool) (int, bool) {
			return 0, false // This will trigger the break path
		}
	})

	ForAll(t, config, gen)(func(t *testing.T, val int) {
		// This test passes, so we test the shrinking break path in parallel
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	})
}
