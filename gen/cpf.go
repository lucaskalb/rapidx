package gen

import (
	"errors"
	"math/rand"
	"strings"
	"unicode"
)

// CPF generates valid CPF numbers; masked controls the format.
func CPF(masked bool) Generator[string] {
	return From(func(r *rand.Rand, _ Size) (string, Shrinker[string]) {
		if r == nil {
			r = rand.New(rand.NewSource(rand.Int63()))
		}

		// generate root 0..9 avoiding all same digits
		root := make([]byte, 9)
		for {
			for i := range 9 {
				root[i] = byte(r.Intn(10))
			}
			if !allSameDigits(root) {
				break
			}
		}
		d1, d2 := computeCPFVerifiersBytes(root)

		raw := make([]byte, 0, 11)
		for _, n := range root {
			raw = append(raw, '0'+n)
		}
		raw = append(raw, d1, d2)

		cur := string(raw)
		if masked {
			cur = MaskCPF(cur)
		}

		// ---------- MULTI-BRANCH SHRINK WITH HEURISTIC ----------
		queue := make([]string, 0, 32)
		seen := make(map[string]struct{}, 64) // dedup
		var last string                       // last proposed

		push := func(s string) {
			if _, ok := seen[s]; ok {
				return
			}
			seen[s] = struct{}{}
			queue = append(queue, s)
		}

		// build neighbors prioritizing: unmask -> zero(i L->R) -> dec(j R->L)
		growNeighbors := func(base string) {
			queue = queue[:0] // reset queue; keeping 'seen' avoids loops
			un := UnmaskCPF(base)

			// (1) unmask first (if applicable)
			if base != un {
				push(un)
			}

			// root as 0..9
			r9 := make([]byte, 9)
			for i := range 9 {
				r9[i] = un[i] - '0'
			}

			// (2) zero digits L->R
			for i := range 9 {
				if r9[i] == 0 {
					continue
				}
				orig := r9[i]
				r9[i] = 0
				if !allSameDigits(r9) {
					d1, d2 := computeCPFVerifiersBytes(r9)
					buf := make([]byte, 0, 11)
					for _, n := range r9 {
						buf = append(buf, '0'+n)
					}
					buf = append(buf, d1, d2)
					push(string(buf))
				}
				r9[i] = orig
			}

			// (3) decrement digits R->L (more "local")
			for j := 8; j >= 0; j-- {
				if r9[j] == 0 {
					continue
				}
				r9[j]--
				if !allSameDigits(r9) {
					d1, d2 := computeCPFVerifiersBytes(r9)
					buf := make([]byte, 0, 11)
					for _, n := range r9 {
						buf = append(buf, '0'+n)
					}
					buf = append(buf, d1, d2)
					push(string(buf))
				}
				r9[j]++
			}
		}

		// initial seed
		seen[cur] = struct{}{}
		growNeighbors(cur)

		popNext := func() (string, bool) {
			if len(queue) == 0 {
				return "", false
			}
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

		// shrinker with feedback: accept==true -> rebase on 'last' and regen neighbors
		return cur, func(accept bool) (string, bool) {
			if accept {
				// the last candidate maintained the failure -> becomes new minimum
				if last != "" && last != cur {
					cur = last
					growNeighbors(cur)
				}
			}
			// get next neighbor to try
			nxt, ok := popNext()
			if !ok {
				return "", false
			}
			last = nxt
			return nxt, true
		}
	})
}

// CPFAny generates CPF numbers with 50/50 chance of being masked or unmasked.
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

// ---------- domain utils and helpers (same as before) ----------

// ValidCPF checks if a string is a valid CPF number.
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

// MaskCPF formats a raw CPF string with dots and dashes.
func MaskCPF(raw string) string {
	raw = UnmaskCPF(raw)
	if len(raw) != 11 {
		panic(errors.New("MaskCPF: needs 11 digits"))
	}
	return raw[0:3] + "." + raw[3:6] + "." + raw[6:9] + "-" + raw[9:11]
}

// UnmaskCPF removes all non-digit characters from a CPF string.
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

// allSame checks if all bytes in a slice are the same.
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

// allSameDigits checks if all bytes in a slice represent the same digit.
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

// computeCPFVerifiers calculates the verification digits for a CPF root.
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

// computeCPFVerifiersBytes calculates the verification digits for a CPF root (byte version).
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
