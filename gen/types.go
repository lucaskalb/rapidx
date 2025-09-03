package gen

import "math/rand"

// Size controla escala/limites dos geradores.
type Size struct{ Min, Max int }

// Shrinker propõe candidatos “menores”.
// Parâmetro accept: true se o candidato ANTERIOR foi aceito (isto é, reproduziu a falha).
// Isso permite ao shrinker “rebasear” e gerar novos vizinhos a partir do novo mínimo.
type Shrinker[T any] func(accept bool) (next T, ok bool)

// Generator é o contrato público.
type Generator[T any] interface {
	Generate(r *rand.Rand, sz Size) (value T, shrink Shrinker[T])
}

var shrinkStrategy = "bfs"
func SetShrinkStrategy(s string) {
	if s == "dfs" {
		shrinkStrategy = "dfs"
	} else {
		shrinkStrategy = "bfs"
	}
}

// Alias (opcional) p/ compat.
type T[T any] = Generator[T]

// Helper para criar um Generator a partir de uma closure.
type GenFunc[T any] struct {
	fn func(r *rand.Rand, sz Size) (T, Shrinker[T])
}

func (g GenFunc[T]) Generate(r *rand.Rand, sz Size) (T, Shrinker[T]) { return g.fn(r, sz) }

func From[T any](fn func(*rand.Rand, Size) (T, Shrinker[T])) Generator[T] {
	return GenFunc[T]{fn: fn}
}
