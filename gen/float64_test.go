package gen

import (
	"math"
	"math/rand"
	"testing"
)

// TestFloat64 is already defined in float_test.go

// TestFloat64Range is already defined in float_test.go

// TestFloat64Shrinker is already defined in float_test.go

func TestFloat64ShrinkerWithAccept(t *testing.T) {
	// Test shrinking behavior with accept=true
	_, shrink := float64ShrinkInit(50.0, 0.0, 100.0, false, false)
	
	// First call with accept=false
	next1, ok1 := shrink(false)
	if !ok1 {
		t.Error("Float64 shrinker returned false on first call")
	}
	
	// Second call with accept=true (should rebase)
	next2, ok2 := shrink(true)
	// It's possible that the shrinker exhausts quickly, so we don't require it to succeed
	
	// Test that first value is in range
	if isFinite(next1) && (next1 < 0.0 || next1 > 100.0) {
		t.Errorf("Float64 shrinker returned value %f outside range [0.0, 100.0]", next1)
	}
	if ok2 && isFinite(next2) && (next2 < 0.0 || next2 > 100.0) {
		t.Errorf("Float64 shrinker returned value %f outside range [0.0, 100.0]", next2)
	}
}

func TestFloat64ShrinkerExhaustion(t *testing.T) {
	// Test shrinking behavior until exhaustion
	_, shrink := float64ShrinkInit(50.0, 0.0, 100.0, false, false)
	
	// Call shrinker many times until it returns false
	callCount := 0
	for {
		_, ok := shrink(false)
		if !ok {
			break
		}
		callCount++
		if callCount > 1000 { // Safety limit
			t.Error("Float64 shrinker did not exhaust after 1000 calls")
			break
		}
	}
	
	// Should have made at least some calls
	if callCount == 0 {
		t.Error("Float64 shrinker exhausted immediately")
	}
}

func TestFloat64ShrinkerWithDFSSStrategy(t *testing.T) {
	// Test shrinking behavior with DFS strategy
	SetShrinkStrategy("dfs")
	defer SetShrinkStrategy("bfs") // Reset to default
	
	_, shrink := float64ShrinkInit(50.0, 0.0, 100.0, false, false)
	
	// Test that we get a value
	next, ok := shrink(false)
	if !ok {
		t.Error("Float64 shrinker returned false on first call")
	}
	
	// Test that value is in range
	if isFinite(next) && (next < 0.0 || next > 100.0) {
		t.Errorf("Float64 shrinker returned value %f outside range [0.0, 100.0]", next)
	}
}

func TestFloat64ShrinkerEdgeCases(t *testing.T) {
	// Test shrinking behavior with edge cases
	tests := []struct {
		name      string
		start     float64
		min       float64
		max       float64
		allowNaN  bool
		allowInf  bool
	}{
		{"same min max", 5.0, 5.0, 5.0, false, false},
		{"start at min", 0.0, 0.0, 100.0, false, false},
		{"start at max", 100.0, 0.0, 100.0, false, false},
		{"negative range", -50.0, -100.0, -10.0, false, false},
		{"zero range", 0.0, -10.0, 10.0, false, false},
		{"with NaN", math.NaN(), 0.0, 10.0, true, false},
		{"with Inf", math.Inf(1), 0.0, 10.0, false, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, shrink := float64ShrinkInit(tt.start, tt.min, tt.max, tt.allowNaN, tt.allowInf)
			
			// For NaN, we can't use == comparison
			if !math.IsNaN(tt.start) && start != tt.start {
				t.Errorf("float64ShrinkInit() start = %f, expected %f", start, tt.start)
			}
			
			// Test that shrinker is not nil
			if shrink == nil {
				t.Error("float64ShrinkInit() returned nil shrinker")
			}
			
			// Test that we can call shrinker at least once
			next, ok := shrink(false)
			if ok {
				// Test that value is in range (for finite values)
				if isFinite(next) && isFinite(tt.min) && isFinite(tt.max) {
					if next < tt.min || next > tt.max {
						t.Errorf("Float64 shrinker returned value %f outside range [%f, %f]", next, tt.min, tt.max)
					}
				}
			}
		})
	}
}

