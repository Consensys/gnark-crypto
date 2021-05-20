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

package bw6761

import (
	"errors"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/internal/fptower"
)

// GT target group of the pairing
type GT = fptower.E6

type lineEvaluation struct {
	r0 fp.Element
	r1 fp.Element
	r2 fp.Element
}

// Pair calculates the reduced pairing for a set of points
func Pair(P []G1Affine, Q []G2Affine) (GT, error) {
	f, err := MillerLoop(P, Q)
	if err != nil {
		return GT{}, err
	}
	return FinalExponentiation(&f), nil
}

// PairingCheck calculates the reduced pairing for a set of points and returns True if the result is One
func PairingCheck(P []G1Affine, Q []G2Affine) (bool, error) {
	f, err := Pair(P, Q)
	if err != nil {
		return false, err
	}
	var one GT
	one.SetOne()
	return f.Equal(&one), nil
}

// FinalExponentiation computes the final expo x**(c*(p**3-1)(p+1)(p**2-p+1)/r)
func FinalExponentiation(z *GT, _z ...*GT) GT {

	var result GT
	result.Set(z)

	for _, e := range _z {
		result.Mul(&result, e)
	}

	var buf GT

	// easy part exponent: (p**3 - 1)*(p+1)
	buf.Conjugate(&result)
	result.Inverse(&result)
	buf.Mul(&buf, &result)
	result.Frobenius(&buf).
		MulAssign(&buf)

		// hard part exponent: 12(u+1)(p**2 - p + 1)/r
	var m1, _m1, m2, _m2, m3, f0, f0_36, g0, g1, _g1, g2, g3, _g3, g4, _g4, g5, _g5, g6, gA, gB, g034, _g1g2, gC, h1, h2, h2g2C, h4 GT
	m1.Expt(&result)
	_m1.Conjugate(&m1)
	m2.Expt(&m1)
	_m2.Conjugate(&m2)
	m3.Expt(&m2)
	f0.Frobenius(&result).
		Mul(&f0, &result).
		Mul(&f0, &m2)
	m2.CyclotomicSquare(&_m1)
	f0.Mul(&f0, &m2)
	f0_36.CyclotomicSquare(&f0).
		CyclotomicSquare(&f0_36).
		CyclotomicSquare(&f0_36).
		Mul(&f0_36, &f0).
		CyclotomicSquare(&f0_36).
		CyclotomicSquare(&f0_36)
	g0.Mul(&result, &m1).
		Frobenius(&g0).
		Mul(&g0, &m3).
		Mul(&g0, &_m2).
		Mul(&g0, &_m1)
	g1.Expt(&g0)
	_g1.Conjugate(&g1)
	g2.Expt(&g1)
	g3.Expt(&g2)
	_g3.Conjugate(&g3)
	g4.Expt(&g3)
	_g4.Conjugate(&g4)
	g5.Expt(&g4)
	_g5.Conjugate(&g5)
	g6.Expt(&g5)
	gA.Mul(&g3, &_g5).
		CyclotomicSquare(&gA).
		Mul(&gA, &g6).
		Mul(&gA, &g1).
		Mul(&gA, &g0)
	g034.Mul(&g0, &g3).
		Mul(&g034, &_g4)
	gB.CyclotomicSquare(&g034).
		Mul(&gB, &g034).
		Mul(&gB, &g5).
		Mul(&gB, &_g1)
	_g1g2.Mul(&_g1, &g2)
	gC.Mul(&_g3, &_g1g2).
		CyclotomicSquare(&gC).
		Mul(&gC, &_g1g2).
		Mul(&gC, &g0).
		CyclotomicSquare(&gC).
		Mul(&gC, &g2).
		Mul(&gC, &g0).
		Mul(&gC, &g4)
		// ht, hy = 13, 9
		// c1 = ht**2+3*hy**2 = 412
	h1.Expc1(&gA)
	// c2 = ht+hy = 22
	h2.Expc2(&gB)
	h2g2C.CyclotomicSquare(&gC).
		Mul(&h2g2C, &h2)
	h4.CyclotomicSquare(&h2g2C).
		Mul(&h4, &h2g2C).
		CyclotomicSquare(&h4)
	result.Mul(&h1, &h4).
		Mul(&result, &f0_36)

	return result
}

var lineEvalPool1 = sync.Pool{
	New: func() interface{} {
		return new([69]lineEvaluation)
	},
}

