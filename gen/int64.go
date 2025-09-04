package gen

import (
	"math/rand"
)

// Int64 generates 64-bit integers with automatic range based on Size.
// If no Size is provided, uses [-100, 100].
func Int64(size Size) Generator[int64] {
	return From(func(r *rand.Rand, sz Size) (int64, Shrinker[int64]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		min, max := autoRange64(size, sz)
		if min > max {
			min, max = max, min
		}
		v := min + int64(r.Intn(int(max-min+1)))
		return int64ShrinkInit(v, min, max)
	})
}

// Int64Range generates int64 uniformly in the range [min, max] (inclusive).
func Int64Range(min, max int64) Generator[int64] {
	if min > max {
		min, max = max, min
	}
	return From(func(r *rand.Rand, _ Size) (int64, Shrinker[int64]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		v := min + int64(r.Intn(int(max-min+1)))
		return int64ShrinkInit(v, min, max)
	})
}

// ---------------- implementation / shrinking ----------------

// int64ShrinkInit initializes the shrinking process for an int64 value.
// It returns the initial value and a shrinker function that can generate
// progressively smaller candidates.
func int64ShrinkInit(start, min, max int64) (int64, Shrinker[int64]) {
	cur, last := clamp64(start, min, max), clamp64(start, min, max)

	queue := make([]int64, 0, 16)
	seen := map[int64]struct{}{cur: {}}

	push := func(x int64) {
		if x < min || x > max {
			return
		}
		if _, ok := seen[x]; ok {
			return
		}
		seen[x] = struct{}{}
		queue = append(queue, x)
	}
	target := shrinkTarget64(min, max)

	grow := func(base int64) {
		queue = queue[:0]
		// (1) target (0 if within range; otherwise closest bound)
		if base != target {
			push(target)
		}
		// (2) bisections towards the target
		if base != target {
			next := midpointTowards64(base, target)
			if next != base {
				push(next)
			}
			series := next
			for i := 0; i < 8 && series != target; i++ {
				series = midpointTowards64(series, target)
				if series != base {
					push(series)
				}
			}
		}
		// (3) unit step
		if base != target {
			push(stepTowards64(base, target))
		}
		// (4) bounds
		if base != min {
			push(min)
		}
		if base != max {
			push(max)
		}
	}
	grow(cur)

	pop := func() (int64, bool) {
		if len(queue) == 0 {
			return 0, false
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

	return cur, func(accept bool) (int64, bool) {
		if accept && last != cur {
			cur = last
			grow(cur)
		}
		nxt, ok := pop()
		if !ok {
			return 0, false
		}
		last = nxt
		return nxt, true
	}
}

// shrinkTarget64 returns the "natural" target to shrink towards for int64:
// - 0 if 0 ∈ [min,max]; otherwise, the bound closest to 0.
func shrinkTarget64(min, max int64) int64 {
	if min <= 0 && 0 <= max {
		return 0
	}
	if min > 0 {
		return min
	}
	return max
}

// clamp64 constrains an int64 value to be within the given bounds.
func clamp64(x, min, max int64) int64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// midpointTowards64 gives a "bisection step" from a towards b for int64,
// with rounding away from 'a' to guarantee progress.
func midpointTowards64(a, b int64) int64 {
	if a == b {
		return a
	}
	d := b - a
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

// stepTowards64 moves one unit step from a towards b for int64.
func stepTowards64(a, b int64) int64 {
	if a == b {
		return a
	}
	if b > a {
		return a + 1
	}
	return a - 1
}

// autoRange64 decides the final range for Int64(...) by combining the local "size" and the
// "size" coming from the runner. We prefer the largest range informed; if nothing is
// informed, we use [-100, 100].
func autoRange64(local, fromRunner Size) (int64, int64) {
	M := int64(0)
	for _, s := range []Size{local, fromRunner} {
		if abs := int64Abs(s.Min); abs > M {
			M = abs
		}
		if abs := int64Abs(s.Max); abs > M {
			M = abs
		}
	}
	if M == 0 {
		M = 100
	}
	return -M, M
}

// int64Abs returns the absolute value of an int as int64.
func int64Abs(x int) int64 {
	if x < 0 {
		return int64(-x)
	}
	return int64(x)
}
