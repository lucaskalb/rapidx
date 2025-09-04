// Package gen provides generators for property-based testing in Go.
// It includes generators for various data types and utilities for creating
// custom generators with shrinking capabilities.
package gen

import "math/rand"

// Size controls the scale and limits of generators.
// It defines the minimum and maximum bounds for generated values.
type Size struct {
	// Min is the minimum bound for generated values.
	Min int
	// Max is the maximum bound for generated values.
	Max int
}

// Shrinker proposes "smaller" candidates during the shrinking process.
// The accept parameter indicates whether the PREVIOUS candidate was accepted
// (i.e., it reproduced the failure). This allows the shrinker to "rebase"
// and generate new neighbors from the new minimum.
type Shrinker[T any] func(accept bool) (next T, ok bool)

// Generator is the public contract for all generators.
// It defines the interface that all generators must implement.
type Generator[T any] interface {
	// Generate produces a value and a shrinker function for that value.
	// The random number generator and size constraints are provided as parameters.
	Generate(r *rand.Rand, sz Size) (value T, shrink Shrinker[T])
}

// Shrinking strategy constants
const (
	ShrinkStrategyBFS = "bfs" // breadth-first search
	ShrinkStrategyDFS = "dfs" // depth-first search
)

// shrinkStrategy holds the current shrinking strategy.
// It can be either "bfs" (breadth-first search) or "dfs" (depth-first search).
var shrinkStrategy = ShrinkStrategyBFS

// SetShrinkStrategy sets the shrinking strategy for all generators.
// Valid strategies are "dfs" (depth-first search) and "bfs" (breadth-first search).
// Any other value defaults to "bfs".
func SetShrinkStrategy(s string) {
	if s == ShrinkStrategyDFS {
		shrinkStrategy = ShrinkStrategyDFS
	} else {
		shrinkStrategy = ShrinkStrategyBFS
	}
}

// GetShrinkStrategy returns the current shrinking strategy.
func GetShrinkStrategy() string {
	return shrinkStrategy
}

// T is an optional alias for Generator[T] for compatibility.
type T[T any] = Generator[T]

// GenFunc is a helper type for creating a Generator from a closure.
// It wraps a function that implements the Generator interface.
type GenFunc[T any] struct {
	fn func(r *rand.Rand, sz Size) (T, Shrinker[T])
}

// Generate implements the Generator interface for GenFunc.
func (g GenFunc[T]) Generate(r *rand.Rand, sz Size) (T, Shrinker[T]) {
	return g.fn(r, sz)
}

// From creates a Generator from a function that implements the Generator interface.
// This is a convenience function for creating custom generators.
func From[T any](fn func(*rand.Rand, Size) (T, Shrinker[T])) Generator[T] {
	return GenFunc[T]{fn: fn}
}
