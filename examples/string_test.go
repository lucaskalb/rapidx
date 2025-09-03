package examples

import (
	"testing"

	"github.com/lucaskalb/rapidx/gen"
	"github.com/lucaskalb/rapidx/prop"
)

func Test_String_FalsaRegra(t *testing.T) {

	prop.ForAll(t, prop.Default(), gen.StringAlphaNum(gen.Size{Min:0, Max:32}))(
		func(t *testing.T, s string) {
			if s != "" {
				t.Fatalf("esperava vazio, veio %q", s)
			}
		},
	)
}

