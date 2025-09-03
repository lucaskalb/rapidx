package gen

import (
	"fmt"
	"math/rand"
)

// SliceOf gera []T a partir de um gerador de elementos.
// - size.Min/Max controlam o comprimento (padrão Min=0, Max=16).
// Shrink:
//  (1) remover blocos grandes (metade, quarto, …) → remove indices
//  (2) remover elemento isolado (direita→esquerda)
//  (3) tentar shrink nos elementos (propagando accept)
func SliceOf[T any](elem Generator[T], size Size) Generator[[]T] {
	return From(func(r *rand.Rand, sz Size) ([]T, Shrinker[[]T]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		// defaults
		if size.Min == 0 && size.Max == 0 {
			size.Min, size.Max = 0, 16
		}
		if sz.Min != 0 || sz.Max != 0 {
			size = sz
		}
		if size.Max < size.Min {
			size.Max = size.Min
		}

		// length
		n := size.Min
		if size.Max > size.Min {
			n += r.Intn(size.Max - size.Min + 1)
		}

		// generate elems + capturar shrinkers
		vals := make([]T, n)
		shks := make([]Shrinker[T], n)
		for i := 0; i < n; i++ {
			v, s := elem.Generate(r, Size{})
			vals[i], shks[i] = v, s
		}
		cur := append(([]T)(nil), vals...) // snapshot

		// dedup por “assinatura” textual (ok para teste; evita ciclos)
		seen := map[string]struct{}{sig(cur): {}}
		queue := make([][]T, 0, 64)
		var last []T

		push := func(s []T) {
			k := sig(s)
			if _, ok := seen[k]; ok {
				return
			}
			seen[k] = struct{}{}
			// copiar para não compartilhar backing array
			cp := append(([]T)(nil), s...)
			queue = append(queue, cp)
		}

		// remove intervalos [i:j) de cur
		rem := func(base []T, i, j int) []T {
			out := make([]T, 0, len(base)-(j-i))
			out = append(out, base[:i]...)
			out = append(out, base[j:]...)
			return out
		}

		growNeighbors := func(base []T) {
			queue = queue[:0]
			L := len(base)
			if L == 0 {
				return
			}
			// (1) remover blocos grandes (binário: metade, quarto, …)
			chunk := L / 2
			for chunk >= 1 {
				for i := 0; i+chunk <= L; i += chunk {
					push(rem(base, i, i+chunk))
				}
				chunk /= 2
			}
			// (2) remover elemento isolado (R->L)
			for i := L - 1; i >= 0; i-- {
				push(rem(base, i, i+1))
			}
			// (3) shrink dos elementos localmente, mantendo tamanho
			//     (gera um vizinho por posição com 1 passo de shrink)
			for i := L - 1; i >= 0; i-- {
				if shks == nil || shks[i] == nil {
					continue
				}
				if nv, ok := shks[i](false); ok { // false: propondo candidato
					cand := append(([]T)(nil), base...)
					cand[i] = nv
					push(cand)
				}
			}
		}
		growNeighbors(cur)

		pop := func() ([]T, bool) {
			if len(queue) == 0 {
				return nil, false
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

		return cur, func(accept bool) ([]T, bool) {
			if accept {
				// rebaseia no último candidato aceito
				if last != nil && sig(last) != sig(cur) {
					cur = last
					// IMPORTANTe: quando rebaseamos, precisamos regenerar shrinkers
					// para manter consistência dos elementos (pode ser caro, mas simples)
					shks = make([]Shrinker[T], len(cur))
					for i := range cur {
						// reconstroi shrinker “focal” partindo do valor atual:
						v := cur[i]
						// truque: crie um gerador Const(v) e peça o shrinker dele (não tem).
						// então, melhor: se queremos shrink futuro nos elementos, precisamos
						// aceitar que só teremos 1 passo no vizinho (já feito em growNeighbors).
						// Para manter simples no MVP, não retemos shrinkers após rebase.
						shks[i] = nil
						_ = v
					}
					growNeighbors(cur)
				}
			}
			nxt, ok := pop()
			if !ok {
				return nil, false
			}
			last = nxt
			return nxt, true
		}
	})
}

// sig cria uma assinatura textual simplificada de um slice genérico.
// Para fins de dedup de shrinking em testes, isso é suficiente.
func sig[T any](s []T) string { return fmt.Sprintf("%#v", s) }

