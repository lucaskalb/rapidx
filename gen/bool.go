package gen

import "math/rand"

// Bool generates boolean values uniformly.
// Shrink: prioritizes reducing to false (smaller counterexample by convention).
func Bool() Generator[bool] {
	return From(func(r *rand.Rand, _ Size) (bool, Shrinker[bool]) {
		if r == nil {
			// Using math/rand for deterministic property-based testing
			r = rand.New(rand.NewSource(rand.Int63())) // #nosec G404 -- Using math/rand for deterministic property-based testing
		}
		v := r.Intn(2) == 0 // true/false
		cur, last := v, v

		queue := make([]bool, 0, 2)
		seen := map[bool]struct{}{cur: {}}

		push := func(b bool) {
			if _, ok := seen[b]; ok {
				return
			}
			seen[b] = struct{}{}
			queue = append(queue, b)
		}

		grow := func(base bool) {
			queue = queue[:0]
			// Heuristic: try false first
			if base {
				push(false)
			}
			if !base {
				push(true)
			}
		}
		grow(cur)

		pop := func() (bool, bool) {
			if len(queue) == 0 {
				return false, false
			}
			if shrinkStrategy == ShrinkStrategyDFS {
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
			if !ok {
				return false, false
			}
			last = nxt
			return nxt, true
		}
	})
}
