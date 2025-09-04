package gen

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestBool(t *testing.T) {
	gen := Bool()
	r := rand.New(rand.NewSource(123))
	
	value, shrink := gen.Generate(r, Size{})
	
	// Test that we get a boolean value
	if value != true && value != false {
		t.Errorf("Bool().Generate() = %v, expected boolean", value)
	}
	
	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Bool().Generate() returned nil shrinker")
	}
}

func TestConst(t *testing.T) {
	gen := Const(42)
	r := rand.New(rand.NewSource(123))
	
	value, shrink := gen.Generate(r, Size{})
	
	// Test that we get the constant value
	if value != 42 {
		t.Errorf("Const().Generate() = %d, expected 42", value)
	}
	
	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Const().Generate() returned nil shrinker")
	}
}

func TestOneOf(t *testing.T) {
	gen := OneOf(Const(1), Const(2), Const(3))
	r := rand.New(rand.NewSource(123))
	
	value, shrink := gen.Generate(r, Size{})
	
	// Test that we get one of the expected values
	if value != 1 && value != 2 && value != 3 {
		t.Errorf("OneOf().Generate() = %d, expected 1, 2, or 3", value)
	}
	
	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("OneOf().Generate() returned nil shrinker")
	}
}

func TestWeighted(t *testing.T) {
	gen := Weighted(func(x int) float64 { return float64(x) }, Const(1), Const(2), Const(3))
	r := rand.New(rand.NewSource(123))
	
	value, shrink := gen.Generate(r, Size{})
	
	// Test that we get one of the expected values
	if value != 1 && value != 2 && value != 3 {
		t.Errorf("Weighted().Generate() = %d, expected 1, 2, or 3", value)
	}
	
	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Weighted().Generate() returned nil shrinker")
	}
}

func TestMap(t *testing.T) {
	intGen := Int(Size{Min: 1, Max: 5})
	gen := Map(intGen, func(x int) string {
		return fmt.Sprintf("value_%d", x)
	})
	r := rand.New(rand.NewSource(123))
	
	value, shrink := gen.Generate(r, Size{})
	
	// Test that we get a mapped string
	if !strings.HasPrefix(value, "value_") {
		t.Errorf("Map().Generate() = %q, expected string starting with 'value_'", value)
	}
	
	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Map().Generate() returned nil shrinker")
	}
}

func TestFilter(t *testing.T) {
	intGen := Int(Size{Min: 1, Max: 10})
	gen := Filter(intGen, func(x int) bool {
		return x%2 == 0 // Only even numbers
	}, 100) // max attempts
	r := rand.New(rand.NewSource(123))
	
	value, shrink := gen.Generate(r, Size{})
	
	// Test that we get an even number
	if value%2 != 0 {
		t.Errorf("Filter().Generate() = %d, expected even number", value)
	}
	
	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Filter().Generate() returned nil shrinker")
	}
}

func TestBind(t *testing.T) {
	intGen := Int(Size{Min: 1, Max: 3})
	gen := Bind(intGen, func(x int) Generator[string] {
		return Const(fmt.Sprintf("bound_%d", x))
	})
	r := rand.New(rand.NewSource(123))
	
	value, shrink := gen.Generate(r, Size{})
	
	// Test that we get a bound string
	if !strings.HasPrefix(value, "bound_") {
		t.Errorf("Bind().Generate() = %q, expected string starting with 'bound_'", value)
	}
	
	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Bind().Generate() returned nil shrinker")
	}
}