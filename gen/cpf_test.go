package gen

import (
	"math/rand"
	"strings"
	"testing"
)

func TestCPF(t *testing.T) {
	gen := CPF(false) // unmasked
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	if len(value) != 11 {
		t.Errorf("CPF().Generate() = %q (len=%d), expected length 11", value, len(value))
	}

	if shrink == nil {
		t.Error("CPF().Generate() returned nil shrinker")
	}
}

func TestCPFAny(t *testing.T) {
	gen := CPFAny()
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	if len(value) != 11 {
		t.Errorf("CPFAny().Generate() = %q (len=%d), expected length 11", value, len(value))
	}

	if shrink == nil {
		t.Error("CPFAny().Generate() returned nil shrinker")
	}
}

func TestValidCPF(t *testing.T) {

	valid := ValidCPF("11144477735")
	if !valid {
		t.Error("ValidCPF() should return true for valid CPF")
	}

	invalid := ValidCPF("11111111111")
	if invalid {
		t.Error("ValidCPF() should return false for invalid CPF")
	}
}

func TestMaskCPF(t *testing.T) {
	cpf := "12345678901"
	masked := MaskCPF(cpf)

	if len(masked) != 14 {
		t.Errorf("MaskCPF() = %q (len=%d), expected length 14", masked, len(masked))
	}

	if !strings.Contains(masked, ".") || !strings.Contains(masked, "-") {
		t.Errorf("MaskCPF() = %q, expected to contain dots and dashes", masked)
	}
}

func TestUnmaskCPF(t *testing.T) {
	masked := "123.456.789-01"
	unmasked := UnmaskCPF(masked)

	if unmasked != "12345678901" {
		t.Errorf("UnmaskCPF() = %q, expected '12345678901'", unmasked)
	}
}