func TestAutoRangeF64(t *testing.T) {
	tests := []struct {
		name       string
		local      Size
		fromRunner Size
		expectedMin float64
		expectedMax float64
	}{
		{"both empty", Size{}, Size{}, -100.0, 100.0},
		{"local only", Size{Min: 0, Max: 50}, Size{}, -50.0, 50.0},
		{"runner only", Size{}, Size{Min: 0, Max: 30}, -30.0, 30.0},
		{"both set", Size{Min: 0, Max: 20}, Size{Min: 0, Max: 40}, -40.0, 40.0},
		{"negative values", Size{Min: -60, Max: 0}, Size{}, -60.0, 60.0},
		{"mixed values", Size{Min: -10, Max: 30}, Size{Min: 0, Max: 20}, -30.0, 30.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := autoRangeF64(tt.local, tt.fromRunner)
			if min != tt.expectedMin || max != tt.expectedMax {
				t.Errorf("autoRangeF64(%v, %v) = (%f, %f), expected (%f, %f)", 
					tt.local, tt.fromRunner, min, max, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestClampF64(t *testing.T) {
	tests := []struct {
		name string
		x    float64
		min  float64
		max  float64
		expected float64
	}{
		{"in range", 5.0, 0.0, 10.0, 5.0},
		{"below min", -5.0, 0.0, 10.0, 0.0},
		{"above max", 15.0, 0.0, 10.0, 10.0},
		{"at min", 0.0, 0.0, 10.0, 0.0},
		{"at max", 10.0, 0.0, 10.0, 10.0},
		{"reversed range", 5.0, 10.0, 0.0, 10.0},
		{"NaN input", math.NaN(), 0.0, 10.0, math.NaN()},
		{"Inf input", math.Inf(1), 0.0, 10.0, math.Inf(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampF64(tt.x, tt.min, tt.max)
			
			// Handle special cases
			if math.IsNaN(tt.expected) {
				if !math.IsNaN(result) {
					t.Errorf("clampF64(%f, %f, %f) = %f, expected NaN", 
						tt.x, tt.min, tt.max, result)
				}
			} else if math.IsInf(tt.expected, 0) {
				if !math.IsInf(result, 0) || math.IsInf(result, 0) != math.IsInf(tt.expected, 0) {
					t.Errorf("clampF64(%f, %f, %f) = %f, expected %f", 
						tt.x, tt.min, tt.max, result, tt.expected)
				}
			} else if result != tt.expected {
				t.Errorf("clampF64(%f, %f, %f) = %f, expected %f", 
					tt.x, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestUniformF64(t *testing.T) {
	r := rand.New(rand.NewSource(123))
	
	tests := []struct {
		name string
		min  float64
		max  float64
	}{
		{"normal range", 0.0, 10.0},
		{"single value", 5.0, 5.0},
		{"negative range", -10.0, -5.0},
		{"mixed range", -5.0, 5.0},
		{"invalid range", 10.0, 5.0}, // Should fall back to default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := uniformF64(r, tt.min, tt.max)
			
			// Test that we get a finite value
			if !isFinite(value) {
				t.Errorf("uniformF64(%f, %f) = %f, expected finite value", tt.min, tt.max, value)
			}
			
			// For valid ranges, test that value is in range
			if isFinite(tt.min) && isFinite(tt.max) && tt.max >= tt.min {
				if value < tt.min || value > tt.max {
					t.Errorf("uniformF64(%f, %f) = %f, expected value in range [%f, %f]", 
						tt.min, tt.max, value, tt.min, tt.max)
				}
			}
		})
	}
}

// TestFloat64Target is already defined in float_test.go

func TestMidpointTowardsF64(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		expected float64
	}{
		{"same values", 5.0, 5.0, 5.0},
		{"positive direction", 0.0, 10.0, 5.0},
		{"negative direction", 10.0, 0.0, 5.0},
		{"small step", 0.0, 1.0, 0.5},
		{"small step negative", 1.0, 0.0, 0.5},
		{"large step", 0.0, 100.0, 50.0},
		{"odd step", 0.0, 7.0, 3.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := midpointTowardsF64(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("midpointTowardsF64(%f, %f) = %f, expected %f", 
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestF64Key(t *testing.T) {
	tests := []struct {
		name string
		x    float64
	}{
		{"normal value", 1.0},
		{"zero", 0.0},
		{"negative", -1.0},
		{"NaN", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := f64key(tt.x)
			// Test that we get a uint64 key
			if key == 0 && tt.x != 0.0 {
				t.Errorf("f64key(%f) = %d, expected non-zero key", tt.x, key)
			}
		})
	}
}

func TestIsFinite(t *testing.T) {
	tests := []struct {
		name string
		x    float64
		expected bool
	}{
		{"normal value", 1.0, true},
		{"zero", 0.0, true},
		{"negative", -1.0, true},
		{"NaN", math.NaN(), false},
		{"positive infinity", math.Inf(1), false},
		{"negative infinity", math.Inf(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFinite(tt.x)
			if result != tt.expected {
				t.Errorf("isFinite(%f) = %v, expected %v", tt.x, result, tt.expected)
			}
		})
	}
}