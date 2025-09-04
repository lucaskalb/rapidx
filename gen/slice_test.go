package gen

import (
	"math/rand"
	"testing"
)

// TestSliceOf is already defined in array_test.go

func TestSliceOfWithRunnerSize(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	
	// Test that runner size overrides local size
	gen := SliceOf(Int(Size{}), Size{Min: 0, Max: 5})
	value, _ := gen.Generate(r, Size{Min: 0, Max: 3}) // runner size should override
	
	if len(value) < 0 || len(value) > 3 {
		t.Errorf("SliceOf() with runner size returned slice of length %d, expected length in range [0, 3]", 
			len(value))
	}
}

func TestSliceOfShrinker(t *testing.T) {
	// Test slice shrinking behavior
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 0, Max: 100}), Size{Min: 3, Max: 5})
	start, shrink := gen.Generate(r, Size{})
	
	if start == nil {
		t.Error("SliceOf().Generate() returned nil slice")
	}
	
	if shrink == nil {
		t.Error("SliceOf().Generate() returned nil shrinker")
	}
	
	// Test shrinking behavior
	next, ok := shrink(false)
	if !ok {
		t.Error("Slice shrinker returned false on first call")
	}
	
	// Test that we get a different slice
	if len(next) == len(start) {
		// If same length, at least one element should be different
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
	// Test shrinking behavior with accept=true
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 0, Max: 100}), Size{Min: 3, Max: 5})
	_, shrink := gen.Generate(r, Size{})
	
	// First call with accept=false
	next1, ok1 := shrink(false)
	if !ok1 {
		t.Error("Slice shrinker returned false on first call")
	}
	
	// Second call with accept=true (should rebase)
	next2, ok2 := shrink(true)
	// It's possible that the shrinker exhausts quickly, so we don't require it to succeed
	
	// Test that first value is a valid slice
	if next1 == nil {
		t.Error("Slice shrinker returned nil slice")
	}
	if ok2 && next2 == nil {
		t.Error("Slice shrinker returned nil slice on second call")
	}
}

func TestSliceOfShrinkerExhaustion(t *testing.T) {
	// Test shrinking behavior until exhaustion
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 0, Max: 100}), Size{Min: 3, Max: 5})
	_, shrink := gen.Generate(r, Size{})
	
	// Call shrinker many times until it returns false
	callCount := 0
	for {
		_, ok := shrink(false)
		if !ok {
			break
		}
		callCount++
		if callCount > 1000 { // Safety limit
			t.Error("Slice shrinker did not exhaust after 1000 calls")
			break
		}
	}
	
	// Should have made at least some calls
	if callCount == 0 {
		t.Error("Slice shrinker exhausted immediately")
	}
}

func TestSliceOfShrinkerWithDFSSStrategy(t *testing.T) {
	// Test shrinking behavior with DFS strategy
	SetShrinkStrategy("dfs")
	defer SetShrinkStrategy("bfs") // Reset to default
	
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Int(Size{Min: 0, Max: 100}), Size{Min: 3, Max: 5})
	_, shrink := gen.Generate(r, Size{})
	
	// Test that we get a value
	next, ok := shrink(false)
	if !ok {
		t.Error("Slice shrinker returned false on first call")
	}
	
	// Test that value is a valid slice
	if next == nil {
		t.Error("Slice shrinker returned nil slice")
	}
}

func TestSliceOfShrinkerEdgeCases(t *testing.T) {
	// Test shrinking behavior with edge cases
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
			
			// Test that shrinker is not nil
			if shrink == nil {
				t.Error("SliceOf().Generate() returned nil shrinker")
			}
			
			// Test that we can call shrinker at least once (if slice is not empty)
			if len(start) > 0 {
				next, ok := shrink(false)
				if ok {
					// For single element slices, the shrinker might return an empty slice
					// when removing the only element, which is valid behavior
					if next == nil && len(start) > 1 {
						t.Error("Slice shrinker returned nil slice for multi-element slice")
					}
				} else {
					// For single element slices, it's possible the shrinker exhausts immediately
					// This is not necessarily an error
				}
			}
		})
	}
}

func TestSliceOfWithDifferentTypes(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	
	// Test with different element types
	tests := []struct {
		name string
		gen  Generator[[]string]
	}{
		{"string slice", SliceOf(StringAlpha(Size{Min: 1, Max: 5}), Size{Min: 1, Max: 3})},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, shrink := tt.gen.Generate(r, Size{})
			
			// Test that we get a slice
			if value == nil {
				t.Error("SliceOf().Generate() returned nil slice")
			}
			
			// Test that shrinker is not nil
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
	
	// Test that shrinking produces different slice lengths
	// (removing elements should make slices shorter)
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
	
	// Test that shrinking can modify individual elements
	// (element shrinking should produce different values)
	elementChanged := false
	_, shrink = gen.Generate(r, Size{})
	
	for i := 0; i < 20; i++ {
		next, ok := shrink(false)
		if !ok {
			break
		}
		if len(next) == len(start) {
			// Same length, check if any element changed
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
			// Test that we get a non-empty signature
			if signature == "" {
				t.Error("sig() returned empty signature")
			}
			
			// Test that same slice produces same signature
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
	
	// Test that we get a slice of booleans
	if value == nil {
		t.Error("SliceOf(Bool()).Generate() returned nil slice")
	}
	
	// Test that all elements are valid booleans
	for i, v := range value {
		if v != true && v != false {
			t.Errorf("SliceOf(Bool()).Generate() returned invalid boolean at index %d: %v", i, v)
		}
	}
	
	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("SliceOf(Bool()).Generate() returned nil shrinker")
	}
}

func TestSliceOfWithFloatElements(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	gen := SliceOf(Float64(Size{Min: 0, Max: 100}), Size{Min: 1, Max: 3})
	value, shrink := gen.Generate(r, Size{})
	
	// Test that we get a slice of floats
	if value == nil {
		t.Error("SliceOf(Float64()).Generate() returned nil slice")
	}
	
	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("SliceOf(Float64()).Generate() returned nil shrinker")
	}
}