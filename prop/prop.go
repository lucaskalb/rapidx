package prop

import (
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/lucaskalb/rapidx/gen"
)

type Config struct {
	Seed               int64
	Examples           int
	MaxShrink          int
	ShrinkStrat        string
	StopOnFirstFailure bool
	UseSubtests        bool
	Parallelism        int
}

var (
	flagSeed        = flag.Int64("rapidx.seed", 0, "seed global (0 => aleatório)")
	flagExamples    = flag.Int("rapidx.examples", 100, "casos por propriedade")
	flagMaxShrink   = flag.Int("rapidx.maxshrink", 400, "passos máx de shrinking")
	flagShrinkStrat = flag.String("rapidx.shrink.strategy", "bfs", "estratégia de shrinking: bfs|dfs")
	flagUseSubtests = flag.Bool("rapidx.shrink.subtests", true, "bool for enable subtests")
	flagParallelism = flag.Int("rapidx.shrink.parallel", 1, "int for enable parallelism")
)

func Default() Config {
	return Config{
		Seed:               *flagSeed,
		Examples:           *flagExamples,
		MaxShrink:          *flagMaxShrink,
		ShrinkStrat:        *flagShrinkStrat,
		StopOnFirstFailure: true,
		UseSubtests:        *flagUseSubtests,
		Parallelism:        *flagParallelism,
	}
}

func (c Config) effectiveSeed() int64 {
	if c.Seed != 0 {
		return c.Seed
	}
	return time.Now().UnixNano()
}

func ForAll[T any](t *testing.T, cfg Config, g gen.Generator[T]) func(func(*testing.T, T)) {
	return func(body func(*testing.T, T)) {
		seed := cfg.effectiveSeed()
		r := rand.New(rand.NewSource(seed))
		gen.SetShrinkStrategy(cfg.ShrinkStrat)

		t.Logf("[rapidx] seed=%d examples=%d maxshrink=%d strategy=%s",
			seed, cfg.Examples, cfg.MaxShrink, cfg.ShrinkStrat)

		for i := 0; i < cfg.Examples; i++ {
			val, shrink := g.Generate(r, gen.Size{})
			name := fmt.Sprintf("ex#%d", i+1)

			passed := t.Run(name, func(st *testing.T) { body(st, val) })
			if passed {
				continue
			}

			// ========= SHRINK =========
			min := val
			steps := 0
			acceptedPrev := true // o primeiro “min” falhou, então aceito

			for steps < cfg.MaxShrink {
				next, ok := shrink(acceptedPrev)
				if !ok {
					break
				}
				steps++
				sname := fmt.Sprintf("%s/shrink#%d", name, steps)

				stillFails := !t.Run(sname, func(st *testing.T) { body(st, next) })
				if stillFails {
					min = next
					acceptedPrev = true
				} else {
					acceptedPrev = false
				}
			}

			// regex amigável de replay
			full := fmt.Sprintf("^%s$/%s(/|$)", t.Name(), name)
			t.Fatalf("[rapidx] property failed; seed=%d; examples_run=%d; shrunk_steps=%d\n"+
				"counterexample (min): %#v\nreplay: go test -run '%s' -rapidx.seed=%d",
				seed, i+1, steps, min, full, seed)

			if cfg.StopOnFirstFailure {
				return
			}
		}
	}
}
