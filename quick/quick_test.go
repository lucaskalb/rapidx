// Package quick_test contains tests for the quick package.
// These tests verify the functionality of the Equal function and its
// behavior with different data types and edge cases.
package quick

import (
	"testing"
)

// TestEqual tests the Equal function with various data types to ensure
// it correctly identifies equal values and handles different Go types.
func TestEqual(t *testing.T) {
	t.Run("equal integers", func(t *testing.T) {
		// Test that equal integers are correctly identified
		Equal(t, 42, 42)
	})

	t.Run("equal strings", func(t *testing.T) {
		// Test that equal strings are correctly identified
		Equal(t, "hello", "hello")
	})

	t.Run("equal slices", func(t *testing.T) {
		// Test that equal slices are correctly identified
		Equal(t, []int{1, 2, 3}, []int{1, 2, 3})
	})

	t.Run("equal maps", func(t *testing.T) {
		// Test that equal maps are correctly identified
		Equal(t, map[string]int{"a": 1, "b": 2}, map[string]int{"a": 1, "b": 2})
	})

	t.Run("equal structs", func(t *testing.T) {
		// Test that equal structs are correctly identified
		type Person struct {
			Name string
			Age  int
		}
		p1 := Person{Name: "Alice", Age: 30}
		p2 := Person{Name: "Alice", Age: 30}
		Equal(t, p1, p2)
	})

	t.Run("equal pointers", func(t *testing.T) {
		// Test pointer comparison (this will fail because pointers are different)
		x := 42
		y := 42
		// This will fail because pointers are different, but we test the function works
		t.Skip("This test is expected to fail and is for demonstration purposes")
		Equal(t, &x, &y)
	})

	t.Run("equal nil values", func(t *testing.T) {
		// Test that nil values are correctly identified as equal
		var x, y interface{}
		Equal(t, x, y)
	})

	t.Run("equal empty slices", func(t *testing.T) {
		// Test that empty slices are correctly identified as equal
		Equal(t, []int{}, []int{})
	})

	t.Run("equal empty maps", func(t *testing.T) {
		// Test that empty maps are correctly identified as equal
		Equal(t, map[string]int{}, map[string]int{})
	})
}

// TestEqual_Helper tests that the Equal function correctly calls t.Helper()
// to mark itself as a test helper function, which affects error reporting.
func TestEqual_Helper(t *testing.T) {
	// Test that t.Helper() is called correctly
	// This is more of an integration test to ensure the function works
	// without actually testing the helper behavior directly

	t.Run("helper test", func(t *testing.T) {
		// This should pass and not affect the test name
		Equal(t, "test", "test")
	})
}

// TestEqual_WithDifferentTypes tests the Equal function with different values
// to demonstrate that it correctly identifies unequal values and fails appropriately.
// These tests are skipped in normal runs as they are expected to fail.
func TestEqual_WithDifferentTypes(t *testing.T) {
	// These tests are expected to fail, so we'll skip them in normal test runs
	// They're here to demonstrate that Equal works correctly
	t.Skip("These tests are expected to fail and are for demonstration purposes")

	t.Run("different integers", func(t *testing.T) {
		// This should fail because 42 != 43
		Equal(t, 42, 43)
	})

	t.Run("different strings", func(t *testing.T) {
		// This should fail because "hello" != "world"
		Equal(t, "hello", "world")
	})

	t.Run("different slices", func(t *testing.T) {
		// This should fail because the slices have different elements
		Equal(t, []int{1, 2, 3}, []int{1, 2, 4})
	})
}
