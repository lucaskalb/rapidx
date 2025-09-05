package gen

import "math/rand"

// Uint64 generates unsigned 64-bit integers with automatic range based on Size.
// If nothing is provided, uses [0, 100].
func Uint64(size Size) Generator[uint64] {
	return From(func(r *rand.Rand, sz Size) (uint64, Shrinker[uint64]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63())) // #nosec G404 -- Using math/rand for deterministic property-based testing
		}
		min, max := autoRangeUnsigned[uint64](size, sz)
		if min > max {
			min, max = max, min
		}
		v := min + uint64(r.Intn(int(max-min+1))) // #nosec G115 -- Safe for property-based testing ranges
		return unsignedShrinkInit(v, min, max)
	})
}

// Uint64Range generates uint64 uniformly in the range [min, max] (inclusive).
func Uint64Range(min, max uint64) Generator[uint64] {
	if min > max {
		min, max = max, min
	}
	return From(func(r *rand.Rand, _ Size) (uint64, Shrinker[uint64]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63())) // #nosec G404 -- Using math/rand for deterministic property-based testing
		}
		v := min + uint64(r.Intn(int(max-min+1))) // #nosec G115 -- Safe for property-based testing ranges
		return unsignedShrinkInit(v, min, max)
	})
}

// ---------------- implementation / shrinking ----------------

// uint64ShrinkInit initializes the shrinking process for a uint64 value.
// It returns the initial value and a shrinker function that can generate
// progressively smaller candidates.
func uint64ShrinkInit(start, min, max uint64) (uint64, Shrinker[uint64]) {
	return unsignedShrinkInit(start, min, max)
}

// autoRangeUint64 decides the final range for Uint64(...) by combining the local "size" and the
// "size" coming from the runner. We prefer the largest range informed; if nothing is
// informed, we use [0, 100].
func autoRangeUint64(local, fromRunner Size) (uint64, uint64) {
	return autoRangeUnsigned[uint64](local, fromRunner)
}

// clampU64 constrains a uint64 value to be within the given bounds.
func clampU64(x, min, max uint64) uint64 {
	return clampUnsigned(x, min, max)
}
