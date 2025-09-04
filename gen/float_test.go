package gen

import (
	"math"
	"math/rand"
	"testing"
)

func TestFloat32(t *testing.T) {
	gen := Float32(Size{Min: 0, Max: 100})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a float32
	if value < 0 || value > 100 {
		t.Errorf("Float32().Generate() = %f, expected value in range [0, 100]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Float32().Generate() returned nil shrinker")
	}
}

func TestFloat32Range(t *testing.T) {
	gen := Float32Range(10.0, 20.0, false, false)
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a float32 in range
	if value < 10.0 || value > 20.0 {
		t.Errorf("Float32Range().Generate() = %f, expected value in range [10.0, 20.0]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Float32Range().Generate() returned nil shrinker")
	}
}

func TestFloat32Shrinker(t *testing.T) {
	// Test float32 shrinking behavior
	start, shrink := float32ShrinkInit(50.0, 0.0, 100.0, false, false)

	if start != 50.0 {
		t.Errorf("float32ShrinkInit() start = %f, expected 50.0", start)
	}

	if shrink == nil {
		t.Error("float32ShrinkInit() returned nil shrinker")
	}

	// Test shrinking behavior
	next, ok := shrink(false)
	if !ok {
		t.Error("Float32 shrinker returned false on first call")
	}

	// Test that value is in range
	if next < 0.0 || next > 100.0 {
		t.Errorf("Float32 shrinker returned value %f outside range [0.0, 100.0]", next)
	}
}

func TestFloat32HelperFunctions(t *testing.T) {
	// Test float32 helper functions
	tests := []struct {
		name string
		f    func() bool
	}{
		{"float32IsFinite", func() bool { return float32IsFinite(1.0) }},
		{"float32IsNaN", func() bool { return float32IsNaN(float32(math.NaN())) }},
		{"float32IsInf", func() bool { return float32IsInf(float32(math.Inf(1))) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test that the function doesn't panic
			tt.f()
		})
	}
}

