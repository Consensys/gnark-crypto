package gkr

import (
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational/polynomial"
)

// fitPoly tries to fit a polynomial of maximum degree maxDeg to f
func fitPoly(f GateFunction, nbIn, maxDeg int) (p polynomial.Polynomial, ok bool, err error) {

	// turn f univariate by defining p(x) as f(x, x, ..., x)
	// evaluate p at random points
	x := make([]small_rational.SmallRational, maxDeg+1)
	y := make([]small_rational.SmallRational, maxDeg+1)
	fIn := make([]small_rational.SmallRational, nbIn)
	for i := range x {
		setRandom(&x[i])
		for j := range fIn {
			fIn[j] = x[i]
		}
		y[i] = f(fIn...)
	}

	// interpolate p
	p, err = polynomial.Interpolate(x, y)
	if err != nil {
		return nil, false, err
	}

	// check if p is equal to f. This not being the case means that f is of a degree higher than maxDeg
	setRandom(&fIn[0])
	for i := range fIn {
		fIn[i] = fIn[0]
	}
	pAt := p.Eval(&fIn[0])
	fAt := f(fIn...)
	if !pAt.Equal(&fAt) {
		return nil, false, nil
	}

	// trim p
	lastNonZero := len(p) - 1
	for lastNonZero >= 0 && p[lastNonZero].IsZero() {
		lastNonZero--
	}
	return p[:lastNonZero+1], true, nil
}
