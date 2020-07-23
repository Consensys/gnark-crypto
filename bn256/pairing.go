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

package bn256

import "math/bits"

// FinalExponentiation multiplies the inputs, applies the final expo on the result and return it
func (curve *Curve) FinalExponentiation(z *PairingResult, _z ...*PairingResult) PairingResult {

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

	// For BN curves use Section 5 of https://eprint.iacr.org/2008/490.pdf
	var mt [4]PairingResult // mt[i] is m^(t^i)

	// easy part
	mt[0].Set(x)
	var temp PairingResult
	temp.FrobeniusCube(&mt[0]).
		FrobeniusCube(&temp)
	mt[0].Inverse(&mt[0])
	temp.Mul(&temp, &mt[0])
	mt[0].FrobeniusSquare(&temp).
		Mul(&mt[0], &temp)

	// hard part
	mt[1].Expt(&mt[0])
	mt[2].Expt(&mt[1])
	mt[3].Expt(&mt[2])

	var y [7]PairingResult

	y[1].InverseUnitary(&mt[0])
	y[4].Set(&mt[1])
	y[5].InverseUnitary(&mt[2])
	y[6].Set(&mt[3])

	mt[0].Frobenius(&mt[0])
	mt[1].Frobenius(&mt[1])
	mt[2].Frobenius(&mt[2])
	mt[3].Frobenius(&mt[3])

	y[0].Set(&mt[0])
	y[3].InverseUnitary(&mt[1])
	y[4].Mul(&y[4], &mt[2]).InverseUnitary(&y[4])
	y[6].Mul(&y[6], &mt[3]).InverseUnitary(&y[6])

	mt[0].Frobenius(&mt[0])
	mt[2].Frobenius(&mt[2])

	y[0].Mul(&y[0], &mt[0])
	y[2].Set(&mt[2])

	mt[0].Frobenius(&mt[0])

	y[0].Mul(&y[0], &mt[0])

	// compute addition chain
	var t [2]PairingResult

	t[0].CyclotomicSquare(&y[6])
	t[0].Mul(&t[0], &y[4])
	t[0].Mul(&t[0], &y[5])
	t[1].Mul(&y[3], &y[5])
	t[1].Mul(&t[1], &t[0])
	t[0].Mul(&t[0], &y[2])
	t[1].CyclotomicSquare(&t[1])
	t[1].Mul(&t[1], &t[0])
	t[1].CyclotomicSquare(&t[1])
	t[0].Mul(&t[1], &y[1])
	t[1].Mul(&t[1], &y[0])
	t[0].CyclotomicSquare(&t[0])
	z.Mul(&t[0], &t[1])
	return z
}

// MillerLoop Miller loop
func (curve *Curve) MillerLoop(P G1Affine, Q G2Affine, result *PairingResult) *PairingResult {

	// init result
	result.SetOne()

	if P.IsInfinity() || Q.IsInfinity() {
		return result
	}

	// the line goes through QCur and QNext
	var QCur, QNext, QNextNeg G2Jac
	var QNeg G2Affine

	// Stores -Q
	QNeg.Neg(&Q)

	// init QCur with Q
	QCur.FromAffine(&Q)

	var lEval lineEvalRes

	// Miller loop
	for i := len(curve.loopCounter) - 2; i >= 0; i-- {
		QNext.Set(&QCur)
		QNext.DoubleAssign()
		QNextNeg.Neg(&QNext)

		result.Square(result)

		// evaluates line though Qcur,2Qcur at P
		lineEvalJac(QCur, QNextNeg, &P, &lEval)
		lEval.mulAssign(result)

		if curve.loopCounter[i] == 1 {
			// evaluates line through 2Qcur, Q at P
			lineEvalAffine(QNext, Q, &P, &lEval)
			lEval.mulAssign(result)
			QNext.AddMixed(&Q)

		} else if curve.loopCounter[i] == -1 {
			// evaluates line through 2Qcur, -Q at P
			lineEvalAffine(QNext, QNeg, &P, &lEval)
			lEval.mulAssign(result)
			QNext.AddMixed(&QNeg)
		}
		QCur.Set(&QNext)
	}

	// cf https://eprint.iacr.org/2010/354.pdf for instance for optimal Ate Pairing
	var Q1, Q2 G2Affine

	//Q1 = Frob(Q)
	Q1.X.Conjugate(&Q.X).MulByNonResidue1Power2(&Q1.X)
	Q1.Y.Conjugate(&Q.Y).MulByNonResidue1Power3(&Q1.Y)

	// Q2 = -Frob2(Q)
	Q2.X.MulByNonResidue2Power2(&Q.X)
	Q2.Y.MulByNonResidue2Power3(&Q.Y).Neg(&Q2.Y)

	lineEvalAffine(QCur, Q1, &P, &lEval)
	lEval.mulAssign(result)

	QCur.AddMixed(&Q1)

	lineEvalAffine(QCur, Q2, &P, &lEval)
	lEval.mulAssign(result)

	return result
}

