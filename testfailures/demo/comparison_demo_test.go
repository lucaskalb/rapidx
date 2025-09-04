//go:build demo
// +build demo

// Package demo contains demonstration tests that are designed to fail intentionally.
// These tests showcase the shrinking mechanism and property-based testing capabilities
// of the rapidx library. They are meant for educational and demonstration purposes.
package demo

import (
	"testing"

	"github.com/lucaskalb/rapidx/quick"
)

// TestEqual_WithDifferentTypes tests the Equal function with different values
// to demonstrate that it correctly identifies unequal values and fails appropriately.
// These tests are skipped in normal runs as they are expected to fail.
func TestEqual_WithDifferentTypes(t *testing.T) {
	// These tests are expected to fail, so we'll skip them in normal test runs
	// They're here to demonstrate that Equal works correctly
	t.Skip("These tests are expected to fail and are for demonstration purposes")

	t.Run("different integers", func(t *testing.T) {
		// This should fail because 42 != 43
		quick.Equal(t, 42, 43)
	})

	t.Run("different strings", func(t *testing.T) {
		// This should fail because "hello" != "world"
		quick.Equal(t, "hello", "world")
	})

	t.Run("different slices", func(t *testing.T) {
		// This should fail because the slices have different elements
		quick.Equal(t, []int{1, 2, 3}, []int{1, 2, 4})
	})
}

// TestEqual_PointerComparison demonstrates pointer comparison behavior.
// This test shows that pointer comparison fails even when values are equal.
func TestEqual_PointerComparison(t *testing.T) {
	t.Run("equal pointers", func(t *testing.T) {
		// Test pointer comparison (this will fail because pointers are different)
		x := 42
		y := 42
		// This will fail because pointers are different, but we test the function works
		t.Skip("This test is expected to fail and is for demonstration purposes")
		quick.Equal(t, &x, &y)
	})
}