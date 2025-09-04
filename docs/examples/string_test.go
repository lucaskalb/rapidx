//go:build examples
// +build examples

// Package examples demonstrates how to use the rapidx property-based testing library.
// These examples show various testing patterns and how the shrinking mechanism
// helps find minimal counterexamples when properties fail.
package examples

import (
	"testing"

	"github.com/lucaskalb/rapidx/gen"
	"github.com/lucaskalb/rapidx/prop"
)

// Test_String_FalsaRegra demonstrates a property-based test that is designed to fail.
// This test verifies a false property: "all generated strings are empty".
// This example shows how the shrinking mechanism will find a minimal counterexample
// when the property fails, helping developers understand why their assumptions are incorrect.
func Test_String_FalsaRegra(t *testing.T) {

	prop.ForAll(t, prop.Default(), gen.StringAlphaNum(gen.Size{Min: 0, Max: 32}))(
		func(t *testing.T, s string) {
			if s != "" {
				t.Fatalf("expected empty string, got %q", s)
			}
		},
	)
}
