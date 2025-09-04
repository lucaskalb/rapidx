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
		min, max := autoRangeUnsigned[uint](size, sz) // [min,max], min>=0
		if min > max {
			min, max = max, min
		}
		v := min + uint(r.Intn(int(max-min+1)))
		return unsignedShrinkInit(v, min, max)
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
		return unsignedShrinkInit(v, min, max)
	})
}

// ---------------- implementation / shrinking ----------------

// uintShrinkInit initializes the shrinking process for a uint value.
// It returns the initial value and a shrinker function that can generate
// progressively smaller candidates.
func uintShrinkInit(start, min, max uint) (uint, Shrinker[uint]) {
	return unsignedShrinkInit(start, min, max)
}
