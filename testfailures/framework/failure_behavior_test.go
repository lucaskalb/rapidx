//go:build demo
// +build demo

// Package framework contains tests that verify the framework's behavior
// when properties fail intentionally. These tests ensure that the framework
// correctly handles failures, shrinking, and parallel execution paths.
package framework

import (
	"math/rand"
	"testing"

	"github.com/lucaskalb/rapidx/gen"
	"github.com/lucaskalb/rapidx/prop"
)

// TestForAll_SequentialFailureCodePath tests the sequential failure code path.
// This test verifies that the framework correctly handles failures in sequential mode.
func TestForAll_SequentialFailureCodePath(t *testing.T) {
	config := prop.Config{
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
		prop.ForAll(st, config, gen)(func(t *testing.T, val int) {
			t.Errorf("This should fail: got %d", val)
		})
	})
}

// TestForAll_SequentialFailureWithShrinking tests sequential failure with shrinking.
// This test verifies that the framework correctly handles shrinking in sequential mode.
func TestForAll_SequentialFailureWithShrinking(t *testing.T) {
	config := prop.Config{
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

	prop.ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}

// TestForAll_SequentialFailureWithShrinkingAcceptance tests sequential failure
// with shrinking and acceptance behavior.
func TestForAll_SequentialFailureWithShrinkingAcceptance(t *testing.T) {
	config := prop.Config{
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

	prop.ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}

// TestForAll_SequentialStopOnFirstFailureFalse tests sequential execution
// with StopOnFirstFailure set to false.
func TestForAll_SequentialStopOnFirstFailureFalse(t *testing.T) {
	config := prop.Config{
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

	prop.ForAll(t, config, gen)(func(t *testing.T, val int) {
		t.Errorf("This should fail: got %d", val)
	})
}