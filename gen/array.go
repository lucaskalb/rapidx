package gen

import "math/rand"

// ArrayOf generates a slice of **exact** length n, using the element generator.
// It is "array-like": great when you need to simulate [N]T.
// Shrink: cannot remove elements; only tries local shrink at each position,
// exploring multiple branches (BFS/DFS) and deduplicating candidates.
func ArrayOf[T any](elem Generator[T], n int) Generator[[]T] {
	return From(func(r *rand.Rand, _ Size) ([]T, Shrinker[[]T]) {
		if r == nil {
			// Using math/rand for deterministic property-based testing
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		if n < 0 {
			n = 0
		}

		// generate values + element shrinkers
		cur := make([]T, n)
		elS := make([]Shrinker[T], n)
		for i := 0; i < n; i++ {
			v, s := elem.Generate(r, Size{})
			cur[i], elS[i] = v, s
		}

		queue := make([][]T, 0, 32)
		seen := map[string]struct{}{sig(cur): {}}
		var last []T

		push := func(s []T) {
			k := sig(s)
			if _, ok := seen[k]; ok {
				return
			}
			seen[k] = struct{}{}
			cp := append(([]T)(nil), s...)
			queue = append(queue, cp)
		}

		// Generate neighbors by trying to "tame" each position with one local shrink step
		grow := func(base []T) {
			queue = queue[:0]
			L := len(base)
			for i := L - 1; i >= 0; i-- {
				if elS[i] == nil {
					continue
				}
				if nv, ok := elS[i](false); ok { // propose 1 candidate for position i
					cand := append(([]T)(nil), base...)
					cand[i] = nv
					push(cand)
				}
			}
		}
		grow(cur)

		pop := func() ([]T, bool) {
			if len(queue) == 0 {
				return nil, false
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

		return cur, func(accept bool) ([]T, bool) {
			if accept {
				// rebase: new minimum becomes the last accepted candidate
				if last != nil && sig(last) != sig(cur) {
					cur = last
					// after rebase, we "forget" element shrinkers to maintain simplicity;
					// still, in the next layer grow() will repropose one step per position.
					for i := range elS {
						elS[i] = nil
					}
					grow(cur)
				}
			}
			nxt, ok := pop()
			if !ok {
				return nil, false
			}
			last = nxt
			return nxt, true
		}
	})
}
