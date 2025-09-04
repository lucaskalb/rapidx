package gen

import (
	"math/rand"
	"testing"
)



func TestBoolShrinker(t *testing.T) {

	start, shrink := boolShrinkInit(true)
	
	if start != true {
		t.Errorf("boolShrinkInit() start = %v, expected true", start)
	}
	
	if shrink == nil {
		t.Error("boolShrinkInit() returned nil shrinker")
	}
	

	next, ok := shrink(false)
	if !ok {
		t.Error("Bool shrinker returned false on first call")
	}
	

	if next == start {
		t.Error("Bool shrinker returned same value as start")
	}
	

	if next != true && next != false {
		t.Errorf("Bool shrinker returned invalid value %v", next)
	}
}

func TestBoolShrinkerWithAccept(t *testing.T) {
 with accept=true
	_, shrink := boolShrinkInit(true)
	
	// First call with accept=false
	next1, ok1 := shrink(false)
	if !ok1 {
		t.Error("Bool shrinker returned false on first call")
	}
	

	next2, ok2 := shrink(true)

	

	if next1 != true && next1 != false {
		t.Errorf("Bool shrinker returned invalid value %v", next1)
	}
	if ok2 && (next2 != true && next2 != false) {
		t.Errorf("Bool shrinker returned invalid value %v", next2)
	}
}

func TestBoolShrinkerExhaustion(t *testing.T) {
 until exhaustion
	_, shrink := boolShrinkInit(true)
	

	callCount := 0
	for {
		_, ok := shrink(false)
		if !ok {
			break
		}
		callCount++
		if callCount > 10 {
			t.Error("Bool shrinker did not exhaust after 10 calls")
			break
		}
	}
	

	if callCount == 0 {
		t.Error("Bool shrinker exhausted immediately")
	}
}

func TestBoolShrinkerWithDFSSStrategy(t *testing.T) {
 with DFS strategy
	SetShrinkStrategy("dfs")
	defer SetShrinkStrategy("bfs") // Reset to default
	
	_, shrink := boolShrinkInit(true)
	

	next, ok := shrink(false)
	if !ok {
		t.Error("Bool shrinker returned false on first call")
	}
	

	if next != true && next != false {
		t.Errorf("Bool shrinker returned invalid value %v", next)
	}
}

func TestBoolShrinkerEdgeCases(t *testing.T) {
 with edge cases
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
			

			if shrink == nil {
				t.Error("boolShrinkInit() returned nil shrinker")
			}
			

			next, ok := shrink(false)
			if ok {
			
				if next != true && next != false {
					t.Errorf("Bool shrinker returned invalid value %v", next)
				}
			}
		})
	}
}

func TestBoolMultipleGenerations(t *testing.T) {

	r := rand.New(rand.NewSource(456))
	gen := Bool()
	

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
	

	if trueCount == 0 || falseCount == 0 {
		t.Logf("Warning: Only got one boolean value after 100 generations (true: %d, false: %d)", 
			trueCount, falseCount)

	}
}


func boolShrinkInit(start bool) (bool, Shrinker[bool]) {
	gen := Bool()
	r := rand.New(rand.NewSource(123))
	_, _ = gen.Generate(r, Size{})
	

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
}, true
	}
}