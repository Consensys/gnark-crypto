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
	buf.FrobeniusCube(&result)
	result.Inverse(&result)
	buf.Mul(&buf, &result)
	result.Frobenius(&buf).
		MulAssign(&buf)

	// hard part exponent: a multiple of (p**2 - p + 1)/r
	// cofactor in exponent = 3(x**3-x**2+1)
	// Appendix B of https://eprint.iacr.org/2020/351.pdf
	// sage code: https://gitlab.inria.fr/zk-curves/bw6-761/-/blob/master/sage/pairing.py#L922
	var f [8]GT
	var fp [10]GT

	f[0].Set(&result)
	for i := 1; i < len(f); i++ {
		f[i].Expt(&f[i-1])
	}
	for i := range f {
		fp[i].Frobenius(&f[i])
	}
	fp[8].Expt(&fp[7])
	fp[9].Expt(&fp[8])

	result.FrobeniusCube(&fp[5]).
		MulAssign(&fp[3]).
		MulAssign(&fp[6]).
		CyclotomicSquare(&result)

	var f4fp2 GT
	f4fp2.Mul(&f[4], &fp[2])
	buf.Mul(&f[0], &f[1]).
		MulAssign(&f[3]).
		MulAssign(&f4fp2).
		MulAssign(&fp[8])
	buf.FrobeniusCube(&buf)
	result.MulAssign(&buf)

	result.MulAssign(&f[5]).
		MulAssign(&fp[0]).
		CyclotomicSquare(&result)

	buf.FrobeniusCube(&f[7])
	result.MulAssign(&buf)

	result.MulAssign(&fp[9]).
		CyclotomicSquare(&result)

	var f2fp4, f4fp2fp5 GT
	f2fp4.Mul(&f[2], &fp[4])
	f4fp2fp5.Mul(&f4fp2, &fp[5])
	buf.Mul(&f2fp4, &f[3]).
		MulAssign(&fp[3])
	buf.FrobeniusCube(&buf)
	result.MulAssign(&buf)

	result.MulAssign(&f4fp2fp5).
		MulAssign(&f[6]).
		MulAssign(&fp[7]).
		CyclotomicSquare(&result)

	buf.Mul(&fp[0], &fp[9])
	buf.FrobeniusCube(&buf)
	result.MulAssign(&buf)
	result.MulAssign(&f[0]).
		MulAssign(&f[7]).
		MulAssign(&fp[1]).
		CyclotomicSquare(&result)

	var fp6fp8, f5fp7 GT
	fp6fp8.Mul(&fp[6], &fp[8])
	f5fp7.Mul(&f[5], &fp[7])
	buf.FrobeniusCube(&fp6fp8)
	result.MulAssign(&buf)

	result.MulAssign(&f5fp7).
		MulAssign(&fp[2]).
		CyclotomicSquare(&result)

	var f3f6, f1f7 GT
	f3f6.Mul(&f[3], &f[6])
	f1f7.Mul(&f[1], &f[7])

	buf.Mul(&f1f7, &f[2])
	buf.FrobeniusCube(&buf)
	result.MulAssign(&buf)

	result.MulAssign(&f3f6).
		MulAssign(&fp[9]).
		CyclotomicSquare(&result)

	buf.Mul(&f4fp2, &f5fp7).
		MulAssign(&fp6fp8)
	buf.FrobeniusCube(&buf)
	result.MulAssign(&buf)

	result.MulAssign(&f[0]).
		MulAssign(&fp[0]).
		MulAssign(&fp[3]).
		MulAssign(&fp[5]).
		CyclotomicSquare(&result)

	buf.FrobeniusCube(&f3f6)
	result.MulAssign(&buf)

	result.MulAssign(&fp[1]).
		CyclotomicSquare(&result)

	buf.Mul(&f2fp4, &f4fp2fp5).MulAssign(&fp[9])
	buf.FrobeniusCube(&buf)
	result.MulAssign(&buf)

	result.MulAssign(&f1f7).MulAssign(&f5fp7).MulAssign(&fp[0])

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
	mxInv.Inverse(&result)
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
