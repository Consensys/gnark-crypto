// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bls381

import (
	"math/bits"
)

// PairingResult target group of the pairing
type PairingResult = E12

type lineEvaluation struct {
	r0 E2
	r1 E2
	r2 E2
}

// FinalExponentiation computes the final expo x**(p**6-1)(p**2+1)(p**4 - p**2 +1)/r
func FinalExponentiation(z *PairingResult, _z ...*PairingResult) PairingResult {

	var result PairingResult
	result.Set(z)

	for _, e := range _z {
		result.Mul(&result, e)
	}

	result.FinalExponentiation(&result)

	return result
}

// FinalExponentiation sets z to the final expo x**((p**12 - 1)/r), returns z
func (z *PairingResult) FinalExponentiation(x *PairingResult) *PairingResult {

	// cf https://eprint.iacr.org/2016/130.pdf
	var result PairingResult
	result.Set(x)

	var t [6]PairingResult

	// easy part
	t[0].FrobeniusCube(&result).
		FrobeniusCube(&t[0])
	result.Inverse(&result)
	t[0].Mul(&t[0], &result)
	result.FrobeniusSquare(&t[0]).
		Mul(&result, &t[0])

	// hard part (up to permutation)
	t[0].InverseUnitary(&result).Square(&t[0])
	t[5].Expt(&result)
	t[1].CyclotomicSquare(&t[5])
	t[3].Mul(&t[0], &t[5])

	t[0].Expt(&t[3])
	t[2].Expt(&t[0])
	t[4].Expt(&t[2])

	t[4].Mul(&t[1], &t[4])
	t[1].Expt(&t[4])
	t[3].InverseUnitary(&t[3])
	t[1].Mul(&t[3], &t[1])
	t[1].Mul(&t[1], &result)

	t[0].Mul(&t[0], &result)
	t[0].FrobeniusCube(&t[0])

	t[3].InverseUnitary(&result)
	t[4].Mul(&t[3], &t[4])
	t[4].Frobenius(&t[4])

	t[5].Mul(&t[2], &t[5])
	t[5].FrobeniusSquare(&t[5])

	t[5].Mul(&t[5], &t[0])
	t[5].Mul(&t[5], &t[4])
	t[5].Mul(&t[5], &t[1])

	result.Set(&t[5])

	z.Set(&result)
	return z
}

// MillerLoop Miller loop
func MillerLoop(P G1Affine, Q G2Affine) *PairingResult {

	var result PairingResult
	result.SetOne()

	if P.IsInfinity() || Q.IsInfinity() {
		return &result
	}

	ch := make(chan struct{}, 10)

	var evaluations [68]lineEvaluation
	go preCompute(&evaluations, &Q, &P, ch)

	j := 0
	for i := len(loopCounter) - 2; i >= 0; i-- {

		result.Square(&result)
		<-ch
		result.mulAssign(&evaluations[j])
		j++

		if loopCounter[i] == 1 {
			<-ch
			result.mulAssign(&evaluations[j])
			j++
		}
	}

	return &result
}

// lineEval computes the evaluation of the line through Q, R (on the twist) at P
// Q, R are in jacobian coordinates
func lineEval(Q, R *G2Jac, P *G1Affine, result *lineEvaluation) {

	// converts _Q and _R to projective coords
	var _Q, _R G2Proj
	_Q.FromJacobian(Q)
	_R.FromJacobian(R)

	result.r1.Mul(&_Q.Y, &_R.Z)
	result.r0.Mul(&_Q.Z, &_R.X)
	result.r2.Mul(&_Q.X, &_R.Y)

	_Q.Z.Mul(&_Q.Z, &_R.Y)
	_Q.X.Mul(&_Q.X, &_R.Z)
	_Q.Y.Mul(&_Q.Y, &_R.X)

	result.r1.Sub(&result.r1, &_Q.Z)
	result.r0.Sub(&result.r0, &_Q.X)
	result.r2.Sub(&result.r2, &_Q.Y)

	result.r1.MulByElement(&result.r1, &P.X)
	result.r0.MulByElement(&result.r0, &P.Y)
}

