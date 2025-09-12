package vortex

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

// BatchEvalFextPolyLagrange evaluates extension field polynomials in Lagrange basis
func BatchEvalFextPolyLagrange(polys [][]fext.E4, x fext.E4, oncoset ...bool) ([]fext.E4, error) {
	if len(polys) == 0 {
		return []fext.E4{}, nil
	}

	denominators, factor, _ := initialization(polys, x, oncoset...)

	// Check for edge case: x is a root of unity
	for i, denom := range denominators {
		if denom.IsZero() {
			results := make([]fext.E4, len(polys))
			for k := range polys {
				results[k] = polys[k][i]
			}
			return results, nil
		}
	}

	// Batch invert denominators
	denominators = fext.BatchInvertE4(denominators)

	// Compute results in parallel
	results := make([]fext.E4, len(polys))
	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {
			res := fext.Vector(polys[k]).InnerProduct(fext.Vector(denominators))
			res.Mul(&res, &factor)
			results[k] = res
		}
	})

	return results, nil
}

// BatchEvalBasePolyLagrange evaluates base field polynomials in Lagrange basis, returning extension field results
func BatchEvalBasePolyLagrange(polys [][]koalabear.Element, x fext.E4, oncoset ...bool) ([]fext.E4, error) {
	if len(polys) == 0 {
		return []fext.E4{}, nil
	}

	denominators, factor, _ := initialization(polys, x, oncoset...)

	// Check for edge case: x is a root of unity
	for i, denom := range denominators {
		if denom.IsZero() {
			results := make([]fext.E4, len(polys))
			for k := range polys {
				results[k] = fext.Lift(polys[k][i])
			}
			return results, nil
		}
	}

	// Batch invert denominators
	denominators = fext.BatchInvertE4(denominators)

	// Compute results in parallel
	results := make([]fext.E4, len(polys))
	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {
			res := fext.Vector(denominators).InnerProductByElement(polys[k])
			res.Mul(&res, &factor)
			results[k] = res
		}
	})

	return results, nil
}

// initialization computes the denominators and factor for Lagrange basis evaluation
func initialization[T any](polys [][]T, x fext.E4, oncoset ...bool) ([]fext.E4, fext.E4, error) {

	n := len(polys[0])
	validateInput(polys, n)

	generator, err := fft.Generator(uint64(n))
	if err != nil {
		return nil, fext.E4{}, fmt.Errorf("failed to get generator: %w", err)
	}

	generatorInv := new(koalabear.Element).Inverse(&generator)
	one := koalabear.One()

	// Handle coset evaluation
	if len(oncoset) > 0 && oncoset[0] {
		frMultiplicativeGen := fft.GeneratorFullMultiplicativeGroup()
		frMultiplicativeGenInv := new(koalabear.Element).Inverse(&frMultiplicativeGen)
		x.MulByElement(&x, frMultiplicativeGenInv)
	}

	// Build denominators: [x, x/ω, x/ω², ...] - 1
	denominators := make([]fext.E4, n)
	denominators[0] = x
	for i := 1; i < n; i++ {
		denominators[i].MulByElement(&denominators[i-1], generatorInv)
	}

	// Subtract 1 from each denominator
	for i := range denominators {
		denominators[i].B0.A0.Sub(&denominators[i].B0.A0, &one)
	}

	// Compute factor: (x^n - 1) / n
	var factor fext.E4
	factor.Exp(x, big.NewInt(int64(n)))
	factor.B0.A0.Sub(&factor.B0.A0, &one)

	cardInv := koalabear.NewElement(uint64(n))
	cardInv.Inverse(&cardInv)
	factor.MulByElement(&factor, &cardInv)

	return denominators, factor, nil
}

// validateInputBase validates base field polynomial input
func validateInput[T any](polys [][]T, expectedLen int) error {
	for i := range polys {
		if len(polys[i]) != expectedLen {
			return fmt.Errorf("all polys should have the same length, expected %d but poly[%d] has length %d",
				expectedLen, i, len(polys[i]))
		}
	}
	if !isPowerOfTwo(expectedLen) {
		return fmt.Errorf("only support powers of two but poly has length %v", expectedLen)
	}
	return nil
}
