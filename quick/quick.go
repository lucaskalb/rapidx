// Package quick provides quick testing utilities for Go.
// It includes helper functions for common testing patterns, particularly
// for value comparison and assertion utilities.
package quick

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Equal compares two values of the same type and fails the test if they are not equal.
// It uses go-cmp for deep comparison and provides detailed diff output when values differ.
// The function calls t.Helper() to mark itself as a test helper function.
//
// Parameters:
//   - t: The testing.T instance for the current test
//   - got: The actual value obtained from the code under test
//   - want: The expected value
//
// Example usage:
//
//	quick.Equal(t, result, expected)
//	quick.Equal(t, []int{1, 2, 3}, []int{1, 2, 3})
//	quick.Equal(t, map[string]int{"a": 1}, map[string]int{"a": 1})
func Equal[T any](t *testing.T, got, want T) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}
