// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package gkr

import (
	"errors"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational/polynomial"
	"slices"
	"sync"
)

type GateName string

var (
	gates     = make(map[GateName]*Gate)
	gatesLock sync.Mutex
)

type registerGateSettings struct {
	solvableVar               int
	noSolvableVarVerification bool
	noDegreeVerification      bool
	degree                    int
}

type RegisterGateOption func(*registerGateSettings)

// WithSolvableVar gives the index of a variable whose value can be uniquely determined from that of the other variables along with the gate's output.
// RegisterGate will return an error if it cannot verify that this claim is correct.
func WithSolvableVar(solvableVar int) RegisterGateOption {
	return func(settings *registerGateSettings) {
		settings.solvableVar = solvableVar
	}
}

// WithUnverifiedSolvableVar sets the index of a variable whose value can be uniquely determined from that of the other variables along with the gate's output.
// RegisterGate will not verify that the given index is correct.
func WithUnverifiedSolvableVar(solvableVar int) RegisterGateOption {
	return func(settings *registerGateSettings) {
		settings.noSolvableVarVerification = true
		settings.solvableVar = solvableVar
	}
}

// WithNoSolvableVar sets the gate as having no variable whose value can be uniquely determined from that of the other variables along with the gate's output.
// RegisterGate will not check the correctness of this claim.
func WithNoSolvableVar() RegisterGateOption {
	return func(settings *registerGateSettings) {
		settings.solvableVar = -1
		settings.noSolvableVarVerification = true
	}
}

// WithUnverifiedDegree sets the degree of the gate. RegisterGate will not verify that the given degree is correct.
func WithUnverifiedDegree(degree int) RegisterGateOption {
	return func(settings *registerGateSettings) {
		settings.noDegreeVerification = true
		settings.degree = degree
	}
}

// WithDegree sets the degree of the gate. RegisterGate will return an error if the degree is not correct.
func WithDegree(degree int) RegisterGateOption {
	return func(settings *registerGateSettings) {
		settings.degree = degree
	}
}

// isAdditive returns whether x_i occurs only in a monomial of total degree 1 in f
func (f GateFunction) isAdditive(i, nbIn int) bool {
	// fix all variables except the i-th one at random points
	// pick random value x1 for the i-th variable
	// check if f(-, 0, -) + f(-, 2*x1, -) = 2*f(-, x1, -)
	x := make(small_rational.Vector, nbIn)
	x.MustSetRandom()
	x0 := x[i]
	x[i].SetZero()
	in := slices.Clone(x)
	y0 := f(in...)

	x[i] = x0
	copy(in, x)
	y1 := f(in...)

	x[i].Double(&x[i])
	copy(in, x)
	y2 := f(in...)

	y2.Sub(&y2, &y1)
	y1.Sub(&y1, &y0)

	if !y2.Equal(&y1) {
		return false // not linear
	}

	// check if the coefficient of x_i is nonzero and independent of the other variables (so that we know it is ALWAYS nonzero)
	if y1.IsZero() { // f(-, x1, -) = f(-, 0, -), so the coefficient of x_i is 0
		return false
	}

	// compute the slope with another assignment for the other variables
	x.MustSetRandom()
	x[i].SetZero()
	copy(in, x)
	y0 = f(in...)

	x[i] = x0
	copy(in, x)
	y1 = f(in...)

	y1.Sub(&y1, &y0)

	return y1.Equal(&y2)
}