var lineEvalPool2 = sync.Pool{
	New: func() interface{} {
		return new([144]lineEvaluation)
	},
}

// MillerLoop Miller loop
func MillerLoop(P []G1Affine, Q []G2Affine) (GT, error) {
	nP := len(P)
	if nP == 0 || nP != len(Q) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	var (
		ch1 = make([]chan struct{}, 0, nP)
		ch2 = make([]chan struct{}, 0, nP)

		evaluations1 = make([]*[69]lineEvaluation, 0, nP)
		evaluations2 = make([]*[144]lineEvaluation, 0, nP)
		Paff         = make([]G1Affine, nP)
		xQjac        = make([]G2Jac, nP)
		QjacSaved    = make([]G2Jac, nP)
	)

	var countInf = 0
	for k := 0; k < nP; k++ {
		if P[k].IsInfinity() || Q[k].IsInfinity() {
			countInf++
			continue
		}

		xQjac[k-countInf].FromAffine(&Q[k])
		QjacSaved[k-countInf].FromAffine(&Q[k])
		Paff[k-countInf].Set(&P[k])
		ch1 = append(ch1, make(chan struct{}, 10))
		evaluations1 = append(evaluations1, lineEvalPool1.Get().(*[69]lineEvaluation))

		go preCompute1(evaluations1[k-countInf], &xQjac[k-countInf], &Paff[k-countInf], ch1[k-countInf])
	}

	nP = nP - countInf

	// Miller loop part 1
	// computes f(P), div(f)=x(Q)-([x]Q)-(x-1)(O)
	var result GT
	result.SetOne()
	j := 0
	for i := len(loopCounter1) - 2; i >= 0; i-- {

		result.Square(&result)
		for k := 0; k < nP; k++ {
			<-ch1[k]
			mulAssign(&result, &evaluations1[k][j])
		}
		j++

		if loopCounter1[i] != 0 {
			for k := 0; k < nP; k++ {
				<-ch1[k]
				mulAssign(&result, &evaluations1[k][j])
			}
			j++
		}
	}

	// release objects into the pool
	for i := 0; i < len(evaluations1); i++ {
		lineEvalPool1.Put(evaluations1[i])
	}

	// store mx=g(P), mxInv=1/g(P), div(g)=x(Q)-([x]Q)-(x-1)(O), because the second Miller loop
	// computes f(P), div(f)=(x**3-x**2-x)(Q)-([x**3-x**2-x](Q)-(x**3-x**2-x-1)(O) and
	// f(P)=g(P)**(u**2-u-1)*h(P), div(h)=(x**2-x-1)([x]Q)-([x**2-x-1][x]Q)-(x**2-x-2)(O)
	var mx, mxInv, mxplusone GT
	mx.Set(&result)
	mxInv.Conjugate(&result)
	mxplusone.Set(&mx)

	var lEval = make([]lineEvaluation, nP)

	for k := 0; k < nP; k++ {
		// finishes the computation of g(P), div(g)=(x+1)(Q)-([x+1]Q)-x(O) (drop the vertical line)
		lineEval(&xQjac[k], &QjacSaved[k], &Paff[k], &lEval[k])
		mulAssign(&mxplusone, &lEval[k])

		ch2 = append(ch2, make(chan struct{}, 10))
		evaluations2 = append(evaluations2, lineEvalPool2.Get().(*[144]lineEvaluation))

		// Miller loop part 2 (xQjac = [x]Q)
		// computes f(P), div(f)=(x**3-x**2-x)(Q)-([x**3-x**2-x](Q)-(x**3-x**2-x-1)(O)
		go preCompute2(evaluations2[k], &xQjac[k], &Paff[k], ch2[k])
	}

	j = 0
	for i := len(loopCounter2) - 2; i >= 0; i-- {

		result.Square(&result)
		for k := 0; k < nP; k++ {
			<-ch2[k]
			mulAssign(&result, &evaluations2[k][j])
		}
		j++

		if loopCounter2[i] == 1 {
			for k := 0; k < nP; k++ {
				<-ch2[k]
				mulAssign(&result, &evaluations2[k][j]) // accumulate g(P), div(g)=x(Q)-([x]Q)-(x-1)(O)
			}
			result.MulAssign(&mx)
			j++
		} else if loopCounter2[i] == -1 {
			for k := 0; k < nP; k++ {
				<-ch2[k]
				mulAssign(&result, &evaluations2[k][j]) // accumulate g(P), div(g)=x(Q)-([x]Q)-(x-1)(O)
			}
			result.MulAssign(&mxInv)
			j++
		}
	}

	// release objects into the pool
	for i := 0; i < len(evaluations2); i++ {
		lineEvalPool2.Put(evaluations2[i])
	}

	// g(P)*(f(P)**q)
	// div(g)=(x+1)(Q)-([x+1]Q)-x(O)
	// div(f)=(x**3-x**2-x)(Q)-([x**3-x**2-x](Q)-(x**3-x**2-x-1)(O)
	result.Frobenius(&result).MulAssign(&mxplusone)

	return result, nil
}

