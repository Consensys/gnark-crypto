package bls381

import (
	"bytes"
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

func TestGTSerializationInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 5

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] GT: gurvy -> bls12-381 -> gurvy should stay constant", prop.ForAll(
		func(start *GT) bool {
			var end GT
			*start = FinalExponentiation(start) // ensure we are in correct subgroup..
			gt := bls12381.NewGT()
			other, err := gt.FromBytes(start.Marshal())
			if err != nil {
				t.Log(err)
				return false
			}
			// reconstruct a GT from  bytes
			err = end.Unmarshal(gt.ToBytes(other))
			if err != nil {
				return false
			}
			return start.Equal(&end)
		},
		GenE12(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

func TestScalarMultiplicationInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] G1: scalarMultiplication interop", prop.ForAll(
		func(a, exp fp.Element) bool {
			var start, end G1Affine
			var ab, bExp big.Int
			a.ToBigIntRegular(&ab)
			exp.ToBigIntRegular(&bExp)
			start.ScalarMultiplication(&g1GenAff, &ab)

			g1 := bls12381.NewG1()
			other, err := g1.FromBytes(start.Marshal())
			if err != nil {
				t.Log(err)
				return false
			}

			// perform the scalar multiplications
			otherRes := g1.MulScalarBig(g1.New(), other, &bExp)
			end.ScalarMultiplication(&start, &bExp)

			if !(bytes.Equal(g1.ToBytes(otherRes), end.Marshal())) {
				t.Log("scalar multiplication between bls12-381 and gurvy is different")
				return false
			}

			return true
		},
		GenFp(),
		GenFp(),
	))

	properties.Property("[BLS381] G2: scalarMultiplication interop", prop.ForAll(
		func(a, exp fp.Element) bool {
			var start, end G2Affine
			var ab, bExp big.Int
			a.ToBigIntRegular(&ab)
			exp.ToBigIntRegular(&bExp)
			start.ScalarMultiplication(&g2GenAff, &ab)

			g2 := bls12381.NewG2()
			other, err := g2.FromBytes(start.Marshal())
			if err != nil {
				t.Log(err)
				return false
			}

			// perform the scalar multiplications
			otherRes := g2.MulScalarBig(g2.New(), other, &bExp)
			end.ScalarMultiplication(&start, &bExp)

			if !(bytes.Equal(g2.ToBytes(otherRes), end.Marshal())) {
				t.Log("scalar multiplication between bls12-381 and gurvy is different")
				return false
			}

			return true
		},
		GenFp(),
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestPointAdditionInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] checking point addition", prop.ForAll(
		func(a fp.Element) bool {
			var g1 G1Affine
			var g2 G2Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			g1.ScalarMultiplication(&g1GenAff, &ab)
			g2.ScalarMultiplication(&g2GenAff, &ab)

			// do the same with other lib
			bls12381g1 := bls12381.NewG1()
			otherG1, err := bls12381g1.FromBytes(g1.Marshal())
			if err != nil {
				return false
			}
			otherG1Gen, err := bls12381g1.FromBytes(g1GenAff.Marshal())
			if err != nil {
				return false
			}
			bls12381g2 := bls12381.NewG2()
			otherG2, err := bls12381g2.FromBytes(g2.Marshal())
			if err != nil {
				return false
			}
			otherG2Gen, err := bls12381g2.FromBytes(g2GenAff.Marshal())
			if err != nil {
				return false
			}

			// add g1 to g1Gen and g2 to g2gen
			var _g1 G1Jac
			var _g2 G2Jac
			_g1.FromAffine(&g1)
			_g2.FromAffine(&g2)

			_g1.AddAssign(&g1Gen)
			g1.FromJacobian(&_g1)

			_g2.AddAssign(&g2Gen)
			g2.FromJacobian(&_g2)

			// results
			r1 := bls12381g1.Add(bls12381g1.New(), otherG1, otherG1Gen)
			r2 := bls12381g2.Add(bls12381g2.New(), otherG2, otherG2Gen)

			if !(bytes.Equal(g1.Marshal(), bls12381g1.ToBytes(r1))) {
				t.Log("g1 point addition doesn't match other implementation")
				return false
			}

			if !(bytes.Equal(g2.Marshal(), bls12381g2.ToBytes(r2))) {
				t.Log("g2 point addition doesn't match other implementation")
				return false
			}

			return true
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
