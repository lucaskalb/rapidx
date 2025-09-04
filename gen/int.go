// File: gen/int.go
package gen

import (
	"math/rand"
)

// Int generates integers with automatic range based on Size:
// - if sz.Max (or |sz.Min|) > 0: range := [-M, M], where M = max(|sz.Min|, |sz.Max|)
// - otherwise, uses default range [-100, 100].
// Example: prop.ForAll(t, cfg, gen.Int(gen.Size{Max: 1000})) ...
func Int(size Size) Generator[int] {
	return From(func(r *rand.Rand, sz Size) (int, Shrinker[int]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		min, max := autoRange(size, sz) // decide the effective range
		if min > max {
			min, max = max, min
		}
		// generate uniformly
		v := min + r.Intn(max-min+1)
		return intShrinkInit(v, min, max)
	})
}

// IntRange generates integers uniformly in the range [min, max] (inclusive).
// Ignores sz for the range (useful when you want explicit control).
func IntRange(min, max int) Generator[int] {
	if min > max {
		min, max = max, min
	}
	return From(func(r *rand.Rand, _ Size) (int, Shrinker[int]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		v := min + r.Intn(max-min+1)
		return intShrinkInit(v, min, max)
	})
}

// -------------------- implementation / shrinking --------------------

// intShrinkInit initializes the shrinking process for an integer value.
// It returns the initial value and a shrinker function that can generate
// progressively smaller candidates.
func intShrinkInit(start, min, max int) (int, Shrinker[int]) {
	// current value (minimum known that fails) and last proposed
	cur := clamp(start, min, max)
	last := cur

	// queue of neighbors + deduplication
	queue := make([]int, 0, 16)
	seen := map[int]struct{}{cur: {}}

	push := func(x int) {
		if x < min || x > max {
			return
		}
		if _, ok := seen[x]; ok {
			return
		}
		seen[x] = struct{}{}
		queue = append(queue, x)
	}

	// neighbor heuristics:
	//  1) approach the target (0 if in range, otherwise closest bound)
	//  2) "halfway" towards the target (bisection)
	//  3) unit step towards the target (+/-1)
	//  4) bounds (min/max)
	growNeighbors := func(base int) {
		queue = queue[:0]
		target := shrinkTarget(min, max) // 0 if possible; otherwise closest bound

		// (1) direct target
		if base != target {
			push(target)
		}

		// (2) halfway towards the target (bisection)
		if base != target {
			next := midpointTowards(base, target)
			if next != base {
				push(next)
			}
			// multiple bisections rounding away from base
			// (generates series base -> base' -> ... -> target)
			series := next
			for i := 0; i < 8; i++ {
				if series == target {
					break
				}
				series = midpointTowards(series, target)
				if series != base {
					push(series)
				}
			}
		}

		// (3) unit step towards the target
		if base != target {
			step := stepTowards(base, target)
			if step != base {
				push(step)
			}
		}

		// (4) bounds
		if base != min {
			push(min)
		}
		if base != max {
			push(max)
		}
	}

	growNeighbors(cur)

	pop := func() (int, bool) {
		if len(queue) == 0 {
			return 0, false
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

	return cur, func(accept bool) (int, bool) {
		// If the last candidate was ACCEPTED (still fails), rebase on it
		if accept {
			if last != cur {
				cur = last
				growNeighbors(cur)
			}
		}
		// propose the next neighbor
		nxt, ok := pop()
		if !ok {
			return 0, false
		}
		last = nxt
		return nxt, true
	}
}

// shrinkTarget returns the "natural" target to shrink towards:
// - 0 if 0 ∈ [min,max]; otherwise, the bound closest to 0.
func shrinkTarget(min, max int) int {
	if min <= 0 && 0 <= max {
		return 0
	}
	// outside range: take the bound closest to 0
	if min > 0 {
		// all positive range -> min is closest to 0
		return min
	}
	// all negative range -> max is closest to 0 (e.g., [-10, -1] → -1)
	return max
}

// midpointTowards gives a "bisection step" from a towards b,
// with rounding away from 'a' to guarantee progress.
func midpointTowards(a, b int) int {
	if a == b {
		return a
	}
	d := b - a
	// round "up" in magnitude to not get stuck when |d| == 1
	step := d / 2
	if step == 0 {
		if d > 0 {
			step = 1
		} else {
			step = -1
		}
	}
	return a + step
}

// stepTowards moves one unit step from a towards b.
func stepTowards(a, b int) int {
	if a == b {
		return a
	}
	if b > a {
		return a + 1
	}
	return a - 1
}

// autoRange decides the final range for Int(...) by combining the local "size" and the
// "size" coming from the runner. We prefer the largest range informed; if nothing is
// informed, we use [-100, 100].
func autoRange(local, fromRunner Size) (int, int) {
	// choose an "M" (magnitude) based on the largest absolute value seen
	M := 0
	for _, s := range []Size{local, fromRunner} {
		M = maxInt(M, absInt(s.Min))
		M = maxInt(M, absInt(s.Max))
	}
	if M == 0 {
		M = 100
	}
	return -M, M
}

// clamp constrains a value to be within the given bounds.
func clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// absInt returns the absolute value of an integer.
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
