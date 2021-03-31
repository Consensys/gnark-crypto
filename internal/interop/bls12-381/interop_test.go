package interop

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"testing"

	bls12381 "github.com/kilic/bls12-381"

	bls381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// Test against github.com/kilic/bls12-381

var (
	g1Gen    bls381.G1Jac
	g2Gen    bls381.G2Jac
	g1GenAff bls381.G1Affine
	g2GenAff bls381.G2Affine
)

func init() {
	g1Gen, g2Gen, g1GenAff, g2GenAff = bls381.Generators()
}

func TestG1AffineSerializationInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] G1: gnark-crypto -> bls12-381 -> gnark-crypto should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end bls381.G1Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g1GenAff, &ab)

			g1 := bls12381.NewG1()
			other, err := g1.FromBytes(start.Marshal())
			if err != nil {
				t.Log(err)
				return false
			}
			// reconstruct a gnark-crypto point from  bytes
			err = end.Unmarshal(g1.ToBytes(other))
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		genFp(),
	))

	properties.Property("[BLS381] G1 compressed: gnark-crypto -> bls12-381 -> gnark-crypto should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end bls381.G1Affine
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
			// reconstruct a gnark-crypto point from  bytes
			err = end.Unmarshal(g1.ToCompressed(other))
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		genFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

func TestG2AffineSerializationInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] G2: gnark-crypto -> bls12-381 -> gnark-crypto should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end bls381.G2Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g2GenAff, &ab)

			g2 := bls12381.NewG2()
			other, err := g2.FromBytes(start.Marshal())
			if err != nil {
				t.Log(err)
				return false
			}
			// reconstruct a gnark-crypto point from  bytes
			err = end.Unmarshal(g2.ToBytes(other))
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		genFp(),
	))

	properties.Property("[BLS381] G2 compressed: gnark-crypto -> bls12-381 -> gnark-crypto should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end bls381.G2Affine
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
			// reconstruct a gnark-crypto point from  bytes
			err = end.Unmarshal(g2.ToCompressed(other))
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		genFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

func TestGTSerializationInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 5

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] bls381.GT: gnark-crypto -> bls12-381 -> gnark-crypto should stay constant", prop.ForAll(
		func(start *bls381.GT) bool {
			var end bls381.GT
			*start = bls381.FinalExponentiation(start) // ensure we are in correct subgroup..
			gt := bls12381.NewGT()
			other, err := gt.FromBytes(start.Marshal())
			if err != nil {
				t.Log(err)
				return false
			}
			// reconstruct a bls381.GT from  bytes
			err = end.Unmarshal(gt.ToBytes(other))
			if err != nil {
				return false
			}
			return start.Equal(&end)
		},
		genGT(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))

}

func TestScalarMultiplicationInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] G1: scalarMultiplication interop", prop.ForAll(
		func(a, exp fp.Element) bool {
			var start, end bls381.G1Affine
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
				t.Log("scalar multiplication between bls12-381 and gnark-crypto is different")
				return false
			}

			return true
		},
		genFp(),
		genFp(),
	))

	properties.Property("[BLS381] G2: scalarMultiplication interop", prop.ForAll(
		func(a, exp fp.Element) bool {
			var start, end bls381.G2Affine
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
				t.Log("scalar multiplication between bls12-381 and gnark-crypto is different")
				return false
			}

			return true
		},
		genFp(),
		genFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestPointAdditionInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] checking point addition", prop.ForAll(
		func(a fp.Element) bool {
			var g1 bls381.G1Affine
			var g2 bls381.G2Affine
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
			var _g1 bls381.G1Jac
			var _g2 bls381.G2Jac
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
		genFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestPairingInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BLS381] pairing check interop", prop.ForAll(
		func(a fp.Element) bool {
			var g1 bls381.G1Affine
			var g2 bls381.G2Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			g1.ScalarMultiplication(&g1GenAff, &ab)
			g2.ScalarMultiplication(&g2GenAff, &ab)

			// do the same with other lib
			otherG1, err := bls12381.NewG1().FromBytes(g1.Marshal())
			if err != nil {
				return false
			}
			otherG2, err := bls12381.NewG2().FromBytes(g2.Marshal())
			if err != nil {
				return false
			}

			// pairings
			engine := bls12381.NewEngine()
			engine.AddPair(otherG1, otherG2)
			otherResult := engine.Result()
			c, _ := bls381.Pair([]bls381.G1Affine{g1}, []bls381.G2Affine{g2})

			if !(bytes.Equal(c.Marshal(), bls12381.NewGT().ToBytes(otherResult))) {
				t.Log("pairing doesn't match other implementation")
				return false
			}

			return true
		},
		genFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func BenchmarkPairingInterop(b *testing.B) {
	var g1 bls381.G1Affine
	var g2 bls381.G2Affine
	var ab big.Int
	ab.SetUint64(42)
	g1.ScalarMultiplication(&g1GenAff, &ab)
	g2.ScalarMultiplication(&g2GenAff, &ab)

	b.Run("[BLS381]bls12381_pairing", func(b *testing.B) {
		otherG1, err := bls12381.NewG1().FromBytes(g1.Marshal())
		if err != nil {
			b.Fatal(err)
		}
		otherG2, err := bls12381.NewG2().FromBytes(g2.Marshal())
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine := bls12381.NewEngine()
			engine.AddPair(otherG1, otherG2)
			_ = engine.Result()
		}
	})

	b.Run("[BLS381]gnark-crypto_pairing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = bls381.Pair([]bls381.G1Affine{g1}, []bls381.G2Affine{g2})
		}
	})

}

func genFp() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var elmt fp.Element
		var b [fp.Bytes]byte
		rand.Read(b[:])
		elmt.SetBytes(b[:])
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}

func genGT() gopter.Gen {
	return gopter.CombineGens(
		genFp(),
		genFp(),
		genFp(),
		genFp(),
		genFp(),
		genFp(),
		genFp(),
		genFp(),
		genFp(),
		genFp(),
		genFp(),
		genFp(),
	).Map(func(values []interface{}) *bls381.GT {
		var b [fp.Bytes * 12]byte
		rand.Read(b[:])
		var r bls381.GT
		offset := 0
		r.C0.B0.A0.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C0.B0.A1.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C0.B1.A0.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C0.B1.A1.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C0.B2.A0.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C0.B2.A1.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes

		r.C1.B0.A0.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C1.B0.A1.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C1.B1.A0.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C1.B1.A1.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C1.B2.A0.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes
		r.C1.B2.A1.SetBytes(b[offset : offset+fp.Bytes])
		offset += fp.Bytes

		return &r
	})
}
