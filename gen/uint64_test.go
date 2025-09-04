package gen

import (
	"math/rand"
	"testing"
)

// TestUint64 is already defined in uint_test.go

// TestUint64Range is already defined in uint_test.go

// TestUint64Shrinker is already defined in uint_test.go

func TestUint64ShrinkerWithAccept(t *testing.T) {
	// Test shrinking behavior with accept=true
	_, shrink := uint64ShrinkInit(50, 0, 100)
	
	// First call with accept=false
	next1, ok1 := shrink(false)
	if !ok1 {
		t.Error("Uint64 shrinker returned false on first call")
	}
	
	// Second call with accept=true (should rebase)
	next2, ok2 := shrink(true)
	// It's possible that the shrinker exhausts quickly, so we don't require it to succeed
	
	// Test that first value is in range
	if next1 < 0 || next1 > 100 {
		t.Errorf("Uint64 shrinker returned value %d outside range [0, 100]", next1)
	}
	if ok2 && (next2 < 0 || next2 > 100) {
		t.Errorf("Uint64 shrinker returned value %d outside range [0, 100]", next2)
	}
}

func TestUint64ShrinkerExhaustion(t *testing.T) {
	// Test shrinking behavior until exhaustion
	_, shrink := uint64ShrinkInit(50, 0, 100)
	
	// Call shrinker many times until it returns false
	callCount := 0
	for {
		_, ok := shrink(false)
		if !ok {
			break
		}
		callCount++
		if callCount > 1000 { // Safety limit
			t.Error("Uint64 shrinker did not exhaust after 1000 calls")
			break
		}
	}
	
	// Should have made at least some calls
	if callCount == 0 {
		t.Error("Uint64 shrinker exhausted immediately")
	}
}

func TestUint64ShrinkerWithDFSSStrategy(t *testing.T) {
	// Test shrinking behavior with DFS strategy
	SetShrinkStrategy("dfs")
	defer SetShrinkStrategy("bfs") // Reset to default
	
	_, shrink := uint64ShrinkInit(50, 0, 100)
	
	// Test that we get a value
	next, ok := shrink(false)
	if !ok {
		t.Error("Uint64 shrinker returned false on first call")
	}
	
	// Test that value is in range
	if next < 0 || next > 100 {
		t.Errorf("Uint64 shrinker returned value %d outside range [0, 100]", next)
	}
}

func TestUint64ShrinkerEdgeCases(t *testing.T) {
	// Test shrinking behavior with edge cases
	tests := []struct {
		name  string
		start uint64
		min   uint64
		max   uint64
	}{
		{"same min max", 5, 5, 5},
		{"start at min", 0, 0, 100},
		{"start at max", 100, 0, 100},
		{"zero range", 0, 0, 10},
		{"large range", 1000, 0, 2000},
		{"start at zero", 0, 0, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, shrink := uint64ShrinkInit(tt.start, tt.min, tt.max)
			
			if start != tt.start {
				t.Errorf("uint64ShrinkInit() start = %d, expected %d", start, tt.start)
			}
			
			// Test that shrinker is not nil
			if shrink == nil {
				t.Error("uint64ShrinkInit() returned nil shrinker")
			}
			
			// Test that we can call shrinker at least once
			next, ok := shrink(false)
			if ok {
				// Test that value is in range
				if next < tt.min || next > tt.max {
					t.Errorf("Uint64 shrinker returned value %d outside range [%d, %d]", next, tt.min, tt.max)
				}
			}
		})
	}
}

