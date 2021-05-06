package kzg

import (
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bn254_pol "github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
)

// dividePolyByXminusA computes (f-f(a))/(x-a), in canonical basis, in regular form
func dividePolyByXminusA(d fft.Domain, f bn254_pol.Polynomial, fa, a fr.Element) bn254_pol.Polynomial {

	// padd f so it has size d.Cardinality
	_f := make([]fr.Element, d.Cardinality)
	copy(_f, f)

	// compute the quotient (f-f(a))/(x-a)
	d.FFT(_f, fft.DIF, 0)

	var acc, sub fr.Element
	acc.SetOne()
	n := d.Cardinality
	nn := uint64(64 - bits.TrailingZeros64(n))
	for i := 0; i < len(_f); i++ {
		irev := bits.Reverse64(uint64(i)) >> nn
		// TODO perform batch inversion
		sub.Sub(&acc, &a)
		_f[irev].Sub(&_f[irev], &fa).Div(&_f[irev], &sub)
		acc.Mul(&acc, &d.Generator)
	}

	d.FFTInverse(_f, fft.DIT, 0)

	// the result is of degree deg(f)-1
	return _f[:len(f)-1]
}
