//go:build demo
// +build demo

// Package demo contains demonstration tests that are designed to fail intentionally.
// These tests showcase the shrinking mechanism and property-based testing capabilities
// of the rapidx library. They are meant for educational and demonstration purposes.
package demo

import (
	"testing"

	"github.com/lucaskalb/rapidx/gen"
	"github.com/lucaskalb/rapidx/gen/domain"
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

// Test_CPF_Invalid demonstrates a property-based test that is designed to fail.
// This test expects all CPF numbers to start with '9', which is not true for
// valid CPF generation. This example shows how the shrinking mechanism will
// find a minimal counterexample when the property fails.
func Test_CPF_Invalid(t *testing.T) {
	cfg := prop.Default()
	prop.ForAll(t, cfg, domain.CPF(false))(func(t *testing.T, cpf string) {
		if cpf[0] != '9' {
			t.Fatalf("expected to start with 9, but got %q", cpf)
		}
	})
}