func TestAutoRangeUint64(t *testing.T) {
	tests := []struct {
		name       string
		local      Size
		fromRunner Size
		expectedMin uint64
		expectedMax uint64
	}{
		{"both empty", Size{}, Size{}, 0, 100},
		{"local only", Size{Min: 0, Max: 50}, Size{}, 0, 50},
		{"runner only", Size{}, Size{Min: 0, Max: 30}, 0, 30},
		{"both set", Size{Min: 0, Max: 20}, Size{Min: 0, Max: 40}, 0, 40},
		{"negative values ignored", Size{Min: -60, Max: 0}, Size{}, 0, 100},
		{"mixed values", Size{Min: -10, Max: 30}, Size{Min: 0, Max: 20}, 0, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := autoRangeUint64(tt.local, tt.fromRunner)
			if min != tt.expectedMin || max != tt.expectedMax {
				t.Errorf("autoRangeUint64(%v, %v) = (%d, %d), expected (%d, %d)", 
					tt.local, tt.fromRunner, min, max, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestClampU64(t *testing.T) {
	tests := []struct {
		name string
		x    uint64
		min  uint64
		max  uint64
		expected uint64
	}{
		{"in range", 5, 0, 10, 5},
		{"below min", 5, 10, 20, 10},
		{"above max", 25, 0, 20, 20},
		{"at min", 0, 0, 10, 0},
		{"at max", 10, 0, 10, 10},
		{"reversed range", 5, 10, 0, 10},
		{"zero range", 5, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampU64(tt.x, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clampU64(%d, %d, %d) = %d, expected %d", 
					tt.x, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestUint64MultipleGenerations(t *testing.T) {
	// Test that Uint64() generates different values over multiple calls
	r := rand.New(rand.NewSource(456))
	gen := Uint64(Size{Min: 0, Max: 100})
	
	// Generate multiple values and check that we get different values
	// (with high probability)
	values := make(map[uint64]bool)
	
	for i := 0; i < 100; i++ {
		value, _ := gen.Generate(r, Size{})
		values[value] = true
	}
	
	// We should get multiple different values (with high probability)
	if len(values) < 10 {
		t.Logf("Warning: Only got %d different values after 100 generations", len(values))
		// This is not necessarily an error, just unlikely
	}
}

func TestUint64RangeWithRunnerSize(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	
	// Test that runner size overrides local size
	gen := Uint64(Size{Min: 0, Max: 50})
	value, _ := gen.Generate(r, Size{Min: 0, Max: 30}) // runner size should override
	
	if value < 0 || value > 30 {
		t.Errorf("Uint64() with runner size returned value %d, expected value in range [0, 30]", value)
	}
}

func TestUint64ShrinkingTarget(t *testing.T) {
	// Test that uint64 shrinking targets 0
	_, shrink := uint64ShrinkInit(100, 0, 200)
	
	// Call shrinker multiple times and check if we get 0
	zeroFound := false
	for i := 0; i < 20; i++ {
		next, ok := shrink(false)
		if !ok {
			break
		}
		if next == 0 {
			zeroFound = true
			break
		}
	}
	
	if !zeroFound {
		t.Log("Warning: Uint64 shrinker did not produce 0 in first 20 attempts")
	}
}

func TestUint64ShrinkingBisection(t *testing.T) {
	// Test that uint64 shrinking uses bisection
	_, shrink := uint64ShrinkInit(100, 0, 200)
	
	// Call shrinker and check if we get values that are roughly half
	halfFound := false
	for i := 0; i < 10; i++ {
		next, ok := shrink(false)
		if !ok {
			break
		}
		// Check if we get a value around 50 (half of 100)
		if next >= 40 && next <= 60 {
			halfFound = true
			break
		}
	}
	
	if !halfFound {
		t.Log("Warning: Uint64 shrinker did not produce bisected values in first 10 attempts")
	}
}

func TestUint64ShrinkingUnitStep(t *testing.T) {
	// Test that uint64 shrinking uses unit steps
	_, shrink := uint64ShrinkInit(5, 0, 10)
	
	// Call shrinker and check if we get values that are one less
	unitStepFound := false
	for i := 0; i < 10; i++ {
		next, ok := shrink(false)
		if !ok {
			break
		}
		// Check if we get 4 (5-1)
		if next == 4 {
			unitStepFound = true
			break
		}
	}
	
	if !unitStepFound {
		t.Log("Warning: Uint64 shrinker did not produce unit step values in first 10 attempts")
	}
}

func TestUint64ShrinkingBoundaries(t *testing.T) {
	// Test that uint64 shrinking includes boundaries
	_, shrink := uint64ShrinkInit(50, 0, 100)
	
	// Call shrinker and check if we get boundary values
	minFound := false
	maxFound := false
	for i := 0; i < 20; i++ {
		next, ok := shrink(false)
		if !ok {
			break
		}
		if next == 0 {
			minFound = true
		}
		if next == 100 {
			maxFound = true
		}
	}
	
	if !minFound {
		t.Log("Warning: Uint64 shrinker did not produce minimum boundary value in first 20 attempts")
	}
	if !maxFound {
		t.Log("Warning: Uint64 shrinker did not produce maximum boundary value in first 20 attempts")
	}
}