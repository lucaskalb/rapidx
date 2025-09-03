// File: gen/int.go
package gen

import (
	"math/rand"
)

// Int gera inteiros com faixa automática a partir de Size:
// - se sz.Max (ou |sz.Min|) > 0: faixa := [-M, M], onde M = max(|sz.Min|, |sz.Max|)
// - caso contrário, usa faixa padrão [-100, 100].
// Ex.: prop.ForAll(t, cfg, gen.Int(gen.Size{Max: 1000})) ...
func Int(size Size) Generator[int] {
	return From(func(r *rand.Rand, sz Size) (int, Shrinker[int]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		min, max := autoRange(size, sz) // decide a faixa efetiva
		if min > max {
			min, max = max, min
		}
		// gera uniforme
		v := min + r.Intn(max-min+1)
		return intShrinkInit(v, min, max)
	})
}

// IntRange gera inteiros uniformemente no intervalo [min, max] (inclusivo).
// Ignora sz para a faixa (útil quando você quer controle explícito).
func IntRange(min, max int) Generator[int] {
	if min > max {
		min, max = max, min
	}
	return From(func(r *rand.Rand, _ Size) (int, Shrinker[int]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		v := min + r.Intn(max-min+1)
		return intShrinkInit(v, min, max)
	})
}

// -------------------- implementação / shrinking --------------------

func intShrinkInit(start, min, max int) (int, Shrinker[int]) {
	// valor corrente (mínimo conhecido que falha) e último proposto
	cur := clamp(start, min, max)
	last := cur

	// fila de vizinhos + deduplicação
	queue := make([]int, 0, 16)
	seen := map[int]struct{}{cur: {}}

	push := func(x int) {
		if x < min || x > max {
			return
		}
		if _, ok := seen[x]; ok {
			return
		}
		seen[x] = struct{}{}
		queue = append(queue, x)
	}

	// heurística de vizinhos:
	//  1) aproximar do alvo (0 se estiver na faixa, senão limite mais próximo)
	//  2) “meio do caminho” em direção ao alvo (bisseção)
	//  3) passo unitário em direção ao alvo (+/-1)
	//  4) limites (min/max)
	growNeighbors := func(base int) {
		queue = queue[:0]
		target := shrinkTarget(min, max) // 0 se possível; senão bound mais próximo

		// (1) alvo direto
		if base != target {
			push(target)
		}

		// (2) meio do caminho em direção ao alvo (bisseção)
		if base != target {
			next := midpointTowards(base, target)
			if next != base {
				push(next)
			}
			// múltiplas bisseções arredondando para longe de base
			// (gera série base -> base' -> ... -> target)
			series := next
			for i := 0; i < 8; i++ {
				if series == target {
					break
				}
				series = midpointTowards(series, target)
				if series != base {
					push(series)
				}
			}
		}

		// (3) passo unitário em direção ao alvo
		if base != target {
			step := stepTowards(base, target)
			if step != base {
				push(step)
			}
		}

		// (4) limites
		if base != min {
			push(min)
		}
		if base != max {
			push(max)
		}
	}

	growNeighbors(cur)

	pop := func() (int, bool) {
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

	return cur, func(accept bool) (int, bool) {
		// Se o último candidato foi ACEITO (continua falhando), rebaseie nele
		if accept {
			if last != cur {
				cur = last
				growNeighbors(cur)
			}
		}
		// proponha o próximo vizinho
		nxt, ok := pop()
		if !ok {
			return 0, false
		}
		last = nxt
		return nxt, true
	}
}

// shrinkTarget retorna o alvo “natural” para onde reduzir:
// - 0 se 0 ∈ [min,max]; caso contrário, o limite mais próximo de 0.
func shrinkTarget(min, max int) int {
	if min <= 0 && 0 <= max {
		return 0
	}
	// fora da faixa: pegue o bound mais próximo de 0
	if min > 0 {
		// faixa toda positiva -> min é o mais próximo de 0
		return min
	}
	// faixa toda negativa -> max é o mais próximo de 0 (ex.: [-10, -1] → -1)
	return max
}

// midpointTowards dá um “passo de bisseção” de a em direção a b,
// com arredondamento para longe de 'a' para garantir progresso.
func midpointTowards(a, b int) int {
	if a == b {
		return a
	}
	d := b - a
	// arredonda “para cima” em magnitude para não travar quando |d| == 1
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

// stepTowards move um passo unitário de a em direção a b.
func stepTowards(a, b int) int {
	if a == b {
		return a
	}
	if b > a {
		return a + 1
	}
	return a - 1
}

// autoRange decide a faixa final para Int(...) combinando o "size" local e o
// "size" vindo do runner. Preferimos o maior alcance informado; se nada for
// informado, usamos [-100, 100].
func autoRange(local, fromRunner Size) (int, int) {
	// escolha um "M" (magnitude) baseado no maior valor absoluto visto
	M := 0
	for _, s := range []Size{local, fromRunner} {
		M = maxInt(M, absInt(s.Min))
		M = maxInt(M, absInt(s.Max))
	}
	if M == 0 {
		M = 100
	}
	return -M, M
}

func clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