// lineEval computes the evaluation of the line through Q, R (on the twist) at P
// Q, R are in jacobian coordinates
// The case in which Q=R=Infinity is not handled as this doesn't happen in the SNARK pairing
func lineEvalJac(Q, R G2Jac, P *G1Affine, result *lineEvalRes) {
	// converts _Q and _R to projective coords
	var _Q, _R G2Proj
	_Q.FromJacobian(&Q)
	_R.FromJacobian(&R)

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

// Same as above but R is in affine coords
func lineEvalAffine(Q G2Jac, R G2Affine, P *G1Affine, result *lineEvalRes) {

	// converts Q and R to projective coords
	var _Q G2Proj
	_Q.FromJacobian(&Q)

	result.r1.Set(&_Q.Y)
	result.r0.Mul(&_Q.Z, &R.X)
	result.r2.Mul(&_Q.X, &R.Y)

	_Q.Z.Mul(&_Q.Z, &R.Y)
	_Q.Y.Mul(&_Q.Y, &R.X)

	result.r1.Sub(&result.r1, &_Q.Z)
	result.r0.Sub(&result.r0, &_Q.X)
	result.r2.Sub(&result.r2, &_Q.Y)

	result.r1.MulByElement(&result.r1, &P.X)
	result.r0.MulByElement(&result.r0, &P.Y)
}

type lineEvalRes struct {
	r0 G2CoordType // c0.b1
	r1 G2CoordType // c1.b1
	r2 G2CoordType // c1.b2
}

func (l *lineEvalRes) mulAssign(z *PairingResult) *PairingResult {

	var a, b, c PairingResult
	a.MulByVW(z, &l.r1)
	b.MulByV(z, &l.r0)
	c.MulByV2W(z, &l.r2)
	z.Add(&a, &b).Add(z, &c)

	return z
}

// MulByVW set z to x*(y*v*w) and return z
// here y*v*w means the PairingResult element with C1.B1=y and all other components 0
func (z *PairingResult) MulByVW(x *PairingResult, y *G2CoordType) *PairingResult {

	var result PairingResult
	var yNR G2CoordType

	yNR.MulByNonResidue(y)
	result.C0.B0.Mul(&x.C1.B1, &yNR)
	result.C0.B1.Mul(&x.C1.B2, &yNR)
	result.C0.B2.Mul(&x.C1.B0, y)
	result.C1.B0.Mul(&x.C0.B2, &yNR)
	result.C1.B1.Mul(&x.C0.B0, y)
	result.C1.B2.Mul(&x.C0.B1, y)
	z.Set(&result)
	return z
}

// MulByV set z to x*(y*v) and return z
// here y*v means the PairingResult element with C0.B1=y and all other components 0
func (z *PairingResult) MulByV(x *PairingResult, y *G2CoordType) *PairingResult {

	var result PairingResult
	var yNR G2CoordType

	yNR.MulByNonResidue(y)
	result.C0.B0.Mul(&x.C0.B2, &yNR)
	result.C0.B1.Mul(&x.C0.B0, y)
	result.C0.B2.Mul(&x.C0.B1, y)
	result.C1.B0.Mul(&x.C1.B2, &yNR)
	result.C1.B1.Mul(&x.C1.B0, y)
	result.C1.B2.Mul(&x.C1.B1, y)
	z.Set(&result)
	return z
}

// MulByV2W set z to x*(y*v^2*w) and return z
// here y*v^2*w means the PairingResult element with C1.B2=y and all other components 0
func (z *PairingResult) MulByV2W(x *PairingResult, y *G2CoordType) *PairingResult {

	var result PairingResult
	var yNR G2CoordType

	yNR.MulByNonResidue(y)
	result.C0.B0.Mul(&x.C1.B0, &yNR)
	result.C0.B1.Mul(&x.C1.B1, &yNR)
	result.C0.B2.Mul(&x.C1.B2, &yNR)
	result.C1.B0.Mul(&x.C0.B1, &yNR)
	result.C1.B1.Mul(&x.C0.B2, &yNR)
	result.C1.B2.Mul(&x.C0.B0, y)
	z.Set(&result)
	return z
}

// Expt set z to x^t in PairingResult and return z
func (z *PairingResult) Expt(x *PairingResult) *PairingResult {

	var tAbsVal uint64
	tAbsVal = 4965661367192848881

	var result PairingResult
	result.Set(x)

	l := bits.Len64(tAbsVal) - 2
	for i := l; i >= 0; i-- {
		result.CyclotomicSquare(&result)
		if tAbsVal&(1<<uint(i)) != 0 {
			result.Mul(&result, x)
		}
	}

	z.Set(&result)
	return z
}