func TestFloat32Clamp(t *testing.T) {
	// Test float32 clamp function
	tests := []struct {
		name     string
		x        float32
		min      float32
		max      float32
		expected float32
	}{
		{"in range", 5.0, 0.0, 10.0, 5.0},
		{"below min", -5.0, 0.0, 10.0, 0.0},
		{"above max", 15.0, 0.0, 10.0, 10.0},
		{"at min", 0.0, 0.0, 10.0, 0.0},
		{"at max", 10.0, 0.0, 10.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampF32(tt.x, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clampF32(%f, %f, %f) = %f, expected %f",
					tt.x, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestFloat32AutoRange(t *testing.T) {
	// Test float32 auto range function
	tests := []struct {
		name        string
		local       Size
		fromRunner  Size
		expectedMin float32
		expectedMax float32
	}{
		{"both empty", Size{}, Size{}, -100.0, 100.0},
		{"local only", Size{Min: 0, Max: 50}, Size{}, -50.0, 50.0},
		{"runner only", Size{}, Size{Min: 0, Max: 30}, -30.0, 30.0},
		{"both set", Size{Min: 0, Max: 20}, Size{Min: 0, Max: 40}, -40.0, 40.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := autoRangeF32(tt.local, tt.fromRunner)
			if min != tt.expectedMin || max != tt.expectedMax {
				t.Errorf("autoRangeF32(%v, %v) = (%f, %f), expected (%f, %f)",
					tt.local, tt.fromRunner, min, max, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestFloat32Target(t *testing.T) {
	// Test float32 target function
	tests := []struct {
		name     string
		min      float32
		max      float32
		expected float32
	}{
		{"zero in range", -10.0, 10.0, 0.0},
		{"zero at min", 0.0, 10.0, 0.0},
		{"zero at max", -10.0, 0.0, 0.0},
		{"all positive", 5.0, 15.0, 5.0},
		{"all negative", -15.0, -5.0, -5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := float32Target(tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("float32Target(%f, %f) = %f, expected %f",
					tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestFloat32MidpointTowards(t *testing.T) {
	// Test float32 midpoint towards function
	tests := []struct {
		name     string
		a        float32
		b        float32
		expected float32
	}{
		{"same values", 5.0, 5.0, 5.0},
		{"positive direction", 0.0, 10.0, 5.0},
		{"negative direction", 10.0, 0.0, 5.0},
		{"small step", 0.0, 1.0, 0.5},
		{"large step", 0.0, 100.0, 50.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := midpointTowardsF32(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("midpointTowardsF32(%f, %f) = %f, expected %f",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestFloat64(t *testing.T) {
	gen := Float64(Size{Min: 0, Max: 100})
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a float64
	if value < 0 || value > 100 {
		t.Errorf("Float64().Generate() = %f, expected value in range [0, 100]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Float64().Generate() returned nil shrinker")
	}
}

func TestFloat64Range(t *testing.T) {
	gen := Float64Range(10.0, 20.0, false, false)
	r := rand.New(rand.NewSource(123))

	value, shrink := gen.Generate(r, Size{})

	// Test that we get a float64 in range
	if value < 10.0 || value > 20.0 {
		t.Errorf("Float64Range().Generate() = %f, expected value in range [10.0, 20.0]", value)
	}

	// Test that shrinker is not nil
	if shrink == nil {
		t.Error("Float64Range().Generate() returned nil shrinker")
	}
}

func TestFloat64Shrinker(t *testing.T) {
	// Test float64 shrinking behavior
	start, shrink := float64ShrinkInit(50.0, 0.0, 100.0, false, false)

	if start != 50.0 {
		t.Errorf("float64ShrinkInit() start = %f, expected 50.0", start)
	}

	if shrink == nil {
		t.Error("float64ShrinkInit() returned nil shrinker")
	}

	// Test shrinking behavior
	next, ok := shrink(false)
	if !ok {
		t.Error("Float64 shrinker returned false on first call")
	}

	// Test that value is in range
	if next < 0.0 || next > 100.0 {
		t.Errorf("Float64 shrinker returned value %f outside range [0.0, 100.0]", next)
	}
}

func TestFloat64HelperFunctions(t *testing.T) {
	// Test float64 helper functions
	tests := []struct {
		name string
		f    func() bool
	}{
		{"isFinite", func() bool { return isFinite(1.0) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test that the function doesn't panic
			tt.f()
		})
	}
}

func TestFloat64Clamp(t *testing.T) {
	// Test float64 clamp function
	tests := []struct {
		name     string
		x        float64
		min      float64
		max      float64
		expected float64
	}{
		{"in range", 5.0, 0.0, 10.0, 5.0},
		{"below min", -5.0, 0.0, 10.0, 0.0},
		{"above max", 15.0, 0.0, 10.0, 10.0},
		{"at min", 0.0, 0.0, 10.0, 0.0},
		{"at max", 10.0, 0.0, 10.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampF64(tt.x, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clampF64(%f, %f, %f) = %f, expected %f",
					tt.x, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestFloat64Target(t *testing.T) {
	// Test float64 target function
	tests := []struct {
		name     string
		min      float64
		max      float64
		expected float64
	}{
		{"zero in range", -10.0, 10.0, 0.0},
		{"zero at min", 0.0, 10.0, 0.0},
		{"zero at max", -10.0, 0.0, 0.0},
		{"all positive", 5.0, 15.0, 5.0},
		{"all negative", -15.0, -5.0, -5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := float64Target(tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("float64Target(%f, %f) = %f, expected %f",
					tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestFloat64MidpointTowards(t *testing.T) {
	// Test float64 midpoint towards function
	tests := []struct {
		name     string
		a        float64
		b        float64
		expected float64
	}{
		{"same values", 5.0, 5.0, 5.0},
		{"positive direction", 0.0, 10.0, 5.0},
		{"negative direction", 10.0, 0.0, 5.0},
		{"small step", 0.0, 1.0, 0.5},
		{"large step", 0.0, 100.0, 50.0},
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