// lineEval computes the evaluation of the line through Q, R (on the twist) at P
// Q, R are in jacobian coordinates
func lineEval(Q, R *G2Jac, P *G1Affine, result *lineEvaluation) {

	// converts _Q and _R to projective coords
	var _Q, _R g2Proj
	_Q.FromJacobian(Q)
	_R.FromJacobian(R)

	result.r1.Mul(&_Q.y, &_R.z)
	result.r0.Mul(&_Q.z, &_R.x)
	result.r2.Mul(&_Q.x, &_R.y)

	_Q.z.Mul(&_Q.z, &_R.y)
	_Q.x.Mul(&_Q.x, &_R.z)
	_Q.y.Mul(&_Q.y, &_R.x)

	result.r1.Sub(&result.r1, &_Q.z)
	result.r0.Sub(&result.r0, &_Q.x)
	result.r2.Sub(&result.r2, &_Q.y)

	result.r1.Mul(&result.r1, &P.X)
	result.r0.Mul(&result.r0, &P.Y)
}

func mulAssign(z *GT, l *lineEvaluation) *GT {

	var a, b, c GT
	a.MulByVMinusThree(z, &l.r1)
	b.MulByVminusTwo(z, &l.r0)
	c.MulByVminusFive(z, &l.r2)
	z.Add(&a, &b).Add(z, &c)

	return z
}

// precomputes the line evaluations used during the Miller loop.
func preCompute1(evaluations *[69]lineEvaluation, Q *G2Jac, P *G1Affine, ch chan struct{}) {

	var Q1, Qbuf G2Jac
	Q1.Set(Q)
	Qbuf.Set(Q)

	j := 0

	for i := len(loopCounter1) - 2; i >= 0; i-- {

		Q1.Set(Q)
		Q.Double(&Q1).Neg(Q)
		lineEval(&Q1, Q, P, &evaluations[j]) // f(P), div(f) = 2(Q1)+(-2Q)-3(O)
		Q.Neg(Q)
		ch <- struct{}{}
		j++

		if loopCounter1[i] == 1 {
			lineEval(Q, &Qbuf, P, &evaluations[j]) // f(P), div(f) = (Q)+(Qbuf)+(-Q-Qbuf)-3(O)
			Q.AddAssign(&Qbuf)
			ch <- struct{}{}
			j++
		}
	}

}

// precomputes the line evaluations used during the Miller loop.
func preCompute2(evaluations *[144]lineEvaluation, Q *G2Jac, P *G1Affine, ch chan struct{}) {

	var Q1, Qbuf, Qneg G2Jac
	Q1.Set(Q)
	Qbuf.Set(Q)
	Qneg.Neg(Q)

	j := 0

	for i := len(loopCounter2) - 2; i >= 0; i-- {

		Q1.Set(Q)
		Q.Double(&Q1).Neg(Q)
		lineEval(&Q1, Q, P, &evaluations[j]) // f(P), div(f) = 2(Q1)+(-2Q)-3(O)
		Q.Neg(Q)
		ch <- struct{}{}
		j++

		if loopCounter2[i] == 1 {
			lineEval(Q, &Qbuf, P, &evaluations[j]) // f(P), div(f) = (Q)+(Qbuf)+(-Q-Qbuf)-3(O)
			Q.AddAssign(&Qbuf)
			ch <- struct{}{}
			j++
		} else if loopCounter2[i] == -1 {
			lineEval(Q, &Qneg, P, &evaluations[j]) // f(P), div(f) = (Q)+(-Qbuf)+(-Q+Qbuf)-3(O)
			Q.AddAssign(&Qneg)
			ch <- struct{}{}
			j++
		}
	}
}
