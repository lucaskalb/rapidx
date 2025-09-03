package examples

import (
	"math/rand"
	"testing"

	"github.com/lucaskalb/rapidx/gen"
	"github.com/lucaskalb/rapidx/prop"
)

func Test_Slice_SomaNaoNegativa(t *testing.T) {
	// Propriedade falsa: “soma do slice é sempre 0”
	ints := gen.From(func(r *rand.Rand, _ gen.Size) (int, gen.Shrinker[int]) {
		if r == nil { r = rand.New(rand.NewSource(rand.Int63())) }
		v := r.Intn(201) - 100 // [-100..100]
		// shrink simples p/ int: caminhar a 0
		cur := v
		return v, func(accept bool) (int, bool) {
			if cur == 0 { return 0, false }
			// aproxima metade em direção a 0
			if cur > 0 { cur = cur / 2 } else { cur = cur / 2 }
			if cur == 0 && v != 0 { // garante ao menos 1 passo a 0
				cur = 0
			}
			return cur, true
		}
	})

	prop.ForAll(t, prop.Default(), gen.SliceOf(ints, gen.Size{Min:0, Max:16}))(
		func(t *testing.T, xs []int) {
			sum := 0
			for _, x := range xs { sum += x }
			if sum != 0 {
				t.Fatalf("esperava soma=0; xs=%v sum=%d", xs, sum)
			}
		},
	)
}

