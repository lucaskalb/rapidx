package examples

import (
	"testing"

	"github.com/lucaskalb/rapidx/gen"
	"github.com/lucaskalb/rapidx/prop"
	"github.com/lucaskalb/rapidx/quick"
)

func Test_CPF_AlwaysValid(t *testing.T) {
	cfg := prop.Default()
	prop.ForAll(t, cfg, gen.CPF(false))(func(t *testing.T, cpf string) {
		if !gen.ValidCPF(cpf) {
			t.Fatalf("cpf válido gerado foi rejeitado: %q", cpf)
		}
		n1 := gen.UnmaskCPF(cpf)
		n2 := gen.UnmaskCPF(n1)
		quick.Equal(t, n1, n2)
	})
}

func Test_CPF_MaskUnmaskRoundTrip(t *testing.T) {
	prop.ForAll(t, prop.Default(), gen.CPF(true))(func(t *testing.T, masked string) {
		raw := gen.UnmaskCPF(masked)
		back := gen.UnmaskCPF(gen.MaskCPF(raw))
		quick.Equal(t, raw, back)
	})
}

func Test_CPF_Any_Valid(t *testing.T) {
	prop.ForAll(t, prop.Default(), gen.CPFAny())(func(t *testing.T, s string) {
		if !gen.ValidCPF(s) {
			t.Fatalf("cpf válido gerado foi rejeitado: %q", s)
		}
	})
}

func Test_CPF_Invalid(t *testing.T) {
	cfg := prop.Default()
	prop.ForAll(t, cfg, gen.CPF(false))(func(t *testing.T, cpf string) {
			if cpf[0] != '9' {
					t.Fatalf("esperava começar com 9, mas veio %q", cpf)
			}
	})
}

