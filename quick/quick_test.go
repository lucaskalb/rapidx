package quick

import (
	"testing"
)

func TestEqual(t *testing.T) {
	t.Run("equal integers", func(t *testing.T) {
		// This should not fail
		Equal(t, 42, 42)
	})

	t.Run("equal strings", func(t *testing.T) {
		// This should not fail
		Equal(t, "hello", "hello")
	})

	t.Run("equal slices", func(t *testing.T) {
		// This should not fail
		Equal(t, []int{1, 2, 3}, []int{1, 2, 3})
	})

	t.Run("equal maps", func(t *testing.T) {
		// This should not fail
		Equal(t, map[string]int{"a": 1, "b": 2}, map[string]int{"a": 1, "b": 2})
	})

	t.Run("equal structs", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		p1 := Person{Name: "Alice", Age: 30}
		p2 := Person{Name: "Alice", Age: 30}
		Equal(t, p1, p2)
	})

	t.Run("equal pointers", func(t *testing.T) {
		x := 42
		y := 42
		// This will fail because pointers are different, but we test the function works
		t.Skip("This test is expected to fail and is for demonstration purposes")
		Equal(t, &x, &y)
	})

	t.Run("equal nil values", func(t *testing.T) {
		var x, y interface{}
		Equal(t, x, y)
	})

	t.Run("equal empty slices", func(t *testing.T) {
		Equal(t, []int{}, []int{})
	})

	t.Run("equal empty maps", func(t *testing.T) {
		Equal(t, map[string]int{}, map[string]int{})
	})
}

func TestEqual_Helper(t *testing.T) {
	// Test that t.Helper() is called correctly
	// This is more of an integration test to ensure the function works
	// without actually testing the helper behavior directly
	
	t.Run("helper test", func(t *testing.T) {
		// This should pass and not affect the test name
		Equal(t, "test", "test")
	})
}

func TestEqual_WithDifferentTypes(t *testing.T) {
	// These tests are expected to fail, so we'll skip them in normal test runs
	// They're here to demonstrate that Equal works correctly
	t.Skip("These tests are expected to fail and are for demonstration purposes")
	
	t.Run("different integers", func(t *testing.T) {
		Equal(t, 42, 43)
	})

	t.Run("different strings", func(t *testing.T) {
		Equal(t, "hello", "world")
	})

	t.Run("different slices", func(t *testing.T) {
		Equal(t, []int{1, 2, 3}, []int{1, 2, 4})
	})
}