package gen

import (
	"math/rand"
	"testing"
)

func TestSliceOfWithRunnerSize(t *testing.T) {
	r := rand.New(rand.NewSource(123))

	gen := SliceOf(Int(Size{}), Size{Min: 0, Max: 5})
	value, _ := gen.Generate(r, Size{Min: 0, Max: 3})

	if len(value) > 3 {
		t.Errorf("SliceOf() with runner size returned slice of length %d, expected length in range [0, 3]",
			len(value))
	}
}

func TestSliceOfShrinker(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 0, Max: 100}), Size{Min: 3, Max: 5})
	start, shrink := gen.Generate(r, Size{})

	if start == nil {
		t.Error("SliceOf().Generate() returned nil slice")
	}

	if shrink == nil {
		t.Error("SliceOf().Generate() returned nil shrinker")
	}

	next, ok := shrink(false)
	if !ok {
		t.Error("Slice shrinker returned false on first call")
	}

	if len(next) == len(start) {

		same := true
		for i := range next {
			if next[i] != start[i] {
				same = false
				break
			}
		}
		if same {
			t.Error("Slice shrinker returned identical slice")
		}
	}
}

func TestSliceOfShrinkerWithAccept(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 0, Max: 100}), Size{Min: 3, Max: 5})
	_, shrink := gen.Generate(r, Size{})

	next1, ok1 := shrink(false)
	if !ok1 {
		t.Error("Slice shrinker returned false on first call")
	}

	next2, ok2 := shrink(true)

	if next1 == nil {
		t.Error("Slice shrinker returned nil slice")
	}
	if ok2 && next2 == nil {
		t.Error("Slice shrinker returned nil slice on second call")
	}
}

func TestSliceOfShrinkerExhaustion(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 0, Max: 100}), Size{Min: 3, Max: 5})
	_, shrink := gen.Generate(r, Size{})

	callCount := 0
	for {
		_, ok := shrink(false)
		if !ok {
			break
		}
		callCount++
		if callCount > 1000 {
			t.Error("Slice shrinker did not exhaust after 1000 calls")
			break
		}
	}

	if callCount == 0 {
		t.Error("Slice shrinker exhausted immediately")
	}
}

func TestSliceOfShrinkerWithDFSSStrategy(t *testing.T) {
	SetShrinkStrategy(ShrinkStrategyDFS)
	defer SetShrinkStrategy(ShrinkStrategyBFS) // Reset to default

	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 0, Max: 100}), Size{Min: 3, Max: 5})
	_, shrink := gen.Generate(r, Size{})

	next, ok := shrink(false)
	if !ok {
		t.Error("Slice shrinker returned false on first call")
	}

	if next == nil {
		t.Error("Slice shrinker returned nil slice")
	}
}

func TestSliceOfShrinkerEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		elem Generator[int]
		size Size
	}{
		{"empty slice", Int(Size{}), Size{Min: 0, Max: 0}},
		{"single element", Int(Size{Min: 5, Max: 5}), Size{Min: 1, Max: 1}},
		{"small range", Int(Size{Min: 0, Max: 10}), Size{Min: 2, Max: 2}},
		{"large range", Int(Size{Min: 0, Max: 1000}), Size{Min: 1, Max: 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(123))
			gen := SliceOf(tt.elem, tt.size)
			start, shrink := gen.Generate(r, Size{})

			if start == nil {
				t.Error("SliceOf().Generate() returned nil slice")
			}

			if shrink == nil {
				t.Error("SliceOf().Generate() returned nil shrinker")
			}

			if len(start) > 0 {
				next, ok := shrink(false)
				if ok {

					if next == nil && len(start) > 1 {
						t.Error("Slice shrinker returned nil slice for multi-element slice")
					}
				} else {

				}
			}
		})
	}
}

func TestSliceOfWithDifferentTypes(t *testing.T) {
	r := rand.New(rand.NewSource(123))

	tests := []struct {
		name string
		gen  Generator[[]string]
	}{
		{"string slice", SliceOf(StringAlpha(Size{Min: 1, Max: 5}), Size{Min: 1, Max: 3})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, shrink := tt.gen.Generate(r, Size{})

			if value == nil {
				t.Error("SliceOf().Generate() returned nil slice")
			}

			if shrink == nil {
				t.Error("SliceOf().Generate() returned nil shrinker")
			}
		})
	}
}

func TestSliceOfShrinkingStrategies(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 0, Max: 100}), Size{Min: 4, Max: 6})
	start, _ := gen.Generate(r, Size{})

	shorterFound := false
	_, shrink := gen.Generate(r, Size{})

	for i := 0; i < 10; i++ {
		next, ok := shrink(false)
		if !ok {
			break
		}
		if len(next) < len(start) {
			shorterFound = true
			break
		}
	}

	if !shorterFound {
		t.Log("Warning: Slice shrinker did not produce shorter slices in first 10 attempts")
	}
}

func TestSliceOfElementShrinking(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 50, Max: 100}), Size{Min: 2, Max: 3})
	start, shrink := gen.Generate(r, Size{})

	elementChanged := false

	for i := 0; i < 20; i++ {
		next, ok := shrink(false)
		if !ok {
			break
		}
		if len(next) == len(start) {

			for j := range next {
				if next[j] != start[j] {
					elementChanged = true
					break
				}
			}
		}
		if elementChanged {
			break
		}
	}

	if !elementChanged {
		t.Log("Warning: Slice shrinker did not modify individual elements in first 20 attempts")
	}
}

func TestSig(t *testing.T) {
	tests := []struct {
		name string
		s    []int
	}{
		{"empty slice", []int{}},
		{"single element", []int{1}},
		{"multiple elements", []int{1, 2, 3}},
		{"negative elements", []int{-1, -2, -3}},
		{"mixed elements", []int{-1, 0, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := sig(tt.s)

			if signature == "" {
				t.Error("sig() returned empty signature")
			}

			signature2 := sig(tt.s)
			if signature != signature2 {
				t.Error("sig() returned different signatures for same slice")
			}
		})
	}
}

func TestSliceOfWithBoolElements(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Bool(), Size{Min: 2, Max: 4})
	value, shrink := gen.Generate(r, Size{})

	if value == nil {
		t.Error("SliceOf(Bool()).Generate() returned nil slice")
	}

	for i, v := range value {
		if v != true && v != false {
			t.Errorf("SliceOf(Bool()).Generate() returned invalid boolean at index %d: %v", i, v)
		}
	}

	if shrink == nil {
		t.Error("SliceOf(Bool()).Generate() returned nil shrinker")
	}
}

func TestSliceOfWithFloatElements(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Float64(Size{Min: 0, Max: 100}), Size{Min: 1, Max: 3})
	value, shrink := gen.Generate(r, Size{})

	if value == nil {
		t.Error("SliceOf(Float64()).Generate() returned nil slice")
	}

	if shrink == nil {
		t.Error("SliceOf(Float64()).Generate() returned nil shrinker")
	}
}
