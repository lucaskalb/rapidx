package gen

// unsignedShrinkInit is a generic implementation for unsigned integer shrinking.
// It works with any unsigned integer type that supports the required operations.
func unsignedShrinkInit[T ~uint | ~uint64](start, min, max T) (T, Shrinker[T]) {
	cur, last := clampUnsigned(start, min, max), clampUnsigned(start, min, max)

	queue := make([]T, 0, 16)
	seen := map[T]struct{}{cur: {}}

	push := func(x T) {
		if x < min || x > max {
			return
		}
		if _, ok := seen[x]; ok {
			return
		}
		seen[x] = struct{}{}
		queue = append(queue, x)
	}

	grow := func(base T) {
		queue = queue[:0]
		// (1) natural target for unsigned integers is 0
		if base != 0 {
			push(0)
		}
		// (2) bisections towards 0
		if base != 0 {
			next := base / 2
			if next != base {
				push(next)
			}
			series := next
			for i := 0; i < 8 && series > 0; i++ {
				series /= 2
				push(series)
			}
		}
		// (3) unit step towards 0
		if base > 0 {
			push(base - 1)
		}
		// (4) bounds
		if base != min {
			push(min)
		}
		if base != max {
			push(max)
		}
	}
	grow(cur)

	pop := func() (T, bool) {
		if len(queue) == 0 {
			return 0, false
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

	return cur, func(accept bool) (T, bool) {
		if accept && last != cur {
			cur = last
			grow(cur)
		}
		nxt, ok := pop()
		if !ok {
			return 0, false
		}
		last = nxt
		return nxt, true
	}
}

// clampUnsigned constrains an unsigned integer value to be within the given bounds.
func clampUnsigned[T ~uint | ~uint64](x, min, max T) T {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// autoRangeUnsigned decides the final range for unsigned integers by combining the local "size" and the
// "size" coming from the runner. We prefer the largest range informed; if nothing is
// informed, we use [0, 100].
func autoRangeUnsigned[T ~uint | ~uint64](local, fromRunner Size) (T, T) {
	M := 0
	for _, s := range []Size{local, fromRunner} {
		if s.Max > M {
			M = s.Max
		}
	}
	if M == 0 {
		M = 100
	}
	return 0, T(M)
}