// fitPoly tries to fit a polynomial of degree less than degreeBound to f.
// degreeBound must be a power of 2.
// It returns the polynomial if successful, nil otherwise
func (f GateFunction) fitPoly(nbIn int, degreeBound uint64) polynomial.Polynomial {
	// turn f univariate by defining p(x) as f(x, x, ..., x)
	fIn := make([]small_rational.SmallRational, nbIn)
	p := make(polynomial.Polynomial, degreeBound)
	x := make(small_rational.Vector, degreeBound)
	x.MustSetRandom()
	for i := range x {
		for j := range fIn {
			fIn[j] = x[i]
		}
		p[i] = f(fIn...)
	}

	// obtain p's coefficients
	p, err := interpolate(x, p)
	if err != nil {
		panic(err)
	}

	// check if p is equal to f. This not being the case means that f is of a degree higher than degreeBound
	fIn[0].MustSetRandom()
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

// FindDegree returns the degree of the gate function, or -1 if it fails.
// Failure could be due to the degree being higher than max or the function not being a polynomial at all.
func (f GateFunction) FindDegree(max, nbIn int) int {
	bound := uint64(max) + 1
	for degreeBound := uint64(4); degreeBound <= bound; degreeBound *= 2 {
		if p := f.fitPoly(nbIn, degreeBound); p != nil {
			return len(p) - 1
		}
	}
	return -1
}

func (f GateFunction) VerifyDegree(claimedDegree, nbIn int) error {
	if p := f.fitPoly(nbIn, ecc.NextPowerOfTwo(uint64(claimedDegree)+1)); p == nil {
		return fmt.Errorf("detected a higher degree than %d", claimedDegree)
	} else if len(p)-1 != claimedDegree {
		return fmt.Errorf("detected degree %d, claimed %d", len(p)-1, claimedDegree)
	}
	return nil
}

// FindSolvableVar returns the index of a variable whose value can be uniquely determined from that of the other variables along with the gate's output.
// It returns -1 if it fails to find one.
// nbIn is the number of inputs to the gate
func (f GateFunction) FindSolvableVar(nbIn int) int {
	for i := range nbIn {
		if f.isAdditive(i, nbIn) {
			return i
		}
	}
	return -1
}

// IsVarSolvable returns whether claimedSolvableVar is a variable whose value can be uniquely determined from that of the other variables along with the gate's output.
// It returns false if it fails to verify this claim.
// nbIn is the number of inputs to the gate.
func (f GateFunction) IsVarSolvable(claimedSolvableVar, nbIn int) bool {
	return f.isAdditive(claimedSolvableVar, nbIn)
}

// RegisterGate creates a gate object and stores it in the gates registry.
// name is a human-readable name for the gate.
// f is the polynomial function defining the gate.
// nbIn is the number of inputs to the gate.
func RegisterGate(name GateName, f GateFunction, nbIn int, options ...RegisterGateOption) error {
	s := registerGateSettings{degree: -1, solvableVar: -1}
	for _, option := range options {
		option(&s)
	}

	if s.degree == -1 { // find a degree
		if s.noDegreeVerification {
			panic("invalid settings")
		}
		const maxAutoDegreeBound = 32
		if s.degree = f.FindDegree(32, nbIn); s.degree == -1 {
			return fmt.Errorf("could not find a degree for gate %s: tried up to %d", name, maxAutoDegreeBound-1)
		}
	} else {
		if !s.noDegreeVerification { // check that the given degree is correct
			if err := f.VerifyDegree(s.degree, nbIn); err != nil {
				return fmt.Errorf("for gate %s: %v", name, err)
			}
		}
	}

	if s.solvableVar == -1 {
		if !s.noSolvableVarVerification { // find a solvable variable
			s.solvableVar = f.FindSolvableVar(nbIn)
		}
	} else {
		// solvable variable given
		if !s.noSolvableVarVerification && !f.IsVarSolvable(s.solvableVar, nbIn) {
			return fmt.Errorf("cannot verify the solvability of variable %d in gate %s", s.solvableVar, name)
		}
	}

	gatesLock.Lock()
	defer gatesLock.Unlock()
	gates[name] = &Gate{Evaluate: f, nbIn: nbIn, degree: s.degree, solvableVar: s.solvableVar}
	return nil
}

func GetGate(name GateName) *Gate {
	gatesLock.Lock()
	defer gatesLock.Unlock()
	return gates[name]
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

const (
	IdentityGateName GateName = "identity"
	Add2GateName     GateName = "add2"
	Sub2GateName     GateName = "sub2"
	NegGateName      GateName = "neg"
	Mul2GateName     GateName = "mul2"
)

func init() {
	// register some basic gates

	if err := RegisterGate(IdentityGateName, func(x ...small_rational.SmallRational) small_rational.SmallRational {
		return x[0]
	}, 1, WithUnverifiedDegree(1), WithUnverifiedSolvableVar(0)); err != nil {
		panic(err)
	}

	if err := RegisterGate(Add2GateName, func(x ...small_rational.SmallRational) small_rational.SmallRational {
		var res small_rational.SmallRational
		res.Add(&x[0], &x[1])
		return res
	}, 2, WithUnverifiedDegree(1), WithUnverifiedSolvableVar(0)); err != nil {
		panic(err)
	}

	if err := RegisterGate(Sub2GateName, func(x ...small_rational.SmallRational) small_rational.SmallRational {
		var res small_rational.SmallRational
		res.Sub(&x[0], &x[1])
		return res
	}, 2, WithUnverifiedDegree(1), WithUnverifiedSolvableVar(0)); err != nil {
		panic(err)
	}

	if err := RegisterGate(NegGateName, func(x ...small_rational.SmallRational) small_rational.SmallRational {
		var res small_rational.SmallRational
		res.Neg(&x[0])
		return res
	}, 1, WithUnverifiedDegree(1), WithUnverifiedSolvableVar(0)); err != nil {
		panic(err)
	}

	if err := RegisterGate(Mul2GateName, func(x ...small_rational.SmallRational) small_rational.SmallRational {
		var res small_rational.SmallRational
		res.Mul(&x[0], &x[1])
		return res
	}, 2, WithUnverifiedDegree(2), WithNoSolvableVar()); err != nil {
		panic(err)
	}
}