// multiplies a result of a line evaluation to the current pairing result, taking care of mapping it
// back to the original  The line evaluation l is f(P) where div(f)=(P')+(Q')+(-P'-Q')-3(O), the support
// being on the twist.
func (z *PairingResult) mulAssign(l *lineEvaluation) *PairingResult {

	var a, b, c PairingResult
	a.MulByVWNRInv(z, &l.r1)
	b.MulByV2NRInv(z, &l.r0)
	c.MulByWNRInv(z, &l.r2)
	z.Add(&a, &b).Add(z, &c)

	return z
}

// precomputes the line evaluations used during the Miller loop.
func preCompute(evaluations *[68]lineEvaluation, Q *G2Affine, P *G1Affine, ch chan struct{}) {

	var Q1, Q2, Qbuf G2Jac
	Q1.FromAffine(Q)
	Q2.FromAffine(Q)
	Qbuf.FromAffine(Q)

	j := 0

	for i := len(loopCounter) - 2; i >= 0; i-- {

		Q1.Set(&Q2)
		Q2.Double(&Q1).Neg(&Q2)
		lineEval(&Q1, &Q2, P, &evaluations[j]) // f(P), div(f) = 2(Q1)+(-2Q2)-3(O)
		ch <- struct{}{}
		Q2.Neg(&Q2)
		j++

		if loopCounter[i] == 1 {
			lineEval(&Q2, &Qbuf, P, &evaluations[j]) // f(P), div(f) = (Q2)+(Q)+(-Q2-Q)-3(O)
			ch <- struct{}{}
			Q2.AddMixed(Q)
			j++
		}
	}
	close(ch)
}

// MulByV2NRInv set z to x*(y*v^2*(1,1)^{-1}) and return z
func (z *PairingResult) MulByV2NRInv(x *PairingResult, y *E2) *PairingResult {

	var result PairingResult
	var yNRInv E2
	yNRInv.MulByNonResidueInv(y)

	result.C0.B0.Mul(&x.C0.B1, y)
	result.C0.B1.Mul(&x.C0.B2, y)
	result.C0.B2.Mul(&x.C0.B0, &yNRInv)

	result.C1.B0.Mul(&x.C1.B1, y)
	result.C1.B1.Mul(&x.C1.B2, y)
	result.C1.B2.Mul(&x.C1.B0, &yNRInv)

	z.Set(&result)
	return z
}

// MulByVWNRInv set z to x*(y*v*w*(1,1)^{-1}) and return z
func (z *PairingResult) MulByVWNRInv(x *PairingResult, y *E2) *PairingResult {
	var result PairingResult
	var yNRInv E2
	yNRInv.MulByNonResidueInv(y)

	result.C0.B0.Mul(&x.C1.B1, y)
	result.C0.B1.Mul(&x.C1.B2, y)
	result.C0.B2.Mul(&x.C1.B0, &yNRInv)

	result.C1.B0.Mul(&x.C0.B2, y)
	result.C1.B1.Mul(&x.C0.B0, &yNRInv)
	result.C1.B2.Mul(&x.C0.B1, &yNRInv)

	z.Set(&result)
	return z
}

// MulByWNRInv set z to x*(y*w*(1,1)^{-1}) and return z
func (z *PairingResult) MulByWNRInv(x *PairingResult, y *E2) *PairingResult {

	var result PairingResult
	var yNRInv E2
	yNRInv.MulByNonResidueInv(y)

	result.C0.B0.Mul(&x.C1.B2, y)
	result.C0.B1.Mul(&x.C1.B0, &yNRInv)
	result.C0.B2.Mul(&x.C1.B1, &yNRInv)

	result.C1.B0.Mul(&x.C0.B0, &yNRInv)
	result.C1.B1.Mul(&x.C0.B1, &yNRInv)
	result.C1.B2.Mul(&x.C0.B2, &yNRInv)

	z.Set(&result)
	return z
}

// Expt set z to x^t in PairingResult and return z
func (z *PairingResult) Expt(x *PairingResult) *PairingResult {

	tAbsVal := uint64(15132376222941642752) // negative

	var result PairingResult
	result.Set(x)

	l := bits.Len64(tAbsVal) - 2
	for i := l; i >= 0; i-- {
		result.CyclotomicSquare(&result)
		if tAbsVal&(1<<uint(i)) != 0 {
			result.Mul(&result, x)
		}
	}
	result.Conjugate(&result) // because tAbsVal is negative

	z.Set(&result)
	return z
}
