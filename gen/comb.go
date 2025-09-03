// File: gen/comb.go
package gen

import (
	"math/rand"

)

// -------------------------
// Helpers básicos
// -------------------------

// Const retorna sempre o mesmo valor (sem shrinking).
func Const[T any](v T) Generator[T] {
	return From(func(_ *rand.Rand, _ Size) (T, Shrinker[T]) {
		return v, func(bool) (T, bool) { var z T; return z, false }
	})
}

// OneOf escolhe uniformemente um dos geradores.
func OneOf[T any](gs ...Generator[T]) Generator[T] {
	return Weighted(func(_ T) float64 { return 1.0 }, gs...)
}

// Weighted escolhe um gerador com base em pesos dinâmicos (por valor).
// A estratégia aqui captura qual índice foi selecionado para poder “shrincar”
// reusando o shrinker do gerador escolhido. Opcionalmente, no shrinking
// também tenta migrar para vizinhos (outros índices) — controlado por `tryNeighbors`.
func Weighted[T any](weight func(T) float64, gs ...Generator[T]) Generator[T] {
	if len(gs) == 0 {
		panic("gen.Weighted: precisa de ao menos um gerador")
	}
	return From(func(r *rand.Rand, sz Size) (T, Shrinker[T]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		// etapa 1: escolher gerador
		idx := r.Intn(len(gs))
		val, shrink := gs[idx].Generate(r, sz)

		// fila de vizinhos (outros geradores) para tentar durante shrinking
		neighbors := make([]int, 0, len(gs)-1)
		for i := range gs {
			if i != idx {
				neighbors = append(neighbors, i)
			}
		}

		return val, func(accept bool) (T, bool) {
			// se o último candidato foi aceito (falhou), continuamos shrink do mesmo gerador
			if accept {
				if next, ok := shrink(true); ok {
					return next, true
				}
				// esgotou o shrink interno → tenta migrar para um vizinho
				for len(neighbors) > 0 {
					j := neighbors[0]
					neighbors = neighbors[1:]
					nv, ns := gs[j].Generate(r, sz)
					// atualiza “contexto” para o novo gerador
					idx, val, shrink = j, nv, ns
					return val, true
				}
				var z T
				return z, false
			}
			// candidato foi rejeitado → tente outro do mesmo shrinker
			if next, ok := shrink(false); ok {
				return next, true
			}
			// ou migre para um vizinho
			for len(neighbors) > 0 {
				j := neighbors[0]
				neighbors = neighbors[1:]
				nv, ns := gs[j].Generate(r, sz)
				idx, val, shrink = j, nv, ns
				return val, true
			}
			var z T
			return z, false
		}
	})
}

// -------------------------
// Combinadores
// -------------------------

// Map aplica f: A -> B preservando o shrinking (mapeia os candidatos de A).
func Map[A, B any](ga Generator[A], f func(A) B) Generator[B] {
	return From(func(r *rand.Rand, sz Size) (B, Shrinker[B]) {
		a, sa := ga.Generate(r, sz)
		b := f(a)
		return b, func(accept bool) (B, bool) {
			na, ok := sa(accept)
			if !ok {
				var z B
				return z, false
			}
			return f(na), true
		}
	})
}

// Filter mantém apenas valores que satisfazem pred.
// Implementa “rebase” no shrink: quando aceita, shrinka em cima do novo mínimo
// garantindo que os próximos candidatos também satisfaçam o predicado.
func Filter[T any](g Generator[T], pred func(T) bool, maxTries int) Generator[T] {
	if maxTries <= 0 {
		maxTries = 1000
	}
	return From(func(r *rand.Rand, sz Size) (T, Shrinker[T]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		// gerar um valor que passe no pred
		var v T
		var s Shrinker[T]
		okv := false
		for tries := 0; tries < maxTries; tries++ {
			v, s = g.Generate(r, sz)
			if pred(v) {
				okv = true
				break
			}
		}
		if !okv {
			var z T
			return z, func(bool) (T, bool) { return z, false }
		}

		// shrinker: sempre que aceitar, precisamos “rebasear” e continuar
		// garantindo pred nos próximos candidatos.
		return v, func(accept bool) (T, bool) {
			for {
				nv, ok := s(accept)
				if !ok {
					var z T
					return z, false
				}
				if pred(nv) {
					return nv, true
				}
				// candidato não cumpre pred → rejeita e tenta próximo
				accept = false
			}
		}
	})
}

// Bind (flatMap): o gerador de saída depende do valor gerado em A.
// Shrinking: primeiro tenta shrink em B; quando esgota, shrink em A e regenera B.
func Bind[A, B any](ga Generator[A], f func(A) Generator[B]) Generator[B] {
	return From(func(r *rand.Rand, sz Size) (B, Shrinker[B]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		a, sa := ga.Generate(r, sz)
		gb := f(a)
		b, sb := gb.Generate(r, sz)

		state := 0 // 0 => shrink B; 1 => shrink A (e regenerar B)

		return b, func(accept bool) (B, bool) {
			switch state {
			case 0:
				if nb, ok := sb(accept); ok {
					return nb, true
				}
				// esgotou shrink de B → partimos para shrink de A
				state = 1
				accept = false // primeiro passo em A é “rejeitar” para pegar próximo candidato
				fallthrough
			case 1:
				na, ok := sa(accept)
				if !ok {
					var z B
					return z, false
				}
				// regenerar B com base no novo A
				a = na
				gb = f(a)
				b, sb = gb.Generate(r, sz)
				return b, true
			default:
				var z B
				return z, false
			}
		}
	})
}

