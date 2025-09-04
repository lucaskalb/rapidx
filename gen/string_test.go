package gen

import (
	"math/rand"
	"testing"
)

func TestString(t *testing.T) {
	gen := String("abc", Size{Min: 5, Max: 10})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a string
	if len(value) < 5 || len(value) > 10 {
		t.Errorf("String().Generate() = %q (len=%d), expected length 5-10", value, len(value))
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("String().Generate() returned nil shrinker")
	}
}

func TestStringAlpha(t *testing.T) {
	gen := StringAlpha(Size{Min: 3, Max: 8})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a string with alpha characters
	if len(value) < 3 || len(value) > 8 {
		t.Errorf("StringAlpha().Generate() = %q (len=%d), expected length 3-8", value, len(value))
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("StringAlpha().Generate() returned nil shrinker")
	}
}

func TestStringAlphaNum(t *testing.T) {
	gen := StringAlphaNum(Size{Min: 3, Max: 8})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a string with alphanumeric characters
	if len(value) < 3 || len(value) > 8 {
		t.Errorf("StringAlphaNum().Generate() = %q (len=%d), expected length 3-8", value, len(value))
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("StringAlphaNum().Generate() returned nil shrinker")
	}
}

func TestStringDigits(t *testing.T) {
	gen := StringDigits(Size{Min: 3, Max: 8})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a string with digit characters
	if len(value) < 3 || len(value) > 8 {
		t.Errorf("StringDigits().Generate() = %q (len=%d), expected length 3-8", value, len(value))
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("StringDigits().Generate() returned nil shrinker")
	}
}

func TestStringASCII(t *testing.T) {
	gen := StringASCII(Size{Min: 3, Max: 8})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a string with ASCII characters
	if len(value) < 3 || len(value) > 8 {
		t.Errorf("StringASCII().Generate() = %q (len=%d), expected length 3-8", value, len(value))
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("StringASCII().Generate() returned nil shrinker")
	}
}

func TestStringShrinker(t *testing.T) {
	// Test string shrinking behavior
	gen := String("abc", Size{Min: 5, Max: 10})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a string
	if len(value) < 5 || len(value) > 10 {
		t.Errorf("String().Generate() = %q (len=%d), expected length 5-10", value, len(value))
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("String().Generate() returned nil shrinker")
	}

	// Test shrinking behavior
	next, ok := shrink(false)
	if !ok {
		t.Error("String shrinker returned false on first call")
	}

	// Test that shrunk value is shorter or equal
	if len(next) > len(value) {
		t.Errorf("String shrinker returned longer string: %q (len=%d) vs %q (len=%d)", next, len(next), value, len(value))
	}
}
