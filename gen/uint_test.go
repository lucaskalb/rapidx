package gen

import (
	"math/rand"
	"testing"
)

func TestUint(t *testing.T) {
	gen := Uint(Size{Min: 0, Max: 100})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a uint
	if value > 100 {
		t.Errorf("Uint().Generate() = %d, expected value in range [0, 100]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Uint().Generate() returned nil shrinker")
	}
}

func TestUintRange(t *testing.T) {
	gen := UintRange(10, 20)
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a uint in range
	if value < 10 || value > 20 {
		t.Errorf("UintRange().Generate() = %d, expected value in range [10, 20]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("UintRange().Generate() returned nil shrinker")
	}
}

func TestUintShrinker(t *testing.T) {
	// Test uint shrinking behavior
	start, shrink := uintShrinkInit(50, 0, 100)

	if start != 50 {
		t.Errorf("uintShrinkInit() start = %d, expected 50", start)
	}

	if shrink == nil {
		t.Error("uintShrinkInit() returned nil shrinker")
	}

	// Test shrinking behavior
	next, ok := shrink(false)
	if !ok {
		t.Error("Uint shrinker returned false on first call")
	}

	// Test that value is in range
	if next > 100 {
		t.Errorf("Uint shrinker returned value %d outside range [0, 100]", next)
	}
}

func TestUint64(t *testing.T) {
	gen := Uint64(Size{Min: 0, Max: 100})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a uint64
	if value > 100 {
		t.Errorf("Uint64().Generate() = %d, expected value in range [0, 100]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Uint64().Generate() returned nil shrinker")
	}
}

func TestUint64Range(t *testing.T) {
	gen := Uint64Range(10, 20)
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a uint64 in range
	if value < 10 || value > 20 {
		t.Errorf("Uint64Range().Generate() = %d, expected value in range [10, 20]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Uint64Range().Generate() returned nil shrinker")
	}
}

func TestUint64Shrinker(t *testing.T) {
	// Test uint64 shrinking behavior
	start, shrink := uint64ShrinkInit(50, 0, 100)

	if start != 50 {
		t.Errorf("uint64ShrinkInit() start = %d, expected 50", start)
	}

	if shrink == nil {
		t.Error("uint64ShrinkInit() returned nil shrinker")
	}

	// Test shrinking behavior
	next, ok := shrink(false)
	if !ok {
		t.Error("Uint64 shrinker returned false on first call")
	}

	// Test that value is in range
	if next > 100 {
		t.Errorf("Uint64 shrinker returned value %d outside range [0, 100]", next)
	}
}
