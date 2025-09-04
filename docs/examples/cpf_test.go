// Package examples demonstrates how to use the rapidx property-based testing library.
// These examples show various testing patterns and how the shrinking mechanism
// helps find minimal counterexamples when properties fail.
package examples

import (
	"testing"

	"github.com/lucaskalb/rapidx/gen/domain"
	"github.com/lucaskalb/rapidx/prop"
	"github.com/lucaskalb/rapidx/quick"
)

// Test_CPF_AlwaysValid demonstrates a property-based test for CPF generation.
// This test verifies that all generated CPF numbers are valid according to
// the CPF validation algorithm, and that the UnmaskCPF function is idempotent.
func Test_CPF_AlwaysValid(t *testing.T) {
	cfg := prop.Default()
	prop.ForAll(t, cfg, domain.CPF(false))(func(t *testing.T, cpf string) {
		if !domain.ValidCPF(cpf) {
			t.Fatalf("valid CPF generated was rejected: %q", cpf)
		}
		n1 := domain.UnmaskCPF(cpf)
		n2 := domain.UnmaskCPF(n1)
		quick.Equal(t, n1, n2)
	})
}

// Test_CPF_MaskUnmaskRoundTrip demonstrates testing the round-trip property
// of CPF masking and unmasking operations. This test verifies that
// unmasking a masked CPF and then masking it again produces the same result.
func Test_CPF_MaskUnmaskRoundTrip(t *testing.T) {
	prop.ForAll(t, prop.Default(), domain.CPF(true))(func(t *testing.T, masked string) {
		raw := domain.UnmaskCPF(masked)
		back := domain.UnmaskCPF(domain.MaskCPF(raw))
		quick.Equal(t, raw, back)
	})
}

// Test_CPF_Any_Valid demonstrates testing CPFAny() generator which produces
// CPF numbers with random masking (50/50 chance of masked or unmasked).
// This test verifies that all generated CPF numbers are valid regardless of format.
func Test_CPF_Any_Valid(t *testing.T) {
	prop.ForAll(t, prop.Default(), domain.CPFAny())(func(t *testing.T, s string) {
		if !domain.ValidCPF(s) {
			t.Fatalf("valid CPF generated was rejected: %q", s)
		}
	})
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
