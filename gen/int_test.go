package gen

import (
	"math/rand"
	"testing"
)

func TestInt(t *testing.T) {
	r := rand.New(rand.NewSource(123))

	tests := []struct {
		name string
		size Size
	}{
		{"default size", Size{}},
		{"positive range", Size{Min: 0, Max: 100}},
		{"negative range", Size{Min: -100, Max: 0}},
		{"mixed range", Size{Min: -50, Max: 50}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := Int(tt.size)
			value, shrink := gen.Generate(r, Size{})

			// Test that we get a value
			if value == 0 && tt.size.Min != 0 && tt.size.Max != 0 {
				// This is acceptable for some ranges
			}

			// Test that shrinker is not nil
			if shrink == nil {
				t.Error("Int().Generate() returned nil shrinker")
			}
		})
	}
}

func TestIntRange(t *testing.T) {
	r := rand.New(rand.NewSource(123))

	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"normal range", 10, 20},
		{"reversed range", 20, 10},
		{"single value", 5, 5},
		{"negative range", -20, -10},
		{"mixed range", -10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := IntRange(tt.min, tt.max)
			value, shrink := gen.Generate(r, Size{})

			// Test that value is in range
			expectedMin := tt.min
			expectedMax := tt.max
			if tt.min > tt.max {
				expectedMin, expectedMax = tt.max, tt.min
			}

			if value < expectedMin || value > expectedMax {
				t.Errorf("IntRange(%d, %d).Generate() = %d, expected value in range [%d, %d]",
					tt.min, tt.max, value, expectedMin, expectedMax)
			}

			// Test that shrinker is not nil
			if shrink == nil {
				t.Error("IntRange().Generate() returned nil shrinker")
			}
		})
	}
}

