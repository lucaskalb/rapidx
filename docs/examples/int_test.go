// Package examples demonstrates how to use the rapidx property-based testing library.
// These examples show various testing patterns and how the shrinking mechanism
// helps find minimal counterexamples when properties fail.
package examples

import (
	"math/rand"
	"testing"

	"github.com/lucaskalb/rapidx/gen"
	"github.com/lucaskalb/rapidx/prop"
)

// Test_Slice_SomaNaoNegativa demonstrates a property-based test with a custom generator
// that is designed to fail. This test verifies a false property: "the sum of a slice is always 0".
// The custom integer generator creates values in the range [-100, 100] with a simple
// shrinking strategy that approaches 0. This example shows how to create custom generators
// and how the shrinking mechanism will find a minimal counterexample when the property fails.
func Test_Slice_SomaNaoNegativa(t *testing.T) {
	// False property: "slice sum is always 0"
	ints := gen.From(func(r *rand.Rand, _ gen.Size) (int, gen.Shrinker[int]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		v := r.Intn(201) - 100 // [-100..100]
		// simple shrink for int: walk towards 0
		cur := v
		return v, func(accept bool) (int, bool) {
			if cur == 0 {
				return 0, false
			}
			// approach half way towards 0
			if cur > 0 {
				cur = cur / 2
			} else {
				cur = cur / 2
			}
			if cur == 0 && v != 0 { // ensure at least 1 step to 0
				cur = 0
			}
			return cur, true
		}
	})

	prop.ForAll(t, prop.Default(), gen.SliceOf(ints, gen.Size{Min: 0, Max: 16}))(
		func(t *testing.T, xs []int) {
			sum := 0
			for _, x := range xs {
				sum += x
			}
			if sum != 0 {
				t.Fatalf("expected sum=0; xs=%v sum=%d", xs, sum)
			}
		},
	)
}
