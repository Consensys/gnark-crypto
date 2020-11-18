package bn256

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/consensys/gurvy/bn256/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"

	cloudflare "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	google "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

// Test against go-ethereum/crypto/bn256 implementations (google and cloudflare)

func TestG1AffineSerializationInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BN256] G1: gurvy -> cloudflare -> gurvy should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G1Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g1GenAff, &ab)

			other, err := cloudflareG1(&start)
			if err != nil {
				t.Log(err)
				return false
			}

			// reconstruct a gurvy point from  bytes
			err = end.Unmarshal(other.Marshal())
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		GenFp(),
	))

	properties.Property("[BN256] G1: gurvy -> google -> gurvy should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G1Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g1GenAff, &ab)

			other, err := googleG1(&start)
			if err != nil {
				t.Log(err)
				return false
			}

			// reconstruct a gurvy point from  bytes
			err = end.Unmarshal(other.Marshal())
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

	properties.Property("[BN256] G2: gurvy -> cloudflare -> gurvy should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G2Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g2GenAff, &ab)

			other, err := cloudflareG2(&start)
			if err != nil {
				t.Log(err)
				return false
			}

			// reconstruct a gurvy point from  bytes
			err = end.Unmarshal(other.Marshal())
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		GenFp(),
	))

	properties.Property("[BN256] G2: gurvy -> google -> gurvy should stay constant", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G2Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			start.ScalarMultiplication(&g2GenAff, &ab)

			other, err := googleG2(&start)
			if err != nil {
				t.Log(err)
				return false
			}

			// reconstruct a gurvy point from  bytes
			err = end.Unmarshal(other.Marshal())
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
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BN256] GT: gurvy -> cloudflare -> gurvy should stay constant", prop.ForAll(
		func(start *GT) bool {
			var end GT
			cgt := new(cloudflare.GT)

			if _, err := cgt.Unmarshal(start.Marshal()); err != nil {
				t.Log(err)
				return false
			}

			err := end.Unmarshal(cgt.Marshal())
			if err != nil {
				return false
			}
			return start.Equal(&end)
		},
		GenE12(),
	))

	properties.Property("[BN256] GT: gurvy -> google -> gurvy should stay constant", prop.ForAll(
		func(start *GT) bool {
			var end GT
			cgt := new(google.GT)

			if _, ok := cgt.Unmarshal(start.Marshal()); !ok {
				return false
			}

			err := end.Unmarshal(cgt.Marshal())
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

	properties.Property("[BN256] G1: scalarMultiplication interop", prop.ForAll(
		func(a, exp fp.Element) bool {
			var start, end G1Affine
			var ab, bExp big.Int
			a.ToBigIntRegular(&ab)
			exp.ToBigIntRegular(&bExp)
			start.ScalarMultiplication(&g1GenAff, &ab)

			gPoint, err := googleG1(&start)
			if err != nil {
				t.Log(err)
				return false
			}
			cPoint, err := cloudflareG1(&start)
			if err != nil {
				t.Log(err)
				return false
			}

			// perform the scalar multiplications
			gPoint.ScalarMult(gPoint, &bExp)
			cPoint.ScalarMult(cPoint, &bExp)
			end.ScalarMultiplication(&start, &bExp)

			if !(bytes.Equal(gPoint.Marshal(), end.Marshal())) {
				t.Log("scalar multiplication between google and gurvy is different")
				return false
			}

			if !(bytes.Equal(cPoint.Marshal(), end.Marshal())) {
				t.Log("scalar multiplication between cloudflare and gurvy is different")
				return false
			}
			return true
		},
		GenFp(),
		GenFp(),
	))

	properties.Property("[BN256] G2: scalarMultiplication interop", prop.ForAll(
		func(a, exp fp.Element) bool {
			var start, end G2Affine
			var ab, bExp big.Int
			a.ToBigIntRegular(&ab)
			exp.ToBigIntRegular(&bExp)
			start.ScalarMultiplication(&g2GenAff, &ab)

			gPoint, err := googleG2(&start)
			if err != nil {
				t.Log(err)
				return false
			}
			cPoint, err := cloudflareG2(&start)
			if err != nil {
				t.Log(err)
				return false
			}
			// perform the scalar multiplications
			gPoint.ScalarMult(gPoint, &bExp)
			cPoint.ScalarMult(cPoint, &bExp)
			end.ScalarMultiplication(&start, &bExp)

			if !(bytes.Equal(gPoint.Marshal(), end.Marshal())) {
				t.Log("scalar multiplication between google and gurvy is different")
				return false
			}

			if !(bytes.Equal(cPoint.Marshal(), end.Marshal())) {
				t.Log("scalar multiplication between cloudflare and gurvy is different")
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

	properties.Property("[BN256] checking point addition", prop.ForAll(
		func(a fp.Element) bool {
			var g1 G1Affine
			var g2 G2Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			g1.ScalarMultiplication(&g1GenAff, &ab)
			g2.ScalarMultiplication(&g2GenAff, &ab)

			// do the same with google and cloud flare
			g1g, err := googleG1(&g1)
			if err != nil {
				t.Log(err)
				return false
			}
			g1c, err := cloudflareG1(&g1)
			if err != nil {
				t.Log(err)
				return false
			}
			g2g, err := googleG2(&g2)
			if err != nil {
				t.Log(err)
				return false
			}
			g2c, err := cloudflareG2(&g2)
			if err != nil {
				t.Log(err)
				return false
			}
			g1gGen, err := googleG1(&g1GenAff)
			if err != nil {
				t.Log(err)
				return false
			}
			g1cGen, err := cloudflareG1(&g1GenAff)
			if err != nil {
				t.Log(err)
				return false
			}
			g2gGen, err := googleG2(&g2GenAff)
			if err != nil {
				t.Log(err)
				return false
			}
			g2cGen, err := cloudflareG2(&g2GenAff)
			if err != nil {
				t.Log(err)
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
			g1c.Add(g1c, g1cGen)
			g1g.Add(g1g, g1gGen)
			g2c.Add(g2c, g2cGen)
			g2g.Add(g2g, g2gGen)

			if !(bytes.Equal(g1.Marshal(), g1c.Marshal())) {
				t.Log("g1 point addition doesn't match google implementation")
				return false
			}

			if !(bytes.Equal(g1.Marshal(), g1g.Marshal())) {
				t.Log("g1 point addition doesn't match cloudflare implementation")
				return false
			}

			if !(bytes.Equal(g2.Marshal(), g2c.Marshal())) {
				t.Log("g2 point addition doesn't match google implementation")
				return false
			}

			if !(bytes.Equal(g2.Marshal(), g2g.Marshal())) {
				t.Log("g2 point addition doesn't match cloudflare implementation")
				return false
			}

			return true
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestPairingInterop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("[BN256] pairing check interop", prop.ForAll(
		func(a fp.Element) bool {
			var g1 G1Affine
			var g2 G2Affine
			var ab big.Int
			a.ToBigIntRegular(&ab)
			g1.ScalarMultiplication(&g1GenAff, &ab)
			g2.ScalarMultiplication(&g2GenAff, &ab)

			g1g, err := googleG1(&g1)
			if err != nil {
				t.Log(err)
				return false
			}
			g2g, err := googleG2(&g2)
			if err != nil {
				t.Log(err)
				return false
			}

			g1c, err := cloudflareG1(&g1)
			if err != nil {
				t.Log(err)
				return false
			}
			g2c, err := cloudflareG2(&g2)
			if err != nil {
				t.Log(err)
				return false
			}

			// pairings
			pc := cloudflare.Pair(g1c, g2c)
			gc := google.Pair(g1g, g2g)
			c := Pair(g1, g2)

			if !(bytes.Equal(c.Marshal(), gc.Marshal())) {
				t.Log("pairing doesn't match google implementation")
				return false
			}

			if !(bytes.Equal(c.Marshal(), pc.Marshal())) {
				t.Log("pairing doesn't match cloudflare implementation")
				return false
			}

			return true
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func BenchmarkPairingInterop(b *testing.B) {
	var g1 G1Affine
	var g2 G2Affine
	var ab big.Int
	ab.SetUint64(42)
	g1.ScalarMultiplication(&g1GenAff, &ab)
	g2.ScalarMultiplication(&g2GenAff, &ab)

	b.Run("[bn256]google_pairing", func(b *testing.B) {
		g1g, err := googleG1(&g1)
		if err != nil {
			b.Fatal(err)
		}
		g2g, err := googleG2(&g2)
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = google.Pair(g1g, g2g)
		}
	})
	b.Run("[bn256]cloudflare_pairing", func(b *testing.B) {
		g1c, err := cloudflareG1(&g1)
		if err != nil {
			b.Fatal(err)
		}
		g2c, err := cloudflareG2(&g2)
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cloudflare.Pair(g1c, g2c)
		}
	})

	b.Run("[bn256]gurvy_pairing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Pair(g1, g2)
		}
	})

}

func BenchmarkPointAdditionInterop(b *testing.B) {
	var g1 G1Affine
	var ab big.Int
	ab.SetUint64(42)
	g1.ScalarMultiplication(&g1GenAff, &ab)

	b.Run("[bn256]cloudflare_add_jacobian", func(b *testing.B) {
		g1g, err := cloudflareG1(&g1)
		if err != nil {
			b.Fatal(err)
		}
		g1gen, err := cloudflareG1(&g1GenAff)
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			g1g.Add(g1g, g1gen)
		}
	})

	b.Run("[bn256]gurvy_add_jacobian", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var _g1 G1Jac
			_g1.FromAffine(&g1)
			_g1.AddAssign(&g1Gen)
		}
	})

}

func cloudflareG1(p *G1Affine) (*cloudflare.G1, error) {
	r := new(cloudflare.G1)
	if _, err := r.Unmarshal(p.Marshal()); err != nil {
		return nil, err
	}
	return r, nil
}

func cloudflareG2(p *G2Affine) (*cloudflare.G2, error) {
	r := new(cloudflare.G2)
	if _, err := r.Unmarshal(p.Marshal()); err != nil {
		return nil, err
	}
	return r, nil
}

func googleG1(p *G1Affine) (*google.G1, error) {
	r := new(google.G1)
	if _, err := r.Unmarshal(p.Marshal()); err != nil {
		return nil, err
	}
	return r, nil
}

func googleG2(p *G2Affine) (*google.G2, error) {
	r := new(google.G2)
	if _, err := r.Unmarshal(p.Marshal()); err != nil {
		return nil, err
	}
	return r, nil
}