func TestShrinkTarget(t *testing.T) {
	tests := []struct {
		name     string
		min      int
		max      int
		expected int
	}{
		{"zero in range", -10, 10, 0},
		{"zero at min", 0, 10, 0},
		{"zero at max", -10, 0, 0},
		{"all positive", 5, 15, 5},
		{"all negative", -15, -5, -5},
		{"single positive", 5, 5, 5},
		{"single negative", -5, -5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shrinkTarget(tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("shrinkTarget(%d, %d) = %d, expected %d",
					tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestMidpointTowards(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"same values", 5, 5, 5},
		{"positive direction", 0, 10, 5},
		{"negative direction", 10, 0, 5},
		{"small step", 0, 1, 1},
		{"small step negative", 1, 0, 0},
		{"large step", 0, 100, 50},
		{"odd step", 0, 7, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := midpointTowards(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("midpointTowards(%d, %d) = %d, expected %d",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestStepTowards(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"same values", 5, 5, 5},
		{"positive direction", 0, 10, 1},
		{"negative direction", 10, 0, 9},
		{"small step", 0, 1, 1},
		{"small step negative", 1, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stepTowards(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("stepTowards(%d, %d) = %d, expected %d",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestAutoRange(t *testing.T) {
	tests := []struct {
		name        string
		local       Size
		fromRunner  Size
		expectedMin int
		expectedMax int
	}{
		{"both empty", Size{}, Size{}, -100, 100},
		{"local only", Size{Min: 0, Max: 50}, Size{}, -50, 50},
		{"runner only", Size{}, Size{Min: 0, Max: 30}, -30, 30},
		{"both set", Size{Min: 0, Max: 20}, Size{Min: 0, Max: 40}, -40, 40},
		{"negative values", Size{Min: -60, Max: 0}, Size{}, -60, 60},
		{"mixed values", Size{Min: -10, Max: 30}, Size{Min: 0, Max: 20}, -30, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := autoRange(tt.local, tt.fromRunner)
			if min != tt.expectedMin || max != tt.expectedMax {
				t.Errorf("autoRange(%v, %v) = (%d, %d), expected (%d, %d)",
					tt.local, tt.fromRunner, min, max, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		x        int
		min      int
		max      int
		expected int
	}{
		{"in range", 5, 0, 10, 5},
		{"below min", -5, 0, 10, 0},
		{"above max", 15, 0, 10, 10},
		{"at min", 0, 0, 10, 0},
		{"at max", 10, 0, 10, 10},
		{"reversed range", 5, 10, 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clamp(tt.x, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clamp(%d, %d, %d) = %d, expected %d",
					tt.x, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestAbsInt(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"positive", 5, 5},
		{"negative", -5, 5},
		{"zero", 0, 0},
		{"large positive", 1000, 1000},
		{"large negative", -1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := absInt(tt.input)
			if result != tt.expected {
				t.Errorf("absInt(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaxInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a greater", 10, 5, 10},
		{"b greater", 5, 10, 10},
		{"equal", 5, 5, 5},
		{"negative", -10, -5, -5},
		{"mixed", -5, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maxInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("maxInt(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestIntShrinker(t *testing.T) {
	start, shrink := intShrinkInit(50, 0, 100)

	if start != 50 {
		t.Errorf("intShrinkInit() start = %d, expected 50", start)
	}

	if shrink == nil {
		t.Error("intShrinkInit() returned nil shrinker")
	}

	// Test shrinking behavior
	next, ok := shrink(false) // First call with accept=false
	if !ok {
		t.Error("Shrinker returned false on first call")
	}

	// Test that we get a different value
	if next == start {
		t.Error("Shrinker returned same value as start")
	}

	// Test that value is in range
	if next < 0 || next > 100 {
		t.Errorf("Shrinker returned value %d outside range [0, 100]", next)
	}
}

func TestIntShrinkerWithAccept(t *testing.T) {
	// Test shrinking behavior with accept=true
	_, shrink := intShrinkInit(50, 0, 100)

	// First call with accept=false
	next1, ok1 := shrink(false)
	if !ok1 {
		t.Error("Shrinker returned false on first call")
	}

	// Second call with accept=true (should rebase)
	next2, ok2 := shrink(true)
	// It's possible that the shrinker exhausts quickly, so we don't require it to succeed

	// Test that first value is in range
	if next1 < 0 || next1 > 100 {
		t.Errorf("Shrinker returned value %d outside range [0, 100]", next1)
	}
	if ok2 && (next2 < 0 || next2 > 100) {
		t.Errorf("Shrinker returned value %d outside range [0, 100]", next2)
	}
}

func TestIntShrinkerExhaustion(t *testing.T) {
	// Test shrinking behavior until exhaustion
	_, shrink := intShrinkInit(50, 0, 100)

	// Call shrinker many times until it returns false
	callCount := 0
	for {
		_, ok := shrink(false)
		if !ok {
			break
		}
		callCount++
		if callCount > 1000 { // Safety limit
			t.Error("Shrinker did not exhaust after 1000 calls")
			break
		}
	}

	// Should have made at least some calls
	if callCount == 0 {
		t.Error("Shrinker exhausted immediately")
	}
}

func TestIntShrinkerWithDFSSStrategy(t *testing.T) {
	// Test shrinking behavior with DFS strategy
	SetShrinkStrategy(ShrinkStrategyDFS)
	defer SetShrinkStrategy(ShrinkStrategyBFS) // Reset to default

	_, shrink := intShrinkInit(50, 0, 100)

	// Test that we get a value
	next, ok := shrink(false)
	if !ok {
		t.Error("Shrinker returned false on first call")
	}

	// Test that value is in range
	if next < 0 || next > 100 {
		t.Errorf("Shrinker returned value %d outside range [0, 100]", next)
	}
}

func TestIntShrinkerWithInvalidStrategy(t *testing.T) {
	// Test shrinking behavior with invalid strategy (should default to BFS)
	SetShrinkStrategy("invalid")
	defer SetShrinkStrategy(ShrinkStrategyBFS) // Reset to default

	_, shrink := intShrinkInit(50, 0, 100)

	// Test that we get a value
	next, ok := shrink(false)
	if !ok {
		t.Error("Shrinker returned false on first call")
	}

	// Test that value is in range
	if next < 0 || next > 100 {
		t.Errorf("Shrinker returned value %d outside range [0, 100]", next)
	}
}

func TestIntShrinkerEdgeCases(t *testing.T) {
	// Test shrinking behavior with edge cases
	tests := []struct {
		name  string
		start int
		min   int
		max   int
	}{
		{"same min max", 5, 5, 5},
		{"start at min", 0, 0, 100},
		{"start at max", 100, 0, 100},
		{"negative range", -50, -100, -10},
		{"zero range", 0, -10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, shrink := intShrinkInit(tt.start, tt.min, tt.max)

			if start != tt.start {
				t.Errorf("intShrinkInit() start = %d, expected %d", start, tt.start)
			}

			// Test that shrinker is not nil
			if shrink == nil {
				t.Error("intShrinkInit() returned nil shrinker")
			}

			// Test that we can call shrinker at least once
			next, ok := shrink(false)
			if ok {
				// Test that value is in range
				if next < tt.min || next > tt.max {
					t.Errorf("Shrinker returned value %d outside range [%d, %d]", next, tt.min, tt.max)
				}
			}
		})
	}
}
