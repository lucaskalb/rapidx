package gen

import (
	"math/rand"
	"unicode/utf8"
)

// Atalhos de alfabetos comuns (ASCII puro pra evitar surpresas)
const (
	AlphabetLower   = "abcdefghijklmnopqrstuvwxyz"
	AlphabetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AlphabetAlpha   = AlphabetLower + AlphabetUpper
	AlphabetDigits  = "0123456789"
	AlphabetAlphaNum = AlphabetAlpha + AlphabetDigits
	AlphabetASCII   = AlphabetAlphaNum + " !\"#$%&'()*+,-./:;<=>?@[\\]^_{|}~"
)

// String gera strings usando um alfabeto (conjunto de runas) e um Size.
// - Se size.Min/Max = 0, usa padrão: Min=0, Max=32.
// - Se alphabet vazio, usa AlphabetAlphaNum.
func String(alphabet string, size Size) Generator[string] {
	return From(func(r *rand.Rand, sz Size) (string, Shrinker[string]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		// defaults
		if len(alphabet) == 0 {
			alphabet = AlphabetAlphaNum
		}
		if size.Min == 0 && size.Max == 0 {
			size.Min, size.Max = 0, 32
		}
		if sz.Min != 0 || sz.Max != 0 { // permitir override externo
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

		// ---- shrinking: multi-ramo (BFS/DFS) com dedup ----
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

		// heurística:
		// (1) encurtar (remover sufixo)
		// (2) substituir caracteres por “mais simples” (primeiro da tabela; ex.: 'a' ou '0')
		growNeighbors := func(base string) {
			queue = queue[:0]
			// (1) encurtar vários passos de uma vez (gerar vários comprimentos)
			if len(base) > 0 {
				for newLen := len(base) - 1; newLen >= 0; newLen-- {
					push(base[:newLen])
				}
			}
			// (2) amansar caracteres para o primeiro do alfabeto
			if len(base) > 0 {
				target := rune(alphabet[0]) // ex.: 'a' ou '0'
				rs := []rune(base)
				// direita→esquerda para estabilizar logo sufixos
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
			if shrinkStrategy == "dfs" {
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

// Açúcares sintáticos
func StringAlpha(size Size) Generator[string]    { return String(AlphabetAlpha, size) }
func StringAlphaNum(size Size) Generator[string] { return String(AlphabetAlphaNum, size) }
func StringDigits(size Size) Generator[string]   { return String(AlphabetDigits, size) }
func StringASCII(size Size) Generator[string]    { return String(AlphabetASCII, size) }

