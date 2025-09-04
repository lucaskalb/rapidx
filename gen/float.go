package gen

import (
	"math"
	"math/rand"
)

// Float32 generates float32 values with automatic range based on Size.
// Default: [-100, 100]. Does not include NaN/Inf.
func Float32(size Size) Generator[float32] {
	return From(func(r *rand.Rand, sz Size) (float32, Shrinker[float32]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		min, max := autoRangeF32(size, sz)
		if min > max {
			min, max = max, min
		}
		v := uniformF32(r, min, max)
		return float32ShrinkInit(v, min, max, false, false)
	})
}

// Float32Range generates float32 in [min, max]; can optionally produce NaN/±Inf.
func Float32Range(min, max float32, includeNaN, includeInf bool) Generator[float32] {
	if min > max {
		min, max = max, min
	}
	return From(func(r *rand.Rand, _ Size) (float32, Shrinker[float32]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		v := uniformF32(r, min, max)
		if includeNaN && r.Intn(50) == 0 {
			v = float32(math.NaN())
		} else if includeInf && r.Intn(50) == 1 {
			if r.Intn(2) == 0 {
				v = float32(math.Inf(+1))
			} else {
				v = float32(math.Inf(-1))
			}
		}
		return float32ShrinkInit(v, min, max, includeNaN, includeInf)
	})
}

// -------------- implementation / shrinking (float32) --------------

// float32ShrinkInit initializes the shrinking process for a float32 value.
// It returns the initial value and a shrinker function that can generate
// progressively smaller candidates.
func float32ShrinkInit(start, min, max float32, allowNaN, allowInf bool) (float32, Shrinker[float32]) {
	cur := clampF32(start, min, max)
	last := cur

	queue := make([]float32, 0, 32)
	seen := map[uint32]struct{}{f32key(cur): {}}

	push := func(x float32) {
		if float32IsNaN(x) && !allowNaN {
			return
		}
		if float32IsInf(x) && !allowInf {
			return
		}
		if float32IsFinite(x) && float32IsFinite(min) && float32IsFinite(max) {
			if x < min || x > max {
				return
			}
		}
		k := f32key(x)
		if _, ok := seen[k]; ok {
			return
		}
		seen[k] = struct{}{}
		queue = append(queue, x)
	}

	grow := func(base float32) {
		queue = queue[:0]

		if float32IsNaN(base) {
			push(0)
			push(1)
			push(-1)
			if allowInf {
				push(float32(math.Inf(+1)))
				push(float32(math.Inf(-1)))
			}
			if float32IsFinite(min) {
				push(min)
			}
			if float32IsFinite(max) {
				push(max)
			}
			return
		}
		if float32IsInf(base) {
			if math.IsInf(float64(base), +1) && float32IsFinite(max) {
				push(max)
			}
			if math.IsInf(float64(base), -1) && float32IsFinite(min) {
				push(min)
			}
			push(0)
			return
		}

		// Finite
		target := float32Target(min, max)
		if base != target {
			push(target)
		}

		if base != target {
			next := midpointTowardsF32(base, target)
			if next != base {
				push(next)
			}
			series := next
			for i := 0; i < 8 && series != target; i++ {
				series = midpointTowardsF32(series, target)
				if series != base {
					push(series)
				}
			}
		}

		if base != target {
			step := math.Nextafter32(base, target)
			if step != base {
				push(step)
			}
		}

		// try to flip sign if target=0
		if target == 0 && base != 0 {
			push(-base)
		}

		if float32IsFinite(min) && base != min {
			push(min)
		}
		if float32IsFinite(max) && base != max {
			push(max)
		}
	}

	grow(cur)

	pop := func() (float32, bool) {
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

	return cur, func(accept bool) (float32, bool) {
		if accept && f32key(last) != f32key(cur) {
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

// ---------- helpers float32 ----------

// float32IsFinite checks if a float32 value is finite (not NaN or Inf).
func float32IsFinite(x float32) bool { return !math.IsNaN(float64(x)) && !math.IsInf(float64(x), 0) }

// float32IsNaN checks if a float32 value is NaN.
func float32IsNaN(x float32) bool { return math.IsNaN(float64(x)) }

// float32IsInf checks if a float32 value is infinite.
func float32IsInf(x float32) bool { return math.IsInf(float64(x), 0) }

// f32key creates a unique key for a float32 value using its bit representation.
func f32key(x float32) uint32 { return math.Float32bits(x) }

// clampF32 constrains a float32 value to be within the given bounds.
func clampF32(x, min, max float32) float32 {
	if !float32IsFinite(x) {
		return x
	}
	if float32IsFinite(min) && x < min {
		return min
	}
	if float32IsFinite(max) && x > max {
		return max
	}
	return x
}

// autoRangeF32 decides the final range for Float32(...) by combining the local "size" and the
// "size" coming from the runner. We prefer the largest range informed; if nothing is
// informed, we use [-100, 100].
func autoRangeF32(local, fromRunner Size) (float32, float32) {
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
	return -float32(M), float32(M)
}

// uniformF32 generates a uniform random float32 in the given range.
func uniformF32(r *rand.Rand, min, max float32) float32 {
	if float32IsFinite(min) && float32IsFinite(max) && max >= min {
		if min == max {
			return min
		}
		return min + float32(r.Float64())*(max-min)
	}
	return -100 + float32(r.Float64())*200
}

// float32Target returns the "natural" target to shrink towards for float32:
// - 0 if 0 ∈ [min,max]; otherwise, the bound closest to 0.
func float32Target(min, max float32) float32 {
	if float32IsFinite(min) && float32IsFinite(max) && min <= 0 && 0 <= max {
		return 0
	}
	if !float32IsFinite(min) && !float32IsFinite(max) {
		return 0
	}
	amin := float32(math.Abs(float64(min)))
	amax := float32(math.Abs(float64(max)))
	if amin < amax {
		return min
	}
	return max
}

// midpointTowardsF32 gives a "bisection step" from a towards b for float32.
func midpointTowardsF32(a, b float32) float32 {
	if a == b {
		return a
	}
	return a + (b-a)/2
}
