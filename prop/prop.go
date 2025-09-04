package prop

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
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
	flagSeed        = flag.Int64("rapidx.seed", 0, "")
	flagExamples    = flag.Int("rapidx.examples", 100, "")
	flagMaxShrink   = flag.Int("rapidx.maxshrink", 400, "")
	flagShrinkStrat = flag.String("rapidx.shrink.strategy", "bfs", "")
	flagUseSubtests = flag.Bool("rapidx.shrink.subtests", true, "")
	flagParallelism = flag.Int("rapidx.shrink.parallel", 1, "")
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

		t.Logf("[rapidx] seed=%d examples=%d maxshrink=%d strategy=%s parallelism=%d",
			seed, cfg.Examples, cfg.MaxShrink, cfg.ShrinkStrat, cfg.Parallelism)

		if cfg.Parallelism <= 1 {
			runSequential(t, cfg, g, body, seed, r)
		} else {
			runParallel(t, cfg, g, body, seed, r)
		}
	}
}

func runSequential[T any](t *testing.T, cfg Config, g gen.Generator[T], body func(*testing.T, T), seed int64, r *rand.Rand) {
	for i := 0; i < cfg.Examples; i++ {
		val, shrink := g.Generate(r, gen.Size{})
		name := fmt.Sprintf("ex#%d", i+1)

		passed := t.Run(name, func(st *testing.T) { body(st, val) })
		if passed {
			continue
		}


		min := val
		steps := 0
		acceptedPrev := true

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


		full := fmt.Sprintf("^%s$/%s(/|$)", t.Name(), name)
		t.Fatalf("[rapidx] property failed; seed=%d; examples_run=%d; shrunk_steps=%d\n"+
			"counterexample (min): %#v\nreplay: go test -run '%s' -rapidx.seed=%d",
			seed, i+1, steps, min, full, seed)

		if cfg.StopOnFirstFailure {
			return
		}
	}
}

func runParallel[T any](t *testing.T, cfg Config, g gen.Generator[T], body func(*testing.T, T), seed int64, r *rand.Rand) {

	testChan := make(chan int, cfg.Examples)
	

	for i := 0; i < cfg.Examples; i++ {
		testChan <- i
	}
	close(testChan)


	var wg sync.WaitGroup
	

	var randMutex sync.Mutex
	

	failureChan := make(chan failureResult, cfg.Examples)
	

	for i := 0; i < cfg.Parallelism; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for testIndex := range testChan {

				randMutex.Lock()
				val, shrink := g.Generate(r, gen.Size{})
				randMutex.Unlock()
				
				name := fmt.Sprintf("ex#%d", testIndex+1)

				passed := t.Run(name, func(st *testing.T) { body(st, val) })
				if passed {
					continue
				}

		
				min := val
				steps := 0
				acceptedPrev := true

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


				failureChan <- failureResult{
					testIndex: testIndex,
					name:      name,
					min:       min,
					steps:     steps,
				}
				
				if cfg.StopOnFirstFailure {
					return
				}
			}
		}(i)
	}


	go func() {
		wg.Wait()
		close(failureChan)
	}()


	for failure := range failureChan {
		full := fmt.Sprintf("^%s$/%s(/|$)", t.Name(), failure.name)
		t.Fatalf("[rapidx] property failed; seed=%d; examples_run=%d; shrunk_steps=%d\n"+
			"counterexample (min): %#v\nreplay: go test -run '%s' -rapidx.seed=%d",
			seed, failure.testIndex+1, failure.steps, failure.min, full, seed)
		
		if cfg.StopOnFirstFailure {
			return
		}
	}
}

type failureResult struct {
	testIndex int
	name      string
	min       interface{}
	steps     int
}
