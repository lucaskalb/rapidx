package gen

import (
	"math/rand"
)

// Int64 gera inteiros 64-bit com faixa automática baseada em Size.
// Se nenhum Size for informado, usa [-100, 100].
func Int64(size Size) Generator[int64] {
	return From(func(r *rand.Rand, sz Size) (int64, Shrinker[int64]) {
		if r == nil { r = rand.New(rand.NewSource(rand.Int63())) }
		min, max := autoRange64(size, sz)
		if min > max { min, max = max, min }
		v := min + int64(r.Intn(int(max-min+1)))
		return int64ShrinkInit(v, min, max)
	})
}

// Int64Range gera int64 uniformemente no intervalo [min, max] (inclusivo).
func Int64Range(min, max int64) Generator[int64] {
	if min > max { min, max = max, min }
	return From(func(r *rand.Rand, _ Size) (int64, Shrinker[int64]) {
		if r == nil { r = rand.New(rand.NewSource(rand.Int63())) }
		v := min + int64(r.Intn(int(max-min+1)))
		return int64ShrinkInit(v, min, max)
	})
}

// ---------------- impl / shrinking ----------------

func int64ShrinkInit(start, min, max int64) (int64, Shrinker[int64]) {
	cur, last := clamp64(start, min, max), clamp64(start, min, max)

	queue := make([]int64, 0, 16)
	seen  := map[int64]struct{}{cur: {}}

	push := func(x int64) {
		if x < min || x > max { return }
		if _, ok := seen[x]; ok { return }
		seen[x] = struct{}{}
		queue = append(queue, x)
	}
	target := shrinkTarget64(min, max)

	grow := func(base int64) {
		queue = queue[:0]
		// (1) alvo (0 se dentro da faixa; senão bound + próximo)
		if base != target { push(target) }
		// (2) bisseções rumo ao alvo
		if base != target {
			next := midpointTowards64(base, target)
			if next != base { push(next) }
			series := next
			for i := 0; i < 8 && series != target; i++ {
				series = midpointTowards64(series, target)
				if series != base { push(series) }
			}
		}
		// (3) passo unitário
		if base != target { push(stepTowards64(base, target)) }
		// (4) limites
		if base != min { push(min) }
		if base != max { push(max) }
	}
	grow(cur)

	pop := func() (int64, bool) {
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

	return cur, func(accept bool) (int64, bool) {
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

func shrinkTarget64(min, max int64) int64 {
	if min <= 0 && 0 <= max { return 0 }
	if min > 0 { return min }
	return max
}
func clamp64(x, min, max int64) int64 {
	if x < min { return min }
	if x > max { return max }
	return x
}
func midpointTowards64(a, b int64) int64 {
	if a == b { return a }
	d := b - a
	step := d / 2
	if step == 0 { if d > 0 { step = 1 } else { step = -1 } }
	return a + step
}
func stepTowards64(a, b int64) int64 {
	if a == b { return a }
	if b > a { return a + 1 }
	return a - 1
}
func autoRange64(local, fromRunner Size) (int64, int64) {
	M := int64(0)
	for _, s := range []Size{local, fromRunner} {
		if abs := int64Abs(s.Min); abs > M { M = abs }
		if abs := int64Abs(s.Max); abs > M { M = abs }
	}
	if M == 0 { M = 100 }
	return -M, M
}
func int64Abs(x int) int64 {
	if x < 0 { return int64(-x) }
	return int64(x)
}

