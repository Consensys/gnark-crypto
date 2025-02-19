package vortex

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

func TestPolyLagrangeSimple(t *testing.T) {

	t.Run("constant-base", func(t *testing.T) {

		rng := rand.New(rand.NewChaCha8([32]byte{}))

		var (
			vec = make([]koalabear.Element, 16)
			val = randElement(rng)
			x   = randFext(rng)
		)

		for i := range vec {
			vec[i] = val
		}

		y, err := EvalBasePolyLagrange(vec, x)

		if err != nil {
			t.Fatal(err)
		}

		if y != (fext.E4{B0: fext.E2{A0: val}}) {
			t.Errorf("expected %v, got %v", val.String(), y.String())
		}
	})

}
