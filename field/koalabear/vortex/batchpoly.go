package vortex

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

func BatchEvalFextPolyLagrange(polys [][]fext.E4, x fext.E4, oncoset ...bool) ([]fext.E4, error) {

	if len(polys) == 0 {
		return []fext.E4{}, nil
	}

	n := len(polys[0])
	for i := range polys {
		if len(polys[i]) != n {
			return []fext.E4{}, fmt.Errorf("all polys should have the same length %v", n)
		}
	}

	if !isPowerOfTwo(n) {
		return []fext.E4{}, fmt.Errorf("only support powers of two but poly has length %v", n)
	}

	var (
		denominators = make([]fext.E4, n)
		one          = koalabear.One()
		generator, _ = fft.Generator(uint64(n))
		generatorInv = new(koalabear.Element).Inverse(&generator)
		cardInv      koalabear.Element
		results      = make([]fext.E4, len(polys))
	)

	cardInv = koalabear.NewElement(uint64(n))
	cardInv.Inverse(&cardInv)

	if len(oncoset) > 0 && oncoset[0] {
		frMultiplicativeGen := fft.GeneratorFullMultiplicativeGroup()
		frMultiplicativeGenInv := new(koalabear.Element)
		frMultiplicativeGenInv.Inverse(&frMultiplicativeGen)
		x.MulByElement(&x, frMultiplicativeGenInv)
	}

	// The denominator is constructed as:
	// 		D_x = \frac{X}{x} - g for x \in H
	// 	where H is the subgroup of the roots of unity (not the coset)
	// 	and g a field element such that gH is the coset.
	denominators[0] = x
	for i := 1; i < n; i++ {
		denominators[i].MulByElement(&denominators[i-1], generatorInv)
	}

	for i := 0; i < n; i++ {
		// This subtracts a field extension by a base field element.
		denominators[i].B0.A0.Sub(&denominators[i].B0.A0, &one)
		if denominators[i].IsZero() {
			// edge-case : x is a root of unity of the domain. In this case, we can just return
			// the associated value for poly
			for k := range polys {
				results[k] = polys[k][i]
			}

			return results, nil
		}
	}

	/*
		Then, we compute the sum between the inverse of the denominator
		and the poly

		\sum_{x \in H}\frac{P(gx)}{D_x}
	*/
	denominators = fext.BatchInvertE4(denominators)

	factor := fext.E4{}

	// Precompute the value of x^n once outside the loop
	factor.Exp(x, big.NewInt(int64(n)))
	factor.B0.A0.Sub(&factor.B0.A0, &one)
	factor.MulByElement(&factor, &cardInv)

	// Precompute the value of x^n once outside the loop

	parallel.Execute(len(polys), func(start, stop int) {
		for k := start; k < stop; k++ {

			// Compute the scalar product.
			res := fext.Vector(polys[k]).InnerProduct(fext.Vector(denominators))

			// Multiply res with factor.
			res.Mul(&res, &factor)

			// Store the result.
			results[k] = res
		}
	})

	return results, nil
}
