package vortex

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/stretchr/testify/require"
)

func TestPolyLagrangeSimple(t *testing.T) {

	t.Run("constant-base", func(t *testing.T) {
		assert := require.New(t)

		// #nosec #G404 -- test case generation does not require a cryptographic PRNG
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
		assert.NoError(err)

		assert.Equal(y, fext.E4{B0: fext.E2{A0: val}})
	})

}
