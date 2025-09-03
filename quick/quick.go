package quick

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Equal[T any](t *testing.T, got, want T) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}

