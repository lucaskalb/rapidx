package gen

import "math/rand"

// Uint64 gera inteiros sem sinal 64-bit com faixa automática baseada em Size.
// Se nada for informado, usa [0, 100].
func Uint64(size Size) Generator[uint64] {
	return From(func(r *rand.Rand, sz Size) (uint64, Shrinker[uint64]) {
		if r == nil { r = rand.New(rand.NewSource(rand.Int63())) }
		min, max := autoRangeUint64(size, sz)
		if min > max { min, max = max, min }
		v := min + uint64(r.Intn(int(max-min+1)))
		return uint64ShrinkInit(v, min, max)
	})
}

// Uint64Range gera uint64 uniformemente no intervalo [min, max] (inclusivo).
func Uint64Range(min, max uint64) Generator[uint64] {
	if min > max { min, max = max, min }
	return From(func(r *rand.Rand, _ Size) (uint64, Shrinker[uint64]) {
		if r == nil { r = rand.New(rand.NewSource(rand.Int63())) }
		v := min + uint64(r.Intn(int(max-min+1)))
		return uint64ShrinkInit(v, min, max)
	})
}

// ---------------- impl / shrinking ----------------

func uint64ShrinkInit(start, min, max uint64) (uint64, Shrinker[uint64]) {
	cur, last := clampU64(start, min, max), clampU64(start, min, max)

	queue := make([]uint64, 0, 16)
	seen  := map[uint64]struct{}{cur: {}}

	push := func(x uint64) {
		if x < min || x > max { return }
		if _, ok := seen[x]; ok { return }
		seen[x] = struct{}{}
		queue = append(queue, x)
	}

	grow := func(base uint64) {
		queue = queue[:0]
		// (1) alvo natural p/ uint64 é 0
		if base != 0 { push(0) }
		// (2) bisseções rumo a 0
		if base != 0 {
			next := base / 2
			if next != base { push(next) }
			series := next
			for i := 0; i < 8 && series > 0; i++ {
				series /= 2
				push(series)
			}
		}
		// (3) passo unitário
		if base > 0 { push(base - 1) }
		// (4) limites
		if base != min { push(min) }
		if base != max { push(max) }
	}
	grow(cur)

	pop := func() (uint64, bool) {
		if len(queue) == 0 { return 0, false }
		if shrinkStrategy == "dfs" {
			v := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			return v, true
		}
		v := queue[0]
		queue = queue[1:]
		return v, true
	}

	return cur, func(accept bool) (uint64, bool) {
		if accept && last != cur {
			cur = last
			grow(cur)
		}
		nxt, ok := pop()
		if !ok { return 0, false }
		last = nxt
		return nxt, true
	}
}

func autoRangeUint64(local, fromRunner Size) (uint64, uint64) {
	M := 0
	for _, s := range []Size{local, fromRunner} {
		if s.Max > M { M = s.Max }
	}
	if M == 0 { M = 100 }
	return 0, uint64(M)
}
func clampU64(x, min, max uint64) uint64 {
	if x < min { return min }
	if x > max { return max }
	return x
}

