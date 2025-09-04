package gen

import (
	"math/rand"
)

// Uint generates unsigned integers with automatic range based on Size.
// If no Size is provided, uses [0, 100].
func Uint(size Size) Generator[uint] {
	return From(func(r *rand.Rand, sz Size) (uint, Shrinker[uint]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		min, max := autoRangeUint(size, sz) // [min,max], min>=0
		if min > max {
			min, max = max, min
		}
		v := min + uint(r.Intn(int(max-min+1)))
		return uintShrinkInit(v, min, max)
	})
}

// UintRange generates uint uniformly in the range [min, max].
func UintRange(min, max uint) Generator[uint] {
	if min > max {
		min, max = max, min
	}
	return From(func(r *rand.Rand, _ Size) (uint, Shrinker[uint]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		v := min + uint(r.Intn(int(max-min+1)))
		return uintShrinkInit(v, min, max)
	})
}

// ---------------- implementation / shrinking ----------------

// uintShrinkInit initializes the shrinking process for a uint value.
// It returns the initial value and a shrinker function that can generate
// progressively smaller candidates.
func uintShrinkInit(start, min, max uint) (uint, Shrinker[uint]) {
	cur, last := clampU(start, min, max), clampU(start, min, max)

	queue := make([]uint, 0, 16)
	seen := map[uint]struct{}{cur: {}}

	push := func(x uint) {
		if x < min || x > max {
			return
		}
		if _, ok := seen[x]; ok {
			return
		}
		seen[x] = struct{}{}
		queue = append(queue, x)
	}

	grow := func(base uint) {
		queue = queue[:0]
		// (1) natural target for uint is 0
		if base != 0 {
			push(0)
		}
		// (2) bisections towards 0
		if base != 0 {
			next := base / 2
			if next != base {
				push(next)
			}
			series := next
			for i := 0; i < 8 && series > 0; i++ {
				series /= 2
				push(series)
			}
		}
		// (3) unit step towards 0
		if base > 0 {
			push(base - 1)
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

	pop := func() (uint, bool) {
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

	return cur, func(accept bool) (uint, bool) {
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

// autoRangeUint decides the final range for Uint(...) by combining the local "size" and the
// "size" coming from the runner. We prefer the largest range informed; if nothing is
// informed, we use [0, 100].
func autoRangeUint(local, fromRunner Size) (uint, uint) {
	M := 0
	for _, s := range []Size{local, fromRunner} {
		if s.Max > M {
			M = s.Max
		}
	}
	if M == 0 {
		M = 100
	}
	return 0, uint(M)
}

// clampU constrains a uint value to be within the given bounds.
func clampU(x, min, max uint) uint {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
