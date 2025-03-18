package gkr

import (
	"errors"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational/polynomial"
)

// fitPoly tries to fit a polynomial of degree less than degreeBound to f.
// degreeBound must be a power of 2.
// It returns the polynomial if successful, nil otherwise
func fitPoly(f GateFunction, nbIn int, degreeBound uint64) polynomial.Polynomial {
	// turn f univariate by defining p(x) as f(x, x, ..., x)
	// evaluate p at random points
	x := make([]small_rational.SmallRational, degreeBound)
	y := make([]small_rational.SmallRational, degreeBound)
	fIn := make([]small_rational.SmallRational, nbIn)
	for i := range x {
		setRandom(&x[i])
		for j := range fIn {
			fIn[j] = x[i]
		}
		y[i] = f(fIn...)
	}

	// interpolate p
	p, err := interpolate(x, y)
	if err != nil {
		panic(err)
	}

	// check if p is equal to f. This not being the case means that f is of a degree higher than maxDeg
	setRandom(&fIn[0])
	for i := range fIn {
		fIn[i] = fIn[0]
	}
	pAt := p.Eval(&fIn[0])
	fAt := f(fIn...)
	if !pAt.Equal(&fAt) {
		return nil
	}

	// trim p
	lastNonZero := len(p) - 1
	for lastNonZero >= 0 && p[lastNonZero].IsZero() {
		lastNonZero--
	}
	return p[:lastNonZero+1]
}

// interpolate fits a polynomial of degree len(X) - 1 = len(Y) - 1 to the points (X[i], Y[i])
// Note that the runtime is O(len(X)³)
func interpolate(X, Y []small_rational.SmallRational) (polynomial.Polynomial, error) {
	if len(X) != len(Y) {
		return nil, errors.New("X and Y must have the same length")
	}

	// solve the system of equations by Gaussian elimination
	augmentedRows := make([][]small_rational.SmallRational, len(X)) // the last column is the Y values
	for i := range augmentedRows {
		augmentedRows[i] = make([]small_rational.SmallRational, len(X)+1)
		augmentedRows[i][0].SetOne()
		augmentedRows[i][1].Set(&X[i])
		for j := 2; j < len(augmentedRows[i])-1; j++ {
			augmentedRows[i][j].Mul(&augmentedRows[i][j-1], &X[i])
		}
		augmentedRows[i][len(augmentedRows[i])-1].Set(&Y[i])
	}

	// make the upper triangle
	for i := range len(augmentedRows) - 1 {
		// use row i to eliminate the ith element in all rows below
		var negInv small_rational.SmallRational
		if augmentedRows[i][i].IsZero() {
			return nil, errors.New("singular matrix")
		}
		negInv.Inverse(&augmentedRows[i][i])
		negInv.Neg(&negInv)
		for j := i + 1; j < len(augmentedRows); j++ {
			var c small_rational.SmallRational
			c.Mul(&augmentedRows[j][i], &negInv)
			// augmentedRows[j][i].SetZero() omitted
			for k := i + 1; k < len(augmentedRows[i]); k++ {
				var t small_rational.SmallRational
				t.Mul(&augmentedRows[i][k], &c)
				augmentedRows[j][k].Add(&augmentedRows[j][k], &t)
			}
		}
	}

	// back substitution
	res := make(polynomial.Polynomial, len(X))
	for i := len(augmentedRows) - 1; i >= 0; i-- {
		res[i] = augmentedRows[i][len(augmentedRows[i])-1]
		for j := i + 1; j < len(augmentedRows[i])-1; j++ {
			var t small_rational.SmallRational
			t.Mul(&res[j], &augmentedRows[i][j])
			res[i].Sub(&res[i], &t)
		}
		res[i].Div(&res[i], &augmentedRows[i][i])
	}

	return res, nil
}
