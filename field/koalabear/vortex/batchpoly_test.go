package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/stretchr/testify/require"
)

func randomPoly(size int) []koalabear.Element {
	res := make([]koalabear.Element, size)
	for i := range res {
		res[i].SetRandom()
	}
	return res
}

func randomPolyExt(size int) []fext.E4 {
	res := make([]fext.E4, size)
	for i := range res {
		res[i].SetRandom()
	}
	return res
}

func TestBatchEvaluateLagrangeOnFext(t *testing.T) {
	const sizePoly = 16
	const nbPoly = 20

	// Generate test polynomials and evaluation point
	polys := make([][]fext.E4, nbPoly)
	for i := range polys {
		polys[i] = randomPolyExt(sizePoly)
	}

	var x fext.E4
	x.SetRandom()

	// Compute expected results using Horner evaluation
	expected := make([]fext.E4, nbPoly)
	for i := range expected {
		expected[i] = EvalFextPolyHorner(polys[i], x)
	}

	domain := fft.NewDomain(uint64(sizePoly))

	testCases := []struct {
		name    string
		onCoset bool
	}{
		{"without coset", false},
		{"with coset", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Transform polynomials to Lagrange basis
			lagrangePolys := make([][]fext.E4, nbPoly)
			for i := range lagrangePolys {
				lagrangePolys[i] = append([]fext.E4{}, polys[i]...)
				if tc.onCoset {
					domain.FFTExt(lagrangePolys[i], fft.DIF, fft.OnCoset())
				} else {
					domain.FFTExt(lagrangePolys[i], fft.DIF)
				}
				fft.BitReverse(lagrangePolys[i])
			}

			// Evaluate using Lagrange basis
			results, err := BatchEvalFextPolyLagrange(lagrangePolys, x, tc.onCoset)
			require.NoError(t, err)

			// Verify results
			for i := range results {
				require.Equal(t, expected[i].String(), results[i].String(),
					"Mismatch at polynomial %d", i)
			}
		})
	}
}

func TestBatchEvalBasePolyLagrange(t *testing.T) {
	const sizePoly = 64
	const nbPoly = 20

	// Generate test polynomials and evaluation point
	polys := make([][]koalabear.Element, nbPoly)
	for i := range polys {
		polys[i] = randomPoly(sizePoly)
	}

	var x fext.E4
	x.SetRandom()

	// Compute expected results using Horner evaluation
	expected := make([]fext.E4, nbPoly)
	for i := range expected {
		expected[i] = EvalBasePolyHorner(polys[i], x)
	}

	domain := fft.NewDomain(uint64(sizePoly))

	testCases := []struct {
		name    string
		onCoset bool
	}{
		{"without coset", false},
		{"with coset", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Transform polynomials to Lagrange basis
			lagrangePolys := make([][]koalabear.Element, nbPoly)
			for i := range lagrangePolys {
				lagrangePolys[i] = append([]koalabear.Element{}, polys[i]...)
				if tc.onCoset {
					domain.FFT(lagrangePolys[i], fft.DIF, fft.OnCoset())
				} else {
					domain.FFT(lagrangePolys[i], fft.DIF)
				}
				fft.BitReverse(lagrangePolys[i])
			}

			// Evaluate using Lagrange basis
			results, err := BatchEvalBasePolyLagrange(lagrangePolys, x, tc.onCoset)
			require.NoError(t, err)

			// Verify results
			for i := range results {
				require.Equal(t, expected[i].String(), results[i].String(),
					"Mismatch at polynomial %d", i)
			}
		})
	}
}
