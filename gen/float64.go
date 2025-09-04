package gen

import (
	"math"
	"math/rand"
)

// Float64 generates floats with automatic range based on Size.
// - If no Size is provided, uses range [-100, 100].
// - Does not include NaN/Inf (focused on business numeric cases).
func Float64(size Size) Generator[float64] {
	return From(func(r *rand.Rand, sz Size) (float64, Shrinker[float64]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		min, max := autoRangeF64(size, sz)
		if min > max {
			min, max = max, min
		}
		v := uniformF64(r, min, max)
		return float64ShrinkInit(v, min, max, false, false)
	})
}

// autoRangeF64 decides the final range for Float64(...) by combining the local "size" and the
// "size" coming from the runner. We prefer the largest range informed; if nothing is
// informed, we use [-100, 100].
func autoRangeF64(local, fromRunner Size) (float64, float64) {
	M := 0
	for _, s := range []Size{local, fromRunner} {
		if a := absInt(s.Min); a > M {
			M = a
		}
		if a := absInt(s.Max); a > M {
			M = a
		}
	}
	if M == 0 {
		M = 100
	}
	return -float64(M), float64(M)
}

// Float64Range generates floats uniformly in [min, max] (inclusive on finite bounds).
// Parameters includeNaN/includeInf allow injecting special cases.
func Float64Range(min, max float64, includeNaN, includeInf bool) Generator[float64] {
	if min > max {
		min, max = max, min
	}
	return From(func(r *rand.Rand, _ Size) (float64, Shrinker[float64]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		v := uniformF64(r, min, max)
		// small chance of specials, if enabled
		if includeNaN && r.Intn(50) == 0 {
			v = math.NaN()
		} else if includeInf && r.Intn(50) == 1 {
			if r.Intn(2) == 0 {
				v = math.Inf(+1)
			} else {
				v = math.Inf(-1)
			}
		}
		return float64ShrinkInit(v, min, max, includeNaN, includeInf)
	})
}

// ---------------- implementation / shrinking ----------------

// float64ShrinkInit initializes the shrinking process for a float64 value.
// It returns the initial value and a shrinker function that can generate
// progressively smaller candidates.
func float64ShrinkInit(start, min, max float64, allowNaN, allowInf bool) (float64, Shrinker[float64]) {
	cur := clampF64(start, min, max) // NaN stays as NaN; clamp doesn't alter NaN
	last := cur

	queue := make([]float64, 0, 32)
	seen := map[uint64]struct{}{f64key(cur): {}}

	push := func(x float64) {
		// respect range when finite; for Inf/NaN, push if allowed
		if math.IsNaN(x) && !allowNaN {
			return
		}
		if math.IsInf(x, 0) && !allowInf {
			return
		}
		if isFinite(x) && isFinite(min) && isFinite(max) {
			if x < min || x > max {
				return
			}
		}
		k := f64key(x)
		if _, ok := seen[k]; ok {
			return
		}
		seen[k] = struct{}{}
		queue = append(queue, x)
	}

	grow := func(base float64) {
		queue = queue[:0]

		// Special cases first
		if math.IsNaN(base) {
			// NaN -> try 0, 1, -1, ±Inf (if allowed), bounds
			push(0)
			push(1)
			push(-1)
			if allowInf {
				push(math.Inf(+1))
				push(math.Inf(-1))
			}
			if isFinite(min) {
				push(min)
			}
			if isFinite(max) {
				push(max)
			}
			return
		}
		if math.IsInf(base, 0) {
			// +Inf/-Inf -> approach by appropriate bound, then towards 0
			if math.IsInf(base, +1) && isFinite(max) {
				push(max)
			}
			if math.IsInf(base, -1) && isFinite(min) {
				push(min)
			}
			push(0)
			return
		}

		// Finite: normal heuristic towards 0
		target := float64Target(min, max) // 0 if possible; otherwise bound closest to 0
		if base != target {
			push(target)
		}

		// Bisections (halfway to target)
		if base != target {
			next := midpointTowardsF64(base, target)
			if next != base {
				push(next)
			}
			series := next
			for i := 0; i < 8 && series != target; i++ {
				series = midpointTowardsF64(series, target)
				if series != base {
					push(series)
				}
			}
		}

		// Step towards target using Nextafter
		if base != target {
			step := math.Nextafter(base, target)
			if step != base {
				push(step)
			}
		}

		// Try to change sign if this approaches 0 (e.g., -x -> +x when target=0)
		if target == 0 && base != 0 && !math.Signbit(base) {
			// base>0: try -base (may not always be "smaller", but helps)
			push(-base)
		} else if target == 0 && base != 0 && math.Signbit(base) {
			push(-base)
		}

		// Bounds (if finite)
		if isFinite(min) && base != min {
			push(min)
		}
		if isFinite(max) && base != max {
			push(max)
		}
	}

	grow(cur)

	pop := func() (float64, bool) {
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

	return cur, func(accept bool) (float64, bool) {
		if accept && f64key(last) != f64key(cur) {
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

// ---------- helpers float64 ----------

// isFinite checks if a float64 value is finite (not NaN or Inf).
func isFinite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

// f64key creates a unique key for a float64 value using its bit representation.
func f64key(x float64) uint64 { return math.Float64bits(x) }

// clampF64 constrains a float64 value to be within the given bounds.
func clampF64(x, min, max float64) float64 {
	if !isFinite(x) {
		return x
	}
	if isFinite(min) && x < min {
		return min
	}
	if isFinite(max) && x > max {
		return max
	}
	return x
}

// uniformF64 generates a uniform random float64 in the given range.
func uniformF64(r *rand.Rand, min, max float64) float64 {
	if isFinite(min) && isFinite(max) && max >= min {
		if min == max {
			return min
		}
		return min + r.Float64()*(max-min)
	}
	// if invalid range, fall back to default [-100, 100]
	return -100 + r.Float64()*200
}

// 0 within range → target=0; otherwise bound closest to 0
func float64Target(min, max float64) float64 {
	if isFinite(min) && isFinite(max) && min <= 0 && 0 <= max {
		return 0
	}
	// outside range: take bound closest to 0
	if !isFinite(min) && !isFinite(max) {
		return 0
	}
	// choose bound with smaller |x|
	amin := math.Abs(min)
	amax := math.Abs(max)
	if amin < amax {
		return min
	}
	return max
}

// "bisection" step from a -> b
func midpointTowardsF64(a, b float64) float64 {
	if a == b {
		return a
	}
	return a + (b-a)/2
}
