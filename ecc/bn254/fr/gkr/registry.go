// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package gkr

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"slices"
	"sync"
)

var (
	gates     = make(map[string]*Gate)
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
func isAdditive(f GateFunction, i, nbIn int) bool {
	// fix all variables except the i-th one at random points
	// pick random value x1 for the i-th variable
	// check if f(-, 0, -) + f(-, 2*x1, -) = 2*f(-, x1, -)
	x := make(fr.Vector, nbIn)
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
func fitPoly(f GateFunction, nbIn int, degreeBound uint64) polynomial.Polynomial {
	// turn f univariate by defining p(x) as f(x, x, ..., x)
	fIn := make([]fr.Element, nbIn)
	p := make(polynomial.Polynomial, degreeBound)
	domain := fft.NewDomain(degreeBound)
	// evaluate p on the unit circle (first filling p with evaluations rather than coefficients)
	x := fr.One()
	for i := range p {
		for j := range fIn {
			fIn[j] = x
		}
		p[i] = f(fIn...)

		x.Mul(&x, &domain.Generator)
	}

	// obtain p's coefficients
	domain.FFTInverse(p, fft.DIF)
	fft.BitReverse(p)

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

// RegisterGate creates a gate object and stores it in the gates registry
// name is a human-readable name for the gate
// f is the polynomial function defining the gate
// nbIn is the number of inputs to the gate
func RegisterGate(name string, f GateFunction, nbIn int, options ...RegisterGateOption) error {
	s := registerGateSettings{degree: -1, solvableVar: -1}
	for _, option := range options {
		option(&s)
	}

	if s.degree == -1 { // find a degree
		if s.noDegreeVerification {
			panic("invalid settings")
		}
		found := false
		const maxAutoDegreeBound = 32
		for degreeBound := uint64(4); degreeBound <= maxAutoDegreeBound; degreeBound *= 2 {
			if p := fitPoly(f, nbIn, degreeBound); p != nil {
				found = true
				s.degree = len(p) - 1
				break
			}
		}
		if !found {
			return fmt.Errorf("could not find a degree for gate %s: tried up to %d", name, maxAutoDegreeBound-1)
		}
	} else {
		if !s.noDegreeVerification { // check that the given degree is correct
			if p := fitPoly(f, nbIn, ecc.NextPowerOfTwo(uint64(s.degree)+1)); p == nil {
				return fmt.Errorf("detected a higher degree than %d for gate %s", s.degree, name)
			} else if len(p)-1 != s.degree {
				return fmt.Errorf("detected degree %d for gate %s, claimed %d", len(p)-1, name, s.degree)
			}
		}
	}

	if s.solvableVar == -1 {
		if !s.noSolvableVarVerification { // find a solvable variable
			for i := range nbIn {
				if isAdditive(f, i, nbIn) {
					s.solvableVar = i
					break
				}
			}
		}
	} else {
		// solvable variable given
		if !s.noSolvableVarVerification && !isAdditive(f, s.solvableVar, nbIn) {
			return fmt.Errorf("cannot verify the solvability of variable %d in gate %s", s.solvableVar, name)
		}
	}

	gatesLock.Lock()
	defer gatesLock.Unlock()
	gates[name] = &Gate{Evaluate: f, nbIn: nbIn, degree: s.degree, solvableVar: s.solvableVar}
	return nil
}

func GetGate(name string) *Gate {
	gatesLock.Lock()
	defer gatesLock.Unlock()
	return gates[name]
}

func RemoveGate(name string) bool {
	gatesLock.Lock()
	defer gatesLock.Unlock()
	_, found := gates[name]
	if found {
		delete(gates, name)
	}
	return found
}
