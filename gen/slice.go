package gen

import (
	"fmt"
	"math/rand"
)

// SliceOf generates []T from an element generator.
// - size.Min/Max control the length (default Min=0, Max=16).
// Shrink:
//
//	(1) remove large blocks (half, quarter, ...) → remove indices
//	(2) remove isolated element (right→left)
//	(3) try shrink on elements (propagating accept)
func SliceOf[T any](elem Generator[T], size Size) Generator[[]T] {
	return From(func(r *rand.Rand, sz Size) ([]T, Shrinker[[]T]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		// defaults
		if size.Min == 0 && size.Max == 0 {
			size.Min, size.Max = 0, 16
		}
		if sz.Min != 0 || sz.Max != 0 {
			size = sz
		}
		if size.Max < size.Min {
			size.Max = size.Min
		}

		// length
		n := size.Min
		if size.Max > size.Min {
			n += r.Intn(size.Max - size.Min + 1)
		}

		// generate elems + capture shrinkers
		vals := make([]T, n)
		shks := make([]Shrinker[T], n)
		for i := 0; i < n; i++ {
			v, s := elem.Generate(r, Size{})
			vals[i], shks[i] = v, s
		}
		cur := append(([]T)(nil), vals...) // snapshot

		// dedup by textual "signature" (ok for testing; avoids cycles)
		seen := map[string]struct{}{sig(cur): {}}
		queue := make([][]T, 0, 64)
		var last []T

		push := func(s []T) {
			k := sig(s)
			if _, ok := seen[k]; ok {
				return
			}
			seen[k] = struct{}{}
			// copy to avoid sharing backing array
			cp := append(([]T)(nil), s...)
			queue = append(queue, cp)
		}

		// remove intervals [i:j) from cur
		rem := func(base []T, i, j int) []T {
			out := make([]T, 0, len(base)-(j-i))
			out = append(out, base[:i]...)
			out = append(out, base[j:]...)
			return out
		}

		growNeighbors := func(base []T) {
			queue = queue[:0]
			L := len(base)
			if L == 0 {
				return
			}
			// (1) remove large blocks (binary: half, quarter, ...)
			chunk := L / 2
			for chunk >= 1 {
				for i := 0; i+chunk <= L; i += chunk {
					push(rem(base, i, i+chunk))
				}
				chunk /= 2
			}
			// (2) remove isolated element (R->L)
			for i := L - 1; i >= 0; i-- {
				push(rem(base, i, i+1))
			}
			// (3) shrink elements locally, maintaining size
			//     (generates one neighbor per position with 1 shrink step)
			for i := L - 1; i >= 0; i-- {
				if shks == nil || shks[i] == nil {
					continue
				}
				if nv, ok := shks[i](false); ok { // false: proposing candidate
					cand := append(([]T)(nil), base...)
					cand[i] = nv
					push(cand)
				}
			}
		}
		growNeighbors(cur)

		pop := func() ([]T, bool) {
			if len(queue) == 0 {
				return nil, false
			}
			if shrinkStrategy == "dfs" {
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
				// rebase on the last accepted candidate
				if last != nil && sig(last) != sig(cur) {
					cur = last
					// IMPORTANT: when we rebase, we need to regenerate shrinkers
					// to maintain element consistency (can be expensive, but simple)
					shks = make([]Shrinker[T], len(cur))
					for i := range cur {
						// rebuild "focal" shrinker starting from current value:
						v := cur[i]
						// trick: create a Const(v) generator and get its shrinker (doesn't have one).
						// so, better: if we want future shrink on elements, we need to
						// accept that we'll only have 1 step in the neighbor (already done in growNeighbors).
						// To keep it simple in MVP, we don't retain shrinkers after rebase.
						shks[i] = nil
						_ = v
					}
					growNeighbors(cur)
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

// sig creates a simplified textual signature of a generic slice.
// For shrinking dedup purposes in tests, this is sufficient.
func sig[T any](s []T) string { return fmt.Sprintf("%#v", s) }
