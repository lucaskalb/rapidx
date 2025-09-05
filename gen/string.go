package gen

import (
	"math/rand"
	"unicode/utf8"
)

// Common alphabet shortcuts (pure ASCII to avoid surprises)
const (
	// AlphabetLower contains lowercase letters a-z.
	AlphabetLower = "abcdefghijklmnopqrstuvwxyz"

	// AlphabetUpper contains uppercase letters A-Z.
	AlphabetUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// AlphabetAlpha contains both lowercase and uppercase letters.
	AlphabetAlpha = AlphabetLower + AlphabetUpper

	// AlphabetDigits contains digits 0-9.
	AlphabetDigits = "0123456789"

	// AlphabetAlphaNum contains letters and digits.
	AlphabetAlphaNum = AlphabetAlpha + AlphabetDigits

	// AlphabetASCII contains all printable ASCII characters.
	AlphabetASCII = AlphabetAlphaNum + " !\"#$%&'()*+,-./:;<=>?@[\\]^_{|}~"
)

// String generates strings using an alphabet (set of runes) and a Size.
// - If size.Min/Max = 0, uses default: Min=0, Max=32.
// - If alphabet is empty, uses AlphabetAlphaNum.
func String(alphabet string, size Size) Generator[string] {
	return From(func(r *rand.Rand, sz Size) (string, Shrinker[string]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63())) // #nosec G404 -- Using math/rand for deterministic property-based testing
		}
		// defaults
		if len(alphabet) == 0 {
			alphabet = AlphabetAlphaNum
		}
		if size.Min == 0 && size.Max == 0 {
			size.Min, size.Max = 0, 32
		}
		if sz.Min != 0 || sz.Max != 0 { // allow external override
			size = sz
		}
		if size.Max < size.Min {
			size.Max = size.Min
		}

		// generate
		n := size.Min
		if size.Max > size.Min {
			n += r.Intn(size.Max - size.Min + 1)
		}
		b := make([]rune, n)
		for i := 0; i < n; i++ {
			b[i] = rune(alphabet[r.Intn(len(alphabet))])
		}
		cur := string(b)

		// ---- shrinking: multi-branch (BFS/DFS) with dedup ----
		type neighbor = string
		queue := make([]neighbor, 0, 64)
		seen := map[string]struct{}{cur: {}}
		var last string

		push := func(s string) {
			if _, ok := seen[s]; ok {
				return
			}
			seen[s] = struct{}{}
			queue = append(queue, s)
		}

		// heuristic:
		// (1) shorten (remove suffix)
		// (2) replace characters with "simpler" ones (first in table; e.g., 'a' or '0')
		growNeighbors := func(base string) {
			queue = queue[:0]
			// (1) shorten multiple steps at once (generate multiple lengths)
			if len(base) > 0 {
				for newLen := len(base) - 1; newLen >= 0; newLen-- {
					push(base[:newLen])
				}
			}
			// (2) tame characters to the first in the alphabet
			if len(base) > 0 {
				target := rune(alphabet[0]) // e.g., 'a' or '0'
				rs := []rune(base)
				// right→left to quickly stabilize suffixes
				for i := len(rs) - 1; i >= 0; i-- {
					if rs[i] != target {
						rs2 := make([]rune, len(rs))
						copy(rs2, rs)
						rs2[i] = target
						if s := string(rs2); utf8.ValidString(s) {
							push(s)
						}
					}
				}
			}
		}
		growNeighbors(cur)

		pop := func() (string, bool) {
			if len(queue) == 0 {
				return "", false
			}
			if shrinkStrategy == ShrinkStrategyDFS {
				v := queue[len(queue)-1]
				queue = queue[:len(queue)-1]
				return v, true
			}
			v := queue[0]
			queue = queue[1:]
			return v, true
		}

		return cur, func(accept bool) (string, bool) {
			if accept {
				if last != "" && last != cur {
					cur = last
					growNeighbors(cur)
				}
			}
			next, ok := pop()
			if !ok {
				return "", false
			}
			last = next
			return next, true
		}
	})
}

// Syntactic sugar functions for common string generators
// StringAlpha generates strings using only alphabetic characters.
func StringAlpha(size Size) Generator[string] { return String(AlphabetAlpha, size) }

// StringAlphaNum generates strings using alphanumeric characters.
func StringAlphaNum(size Size) Generator[string] { return String(AlphabetAlphaNum, size) }

// StringDigits generates strings using only digits.
func StringDigits(size Size) Generator[string] { return String(AlphabetDigits, size) }

// StringASCII generates strings using all printable ASCII characters.
func StringASCII(size Size) Generator[string] { return String(AlphabetASCII, size) }
