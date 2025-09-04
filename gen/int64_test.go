package gen

import (
	"math/rand"
	"testing"
)

func TestInt64(t *testing.T) {
	gen := Int64(Size{Min: 0, Max: 100})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get an int64 (the range might be different due to autoRange logic)
	if value < -100 || value > 100 {
		t.Errorf("Int64().Generate() = %d, expected value in range [-100, 100]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Int64().Generate() returned nil shrinker")
	}
}

func TestInt64Range(t *testing.T) {
	gen := Int64Range(10, 20)
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get an int64 in range
	if value < 10 || value > 20 {
		t.Errorf("Int64Range().Generate() = %d, expected value in range [10, 20]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Int64Range().Generate() returned nil shrinker")
	}
}

func TestInt64Shrinker(t *testing.T) {
	// Test int64 shrinking behavior
	start, shrink := int64ShrinkInit(50, 0, 100)

	if start != 50 {
		t.Errorf("int64ShrinkInit() start = %d, expected 50", start)
	}

	if shrink == nil {
		t.Error("int64ShrinkInit() returned nil shrinker")
	}

	// Test shrinking behavior
	next, ok := shrink(false)
	if !ok {
		t.Error("Int64 shrinker returned false on first call")
	}

	// Test that value is in range
	if next < 0 || next > 100 {
		t.Errorf("Int64 shrinker returned value %d outside range [0, 100]", next)
	}
}
