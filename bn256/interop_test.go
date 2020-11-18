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

			// create a cloudflare point
			cloudflarePoint := new(cloudflare.G1)

			if _, err := cloudflarePoint.Unmarshal(start.Marshal()); err != nil {
				t.Log(err)
				return false
			}

			// reconstruct a gurvy point from cloudflare bytes
			err := end.Unmarshal(cloudflarePoint.Marshal())
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

			// create a google point
			googlePoint := new(google.G1)

			if _, err := googlePoint.Unmarshal(start.Marshal()); err != nil {
				t.Log(err)
				return false
			}

			// reconstruct a gurvy point from google bytes
			err := end.Unmarshal(googlePoint.Marshal())
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

			// create a cloudflare point
			cloudflarePoint := new(cloudflare.G2)

			if _, err := cloudflarePoint.Unmarshal(start.Marshal()); err != nil {
				t.Log(err)
				return false
			}

			// reconstruct a gurvy point from cloudflare bytes
			err := end.Unmarshal(cloudflarePoint.Marshal())
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

			// create a google point
			googlePoint := new(google.G2)

			if _, err := googlePoint.Unmarshal(start.Marshal()); err != nil {
				t.Log(err)
				return false
			}

			// reconstruct a gurvy point from google bytes
			err := end.Unmarshal(googlePoint.Marshal())
			if err != nil {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		GenFp(),
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

			gPoint := new(google.G1)
			cPoint := new(cloudflare.G1)
			if _, err := gPoint.Unmarshal(start.Marshal()); err != nil {
				t.Log(err)
				return false
			}
			if _, err := cPoint.Unmarshal(start.Marshal()); err != nil {
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

			gPoint := new(google.G2)
			cPoint := new(cloudflare.G2)
			if _, err := gPoint.Unmarshal(start.Marshal()); err != nil {
				t.Log(err)
				return false
			}
			if _, err := cPoint.Unmarshal(start.Marshal()); err != nil {
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
