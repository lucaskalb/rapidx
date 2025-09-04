package gen

import (
	"math/rand"
	"testing"
)

func TestArrayOf(t *testing.T) {
	intGen := Int(Size{Min: 0, Max: 10})
	gen := ArrayOf(intGen, 3)
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	if len(value) != 3 {
		t.Errorf("ArrayOf().Generate() = %v (len=%d), expected length 3", value, len(value))
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("ArrayOf().Generate() returned nil shrinker")
	}
}

func TestSliceOf(t *testing.T) {
	intGen := Int(Size{Min: 0, Max: 10})
	gen := SliceOf(intGen, Size{Min: 2, Max: 5})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a slice
	if len(value) < 2 || len(value) > 5 {
		t.Errorf("SliceOf().Generate() = %v (len=%d), expected length 2-5", value, len(value))
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("SliceOf().Generate() returned nil shrinker")
	}
}

func TestSliceShrinker(t *testing.T) {

	intGen := Int(Size{Min: 0, Max: 10})
	gen := SliceOf(intGen, Size{Min: 2, Max: 5})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a slice
	if len(value) < 2 || len(value) > 5 {
		t.Errorf("SliceOf().Generate() = %v (len=%d), expected length 2-5", value, len(value))
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("SliceOf().Generate() returned nil shrinker")
	}

	// Test shrinking behavior
	next, ok := shrink(false)
	if !ok {
		t.Error("Slice shrinker returned false on first call")
	}

	// Test that shrunk value is shorter or equal
	if len(next) > len(value) {
		t.Errorf("Slice shrinker returned longer slice: %v (len=%d) vs %v (len=%d)", next, len(next), value, len(value))
	}
}
