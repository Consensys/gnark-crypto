package bls381

import (
	"math/big"
	"testing"

	bls12381 "github.com/kilic/bls12-381"

	"github.com/consensys/gurvy/bls381/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// Test against github.com/kilic/bls12-381

func TestG1AffineSerializationInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] G1: gurvy -> bls12-381 -> gurvy should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G1Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g1GenAff, &ab)

			g1 := bls12381.NewG1()
			other, err := g1.FromBytes(start.Marshal())
			if err != nil {
				t.Log(err)
				return false
			}
			// reconstruct a gurvy point from  bytes
			err = end.Unmarshal(g1.ToBytes(other))
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		GenFp(),
	))

	properties.Property("[BLS381] G1 compressed: gurvy -> bls12-381 -> gurvy should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G1Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g1GenAff, &ab)

			g1 := bls12381.NewG1()
			b := start.Bytes()
			other, err := g1.FromCompressed(b[:])
			if err != nil {
				t.Log(err)
				return false
			}
			// reconstruct a gurvy point from  bytes
			err = end.Unmarshal(g1.ToCompressed(other))
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

func TestG2AffineSerializationInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] G2: gurvy -> bls12-381 -> gurvy should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G2Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g2GenAff, &ab)

			g2 := bls12381.NewG2()
			other, err := g2.FromBytes(start.Marshal())
			if err != nil {
				t.Log(err)
				return false
			}
			// reconstruct a gurvy point from  bytes
			err = end.Unmarshal(g2.ToBytes(other))
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		GenFp(),
	))

	properties.Property("[BLS381] G2 compressed: gurvy -> bls12-381 -> gurvy should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G2Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g2GenAff, &ab)

			g2 := bls12381.NewG2()
			b := start.Bytes()
			other, err := g2.FromCompressed(b[:])
			if err != nil {
				t.Log(err)
				return false
			}
			// reconstruct a gurvy point from  bytes
			err = end.Unmarshal(g2.ToCompressed(other))
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}
