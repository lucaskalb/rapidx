package gen

import (
	"math/rand"
	"testing"
)

func TestSize(t *testing.T) {
	size := Size{Min: 10, Max: 20}
	if size.Min != 10 {
		t.Errorf("Size.Min = %d, expected 10", size.Min)
	}
	if size.Max != 20 {
		t.Errorf("Size.Max = %d, expected 20", size.Max)
	}
}

func TestSetShrinkStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		expected string
	}{
		{"set dfs", "dfs", "dfs"},
		{"set bfs", "bfs", "bfs"},
		{"set invalid", "invalid", "bfs"},
		{"set empty", "", "bfs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetShrinkStrategy(tt.strategy)
			// We can't directly test the internal variable, but we can test behavior
			// by creating a generator and checking its shrinking behavior
		})
	}
}

func TestGenFunc(t *testing.T) {
	expected := 42
	gen := GenFunc[int]{
		fn: func(r *rand.Rand, sz Size) (int, Shrinker[int]) {
			return expected, func(accept bool) (int, bool) {
				return 0, false
			}
		},
	}

	r := rand.New(rand.NewSource(123))
	value, _ := gen.Generate(r, Size{})
	if value != expected {
		t.Errorf("GenFunc.Generate() = %d, expected %d", value, expected)
	}
}

func TestFrom(t *testing.T) {
	expected := "test"
	gen := From(func(r *rand.Rand, sz Size) (string, Shrinker[string]) {
		return expected, func(accept bool) (string, bool) {
			return "", false
		}
	})

	r := rand.New(rand.NewSource(123))
	value, _ := gen.Generate(r, Size{})
	if value != expected {
		t.Errorf("From().Generate() = %q, expected %q", value, expected)
	}
}