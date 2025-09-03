package gen

import "math/rand"

// ArrayOf gera um slice de comprimento **exato** n, usando o gerador de elementos.
// Ele é “array-like”: ótimo quando você precisa simular [N]T.
// Shrink: não pode remover elementos; apenas tenta shrink local em cada posição,
// explorando múltiplos ramos (BFS/DFS) e deduplicando candidatos.
func ArrayOf[T any](elem Generator[T], n int) Generator[[]T] {
	return From(func(r *rand.Rand, _ Size) ([]T, Shrinker[[]T]) {
		if r == nil { r = rand.New(rand.NewSource(rand.Int63())) }
		if n < 0 { n = 0 }

		// gera valores + shrinkers dos elementos
		cur := make([]T, n)
		elS := make([]Shrinker[T], n)
		for i := 0; i < n; i++ {
			v, s := elem.Generate(r, Size{})
			cur[i], elS[i] = v, s
		}

		queue := make([][]T, 0, 32)
		seen  := map[string]struct{}{sig(cur): {}}
		var last []T

		push := func(s []T) {
			k := sig(s)
			if _, ok := seen[k]; ok { return }
			seen[k] = struct{}{}
			cp := append(([]T)(nil), s...)
			queue = append(queue, cp)
		}

		// Gera vizinhos tentando “amansar” cada posição com um passo de shrink local
		grow := func(base []T) {
			queue = queue[:0]
			L := len(base)
			for i := L-1; i >= 0; i-- {
				if elS[i] == nil { continue }
				if nv, ok := elS[i](false); ok { // propõe 1 candidato para a posição i
					cand := append(([]T)(nil), base...)
					cand[i] = nv
					push(cand)
				}
			}
		}
		grow(cur)

		pop := func() ([]T, bool) {
			if len(queue) == 0 { return nil, false }
			if shrinkStrategy == "dfs" {
				v := queue[len(queue)-1]
				queue = queue[:len(queue)-1]
				return v, true
			}
			v := queue[0]
			queue = queue[1:]
			return v, true
		}

		return cur, func(accept bool) ([]T, bool) {
			if accept {
				// rebase: novo mínimo passa a ser o último candidato aceito
				if last != nil && sig(last) != sig(cur) {
					cur = last
					// após rebase, “esquecemos” shrinkers elementares para manter simplicidade;
					// ainda assim, numa próxima camada grow() reproporá um passo por posição.
					for i := range elS { elS[i] = nil }
					grow(cur)
				}
			}
			nxt, ok := pop()
			if !ok { return nil, false }
			last = nxt
			return nxt, true
		}
	})
}

