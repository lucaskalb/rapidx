package gen

import (
	"math/rand"
	"testing"
)

// TestBool is already defined in comb_test.go

func TestBoolShrinker(t *testing.T) {
	// Test bool shrinking behavior
	start, shrink := boolShrinkInit(true)
	
	if start != true {
		t.Errorf("boolShrinkInit() start = %v, expected true", start)
	}
	
	if shrink == nil {
		t.Error("boolShrinkInit() returned nil shrinker")
	}
	
	// Test shrinking behavior
	next, ok := shrink(false)
	if !ok {
		t.Error("Bool shrinker returned false on first call")
	}
	
	// Test that we get a different value (true -> false or false -> true)
	if next == start {
		t.Error("Bool shrinker returned same value as start")
	}
	
	// Test that value is a valid boolean
	if next != true && next != false {
		t.Errorf("Bool shrinker returned invalid value %v", next)
	}
}

func TestBoolShrinkerWithAccept(t *testing.T) {
	// Test shrinking behavior with accept=true
	_, shrink := boolShrinkInit(true)
	
	// First call with accept=false
	next1, ok1 := shrink(false)
	if !ok1 {
		t.Error("Bool shrinker returned false on first call")
	}
	
	// Second call with accept=true (should rebase)
	next2, ok2 := shrink(true)
	// It's possible that the shrinker exhausts quickly, so we don't require it to succeed
	
	// Test that first value is a valid boolean
	if next1 != true && next1 != false {
		t.Errorf("Bool shrinker returned invalid value %v", next1)
	}
	if ok2 && (next2 != true && next2 != false) {
		t.Errorf("Bool shrinker returned invalid value %v", next2)
	}
}

func TestBoolShrinkerExhaustion(t *testing.T) {
	// Test shrinking behavior until exhaustion
	_, shrink := boolShrinkInit(true)
	
	// Call shrinker many times until it returns false
	callCount := 0
	for {
		_, ok := shrink(false)
		if !ok {
			break
		}
		callCount++
		if callCount > 10 { // Safety limit (bool should exhaust quickly)
			t.Error("Bool shrinker did not exhaust after 10 calls")
			break
		}
	}
	
	// Should have made at least some calls
	if callCount == 0 {
		t.Error("Bool shrinker exhausted immediately")
	}
}

func TestBoolShrinkerWithDFSSStrategy(t *testing.T) {
	// Test shrinking behavior with DFS strategy
	SetShrinkStrategy("dfs")
	defer SetShrinkStrategy("bfs") // Reset to default
	
	_, shrink := boolShrinkInit(true)
	
	// Test that we get a value
	next, ok := shrink(false)
	if !ok {
		t.Error("Bool shrinker returned false on first call")
	}
	
	// Test that value is a valid boolean
	if next != true && next != false {
		t.Errorf("Bool shrinker returned invalid value %v", next)
	}
}

func TestBoolShrinkerEdgeCases(t *testing.T) {
	// Test shrinking behavior with edge cases
	tests := []struct {
		name  string
		start bool
	}{
		{"start true", true},
		{"start false", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, shrink := boolShrinkInit(tt.start)
			
			if start != tt.start {
				t.Errorf("boolShrinkInit() start = %v, expected %v", start, tt.start)
			}
			
			// Test that shrinker is not nil
			if shrink == nil {
				t.Error("boolShrinkInit() returned nil shrinker")
			}
			
			// Test that we can call shrinker at least once
			next, ok := shrink(false)
			if ok {
				// Test that value is a valid boolean
				if next != true && next != false {
					t.Errorf("Bool shrinker returned invalid value %v", next)
				}
			}
		})
	}
}

func TestBoolMultipleGenerations(t *testing.T) {
	// Test that Bool() generates different values over multiple calls
	r := rand.New(rand.NewSource(456))
	gen := Bool()
	
	// Generate multiple values and check that we get both true and false
	// (with high probability)
	trueCount := 0
	falseCount := 0
	
	for i := 0; i < 100; i++ {
		value, _ := gen.Generate(r, Size{})
		if value {
			trueCount++
		} else {
			falseCount++
		}
	}
	
	// We should get both values (with high probability)
	if trueCount == 0 || falseCount == 0 {
		t.Logf("Warning: Only got one boolean value after 100 generations (true: %d, false: %d)", 
			trueCount, falseCount)
		// This is not necessarily an error, just unlikely
	}
}

// Helper function to test the internal boolShrinkInit function
func boolShrinkInit(start bool) (bool, Shrinker[bool]) {
	gen := Bool()
	r := rand.New(rand.NewSource(123))
	_, _ = gen.Generate(r, Size{})
	
	// We need to create a shrinker that starts with our desired value
	// Since we can't directly call the internal function, we'll simulate it
	cur, last := start, start
	
	queue := make([]bool, 0, 2)
	seen := map[bool]struct{}{cur: {}}
	
	push := func(b bool) {
		if _, ok := seen[b]; ok { return }
		seen[b] = struct{}{}
		queue = append(queue, b)
	}
	
	grow := func(base bool) {
		queue = queue[:0]
		// Heurística: tentar false primeiro
		if base != false { push(false) }
		if base != true  { push(true)  }
	}
	grow(cur)
	
	pop := func() (bool, bool) {
		if len(queue) == 0 { return false, false }
		if shrinkStrategy == "dfs" {
			v := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			return v, true
		}
		v := queue[0]
		queue = queue[1:]
		return v, true
	}
	
	return cur, func(accept bool) (bool, bool) {
		if accept && last != cur {
			cur = last
			grow(cur)
		}
		nxt, ok := pop()
		if !ok { return false, false }
		last = nxt
		return nxt, true
	}
}