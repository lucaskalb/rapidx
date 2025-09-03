package gen

import (
	"errors"
	"math/rand"
	"strings"
	"unicode"
)

// CPF válido; masked controla formato.
func CPF(masked bool) Generator[string] {
    return From(func(r *rand.Rand, _ Size) (string, Shrinker[string]) {
        if r == nil { r = rand.New(rand.NewSource(rand.Int63())) }

        // gera raiz 0..9 evitando todos iguais
        root := make([]byte, 9)
        for {
            for i := range 9 { root[i] = byte(r.Intn(10)) }
            if !allSameDigits(root) { break }
        }
        d1, d2 := computeCPFVerifiersBytes(root)

        raw := make([]byte, 0, 11)
        for _, n := range root { raw = append(raw, '0'+n) }
        raw = append(raw, d1, d2)

        cur := string(raw)
        if masked { cur = MaskCPF(cur) }

        // ---------- SHRINK MULTI-RAMO COM HEURÍSTICA ----------
        queue := make([]string, 0, 32)
        seen  := make(map[string]struct{}, 64) // dedup
        var last string                         // último proposto

        push := func(s string) {
            if _, ok := seen[s]; ok { return }
            seen[s] = struct{}{}
            queue = append(queue, s)
        }

        // monta vizinhos priorizando: unmask -> zero(i L->R) -> dec(j R->L)
        growNeighbors := func(base string) {
            queue = queue[:0] // reset da fila; manter 'seen' evita loops
            un := UnmaskCPF(base)

            // (1) unmask primeiro (se aplicável)
            if base != un { push(un) }

            // raiz como 0..9
            r9 := make([]byte, 9)
            for i := range 9 { r9[i] = un[i]-'0' }

            // (2) zerar dígitos L->R
            for i := range 9 {
                if r9[i] == 0 { continue }
                orig := r9[i]
                r9[i] = 0
                if !allSameDigits(r9) {
                    d1, d2 := computeCPFVerifiersBytes(r9)
                    buf := make([]byte, 0, 11)
                    for _, n := range r9 { buf = append(buf, '0'+n) }
                    buf = append(buf, d1, d2)
                    push(string(buf))
                }
                r9[i] = orig
            }

            // (3) decrementar dígitos R->L (mais “local”)
            for j := 8; j >= 0; j-- {
                if r9[j] == 0 { continue }
                r9[j]--
                if !allSameDigits(r9) {
                    d1, d2 := computeCPFVerifiersBytes(r9)
                    buf := make([]byte, 0, 11)
                    for _, n := range r9 { buf = append(buf, '0'+n) }
                    buf = append(buf, d1, d2)
                    push(string(buf))
                }
                r9[j]++
            }
        }

        // seed inicial
        seen[cur] = struct{}{}
        growNeighbors(cur)

        popNext := func() (string, bool) {
            if len(queue) == 0 { return "", false }
            if shrinkStrategy == "dfs" {
                // LIFO
                v := queue[len(queue)-1]
                queue = queue[:len(queue)-1]
                return v, true
            }
            // BFS: FIFO
            v := queue[0]
            queue = queue[1:]
            return v, true
        }

        // shrinker com feedback: accept==true -> rebase em 'last' e regen vizinhos
        return cur, func(accept bool) (string, bool) {
            if accept {
                // o último candidato manteve a falha -> vira novo mínimo
                if last != "" && last != cur {
                    cur = last
                    growNeighbors(cur)
                }
            }
            // pegar próximo vizinho para tentar
            nxt, ok := popNext()
            if !ok { return "", false }
            last = nxt
            return nxt, true
        }
    })
}


// CPFAny 50/50 com/sem máscara
func CPFAny() Generator[string] {
	return From(func(r *rand.Rand, sz Size) (string, Shrinker[string]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}
		if r.Intn(2) == 0 {
			return CPF(true).Generate(r, sz)
		}
		return CPF(false).Generate(r, sz)
	})
}

// ---------- utils domínio e helpers (iguais aos anteriores) ----------

func ValidCPF(s string) bool {
	raw := UnmaskCPF(s)
	if len(raw) != 11 {
		return false
	}
	b := []byte(raw)
	if allSame(b) {
		return false
	}
	d1, d2 := computeCPFVerifiers(b[:9])
	return b[9] == d1 && b[10] == d2
}

func MaskCPF(raw string) string {
	raw = UnmaskCPF(raw)
	if len(raw) != 11 {
		panic(errors.New("MaskCPF: precisa de 11 dígitos"))
	}
	return raw[0:3] + "." + raw[3:6] + "." + raw[6:9] + "-" + raw[9:11]
}

func UnmaskCPF(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteByte(byte((int(r) - int('0')) + int('0')))
		}
	}
	return b.String()
}

func allSame(b []byte) bool {
	if len(b) == 0 {
		return true
	}
	f := b[0]
	for _, x := range b[1:] {
		if x != f {
			return false
		}
	}
	return true
}
func allSameDigits(b []byte) bool {
	if len(b) == 0 {
		return true
	}
	f := b[0]
	for _, x := range b[1:] {
		if x != f {
			return false
		}
	}
	return true
}

func computeCPFVerifiers(root []byte) (d1, d2 byte) {
	if len(root) != 9 {
		panic(errors.New("computeCPFVerifiers: root len != 9"))
	}
	sum := 0
	for i := range 9 {
		sum += int(root[i]-'0') * (10 - i)
	}
	rest := sum % 11
	if rest < 2 {
		d1 = '0'
	} else {
		d1 = byte(11-rest) + '0'
	}

	sum = 0
	for i := range 9 {
		sum += int(root[i]-'0') * (11 - i)
	}
	sum += int(d1-'0') * 2
	rest = sum % 11
	if rest < 2 {
		d2 = '0'
	} else {
		d2 = byte(11-rest) + '0'
	}
	return
}
func computeCPFVerifiersBytes(root []byte) (d1, d2 byte) {
	if len(root) != 9 {
		panic(errors.New("computeCPFVerifiersBytes: root len != 9"))
	}
	sum := 0
	for i := range 9 {
		sum += int(root[i]) * (10 - i)
	}
	rest := sum % 11
	if rest < 2 {
		d1 = '0'
	} else {
		d1 = byte(11-rest) + '0'
	}

	sum = 0
	for i := range 9 {
		sum += int(root[i]) * (11 - i)
	}
	sum += int(d1-'0') * 2
	rest = sum % 11
	if rest < 2 {
		d2 = '0'
	} else {
		d2 = byte(11-rest) + '0'
	}
	return
}
