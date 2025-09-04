// File: gen/comb.go
package gen

import (
	"math/rand"
)

// -------------------------
// Basic helpers
// -------------------------

// Const always returns the same value (without shrinking).
func Const[T any](v T) Generator[T] {
	return From(func(_ *rand.Rand, _ Size) (T, Shrinker[T]) {
		return v, func(bool) (T, bool) { var z T; return z, false }
	})
}

// OneOf chooses uniformly from one of the generators.
func OneOf[T any](gs ...Generator[T]) Generator[T] {
	return Weighted(func(_ T) float64 { return 1.0 }, gs...)
}

// Weighted chooses a generator based on dynamic weights (by value).
// The strategy here captures which index was selected to be able to "shrink"
// reusing the shrinker of the chosen generator. Optionally, in shrinking
// it also tries to migrate to neighbors (other indices) — controlled by `tryNeighbors`.
func Weighted[T any](weight func(T) float64, gs ...Generator[T]) Generator[T] {
	if len(gs) == 0 {
		panic("gen.Weighted: needs at least one generator")
	}
	return From(func(r *rand.Rand, sz Size) (T, Shrinker[T]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		// step 1: choose generator
		idx := r.Intn(len(gs))
		val, shrink := gs[idx].Generate(r, sz)

		// queue of neighbors (other generators) to try during shrinking
		neighbors := make([]int, 0, len(gs)-1)
		for i := range gs {
			if i != idx {
				neighbors = append(neighbors, i)
			}
		}

		return val, func(accept bool) (T, bool) {
			// if the last candidate was accepted (failed), we continue shrinking the same generator
			if accept {
				if next, ok := shrink(true); ok {
					return next, true
				}
				// exhausted internal shrink → try to migrate to a neighbor
				for len(neighbors) > 0 {
					j := neighbors[0]
					neighbors = neighbors[1:]
					nv, ns := gs[j].Generate(r, sz)
					// update "context" for the new generator
					idx, val, shrink = j, nv, ns
					return val, true
				}
				var z T
				return z, false
			}
			// candidate was rejected → try another from the same shrinker
			if next, ok := shrink(false); ok {
				return next, true
			}
			// or migrate to a neighbor
			for len(neighbors) > 0 {
				j := neighbors[0]
				neighbors = neighbors[1:]
				nv, ns := gs[j].Generate(r, sz)
				idx, val, shrink = j, nv, ns
				return val, true
			}
			var z T
			return z, false
		}
	})
}

// -------------------------
// Combinators
// -------------------------

// Map applies f: A -> B preserving shrinking (maps A's candidates).
func Map[A, B any](ga Generator[A], f func(A) B) Generator[B] {
	return From(func(r *rand.Rand, sz Size) (B, Shrinker[B]) {
		a, sa := ga.Generate(r, sz)
		b := f(a)
		return b, func(accept bool) (B, bool) {
			na, ok := sa(accept)
			if !ok {
				var z B
				return z, false
			}
			return f(na), true
		}
	})
}

// Filter keeps only values that satisfy pred.
// Implements "rebase" in shrink: when accepting, shrinks on top of the new minimum
// ensuring that the next candidates also satisfy the predicate.
func Filter[T any](g Generator[T], pred func(T) bool, maxTries int) Generator[T] {
	if maxTries <= 0 {
		maxTries = 1000
	}
	return From(func(r *rand.Rand, sz Size) (T, Shrinker[T]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		// generate a value that passes the pred
		var v T
		var s Shrinker[T]
		okv := false
		for tries := 0; tries < maxTries; tries++ {
			v, s = g.Generate(r, sz)
			if pred(v) {
				okv = true
				break
			}
		}
		if !okv {
			var z T
			return z, func(bool) (T, bool) { return z, false }
		}

		// shrinker: whenever we accept, we need to "rebase" and continue
		// ensuring pred on the next candidates.
		return v, func(accept bool) (T, bool) {
			for {
				nv, ok := s(accept)
				if !ok {
					var z T
					return z, false
				}
				if pred(nv) {
					return nv, true
				}
				// candidate doesn't satisfy pred → reject and try next
				accept = false
			}
		}
	})
}

// Bind (flatMap): the output generator depends on the value generated in A.
// Shrinking: first tries to shrink in B; when exhausted, shrinks in A and regenerates B.
func Bind[A, B any](ga Generator[A], f func(A) Generator[B]) Generator[B] {
	return From(func(r *rand.Rand, sz Size) (B, Shrinker[B]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		a, sa := ga.Generate(r, sz)
		gb := f(a)
		b, sb := gb.Generate(r, sz)

		state := 0 // 0 => shrink B; 1 => shrink A (and regenerate B)

		return b, func(accept bool) (B, bool) {
			switch state {
			case 0:
				if nb, ok := sb(accept); ok {
					return nb, true
				}
				// exhausted shrink of B → we move to shrink of A
				state = 1
				accept = false // first step in A is "reject" to get next candidate
				fallthrough
			case 1:
				na, ok := sa(accept)
				if !ok {
					var z B
					return z, false
				}
				// regenerate B based on the new A
				a = na
				gb = f(a)
				b, sb = gb.Generate(r, sz)
				return b, true
			default:
				var z B
				return z, false
			}
		}
	})
}
