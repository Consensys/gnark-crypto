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

package bw761

import (
	"github.com/consensys/gurvy/bw761/fp"
)

// GT target group of the pairing
type GT = e6

type lineEvaluation struct {
	r0 fp.Element
	r1 fp.Element
	r2 fp.Element
}

// FinalExponentiation computes the final expo x**(p**6-1)(p**2+1)(p**4 - p**2 +1)/r
func FinalExponentiation(z *GT, _z ...*GT) GT {

	var result GT
	result.Set(z)

	for _, e := range _z {
		result.Mul(&result, e)
	}

	result.FinalExponentiation(&result)

	return result
}

// FinalExponentiation sets z to the final expo x**((p**6 - 1)/r), returns z
func (z *GT) FinalExponentiation(x *GT) *GT {

	var buf GT
	var result GT
	result.Set(x)

	// easy part exponent: (p**3 - 1)*(p+1)
	buf.FrobeniusCube(&result)
	result.Inverse(&result)
	buf.Mul(&buf, &result)
	result.Frobenius(&buf).
		MulAssign(&buf)

	// hard part exponent: a multiple of (p**2 - p + 1)/r
	// Appendix B of https://eprint.iacr.org/2020/351.pdf
	// sage code: https://gitlab.inria.fr/zk-curves/bw6-761/-/blob/master/sage/pairing.py#L922
	var f [8]GT
	var fp [10]GT

	f[0].Set(&result)
	for i := 1; i < len(f); i++ {
		f[i].expt(&f[i-1])
	}
	for i := range f {
		fp[i].Frobenius(&f[i])
	}
	fp[8].expt(&fp[7])
	fp[9].expt(&fp[8])

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

	z.Set(&result)
	return z
}

// MillerLoop Miller loop
func MillerLoop(P G1, Q G2) *GT {

	var result GT

	if P.IsInfinity() || Q.IsInfinity() {
		return &result
	}

	ch := make(chan struct{}, 213)

	var evaluations1 [69]lineEvaluation
	var evaluations2 [144]lineEvaluation

	var xQjac, QjacSaved g2Jac
	xQjac.FromAffine(&Q)
	QjacSaved.FromAffine(&Q)

	// Miller loop part 1
	// computes f(P), div(f)=x(Q)-([x]Q)-(x-1)(O)
	result.SetOne()
	go preCompute1(&evaluations1, &xQjac, &P, ch)
	j := 0
	for i := len(loopCounter1) - 2; i >= 0; i-- {

		result.Square(&result)
		<-ch
		result.mulAssign(&evaluations1[j])
		j++

		if loopCounter1[i] != 0 {
			<-ch
			result.mulAssign(&evaluations1[j])
			j++
		}
	}

	// store mx=g(P), mxInv=1/g(P), div(g)=x(Q)-([x]Q)-(x-1)(O), because the second Miller loop
	// computes f(P), div(f)=(x**3-x**2-x)(Q)-([x**3-x**2-x](Q)-(x**3-x**2-x-1)(O) and
	// f(P)=g(P)**(u**2-u-1)*h(P), div(h)=(x**2-x-1)([x]Q)-([x**2-x-1][x]Q)-(x**2-x-2)(O)
	var mx, mxInv, mxplusone GT
	mx.Set(&result)
	mxInv.Inverse(&result)

	// finishes the computation of g(P), div(g)=(x+1)(Q)-([x+1]Q)-x(O) (drop the vertical line)
	var lEval lineEvaluation
	lineEval(&xQjac, &QjacSaved, &P, &lEval)
	mxplusone.Set(&mx).mulAssign(&lEval)

	// Miller loop part 2 (xQjac = [x]Q)
	// computes f(P), div(f)=(x**3-x**2-x)(Q)-([x**3-x**2-x](Q)-(x**3-x**2-x-1)(O)
	go preCompute2(&evaluations2, &xQjac, &P, ch)
	j = 0
	for i := len(loopCounter2) - 2; i >= 0; i-- {

		result.Square(&result)
		<-ch
		result.mulAssign(&evaluations2[j])
		j++

		if loopCounter2[i] == 1 {
			<-ch
			result.mulAssign(&evaluations2[j]).MulAssign(&mx) // accumulate g(P), div(g)=x(Q)-([x]Q)-(x-1)(O)
			j++
		} else if loopCounter2[i] == -1 {
			<-ch
			result.mulAssign(&evaluations2[j]).MulAssign(&mxInv) // accumulate g(P), div(g)=x(Q)-([x]Q)-(x-1)(O)
			j++
		}
	}

	close(ch)

	// g(P)*(f(P)**q)
	// div(g)=(x+1)(Q)-([x+1]Q)-x(O)
	// div(f)=(x**3-x**2-x)(Q)-([x**3-x**2-x](Q)-(x**3-x**2-x-1)(O)
	result.Frobenius(&result).MulAssign(&mxplusone)

	return &result
}

// lineEval computes the evaluation of the line through Q, R (on the twist) at P
// Q, R are in jacobian coordinates
func lineEval(Q, R *g2Jac, P *G1, result *lineEvaluation) {

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

func (z *GT) mulAssign(l *lineEvaluation) *GT {

	var a, b, c GT
	a.mulByVMinusThree(z, &l.r1)
	b.mulByVminusTwo(z, &l.r0)
	c.mulByVminusFive(z, &l.r2)
	z.Add(&a, &b).Add(z, &c)

	return z
}

// precomputes the line evaluations used during the Miller loop.
func preCompute1(evaluations *[69]lineEvaluation, Q *g2Jac, P *G1, ch chan struct{}) {

	var Q1, Qbuf g2Jac
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
func preCompute2(evaluations *[144]lineEvaluation, Q *g2Jac, P *G1, ch chan struct{}) {

	var Q1, Qbuf, Qneg g2Jac
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

// expt set z to x^t in GT and return z
func (z *GT) expt(x *GT) *GT {

	// tAbsVal in binary: 1000010100001000110000000000000000000000000000000000000000000001
	// drop the low 46 bits (all 0 except the least significant bit): 100001010000100011 = 136227
	// Shortest addition chains can be found at https://wwwhomes.uni-bielefeld.de/achim/addition_chain.html

	var result, x33 GT

	// a shortest addition chain for 136227
	result.Set(x)             // 0                1
	result.Square(&result)    // 1( 0)            2
	result.Square(&result)    // 2( 1)            4
	result.Square(&result)    // 3( 2)            8
	result.Square(&result)    // 4( 3)           16
	result.Square(&result)    // 5( 4)           32
	result.Mul(&result, x)    // 6( 5, 0)        33
	x33.Set(&result)          // save x33 for step 14
	result.Square(&result)    // 7( 6)           66
	result.Square(&result)    // 8( 7)          132
	result.Square(&result)    // 9( 8)          264
	result.Square(&result)    // 10( 9)          528
	result.Square(&result)    // 11(10)         1056
	result.Square(&result)    // 12(11)         2112
	result.Square(&result)    // 13(12)         4224
	result.Mul(&result, &x33) // 14(13, 6)      4257
	result.Square(&result)    // 15(14)         8514
	result.Square(&result)    // 16(15)        17028
	result.Square(&result)    // 17(16)        34056
	result.Square(&result)    // 18(17)        68112
	result.Mul(&result, x)    // 19(18, 0)     68113
	result.Square(&result)    // 20(19)       136226
	result.Mul(&result, x)    // 21(20, 0)    136227

	// the remaining 46 bits
	for i := 0; i < 46; i++ {
		result.Square(&result)
	}
	result.Mul(&result, x)

	z.Set(&result)
	return z
}

// mulByVMinusThree set z to x*(y*v**-3) and return z (Fp6(v) where v**3=u, v**6=-4, so v**-3 = u**-1 = (-4)**-1*u)
func (z *GT) mulByVMinusThree(x *GT, y *fp.Element) *GT {

	fourinv := fp.Element{
		8571757465769615091,
		6221412002326125864,
		16781361031322833010,
		18148962537424854844,
		6497335359600054623,
		17630955688667215145,
		15638647242705587201,
		830917065158682257,
		6848922060227959954,
		4142027113657578586,
		12050453106507568375,
		55644342162350184,
	}

	// tmp = y*(-4)**-1 * u
	var tmp e2
	tmp.A0.SetZero()
	tmp.A1.Mul(y, &fourinv)

	z.MulByE2(x, &tmp)

	return z
}

// mulByVminusTwo set z to x*(y*v**-2) and return z (Fp6(v) where v**3=u, v**6=-4, so v**-2 = (-4)**-1*u*v)
func (z *GT) mulByVminusTwo(x *GT, y *fp.Element) *GT {

	fourinv := fp.Element{
		8571757465769615091,
		6221412002326125864,
		16781361031322833010,
		18148962537424854844,
		6497335359600054623,
		17630955688667215145,
		15638647242705587201,
		830917065158682257,
		6848922060227959954,
		4142027113657578586,
		12050453106507568375,
		55644342162350184,
	}

	// tmp = y*(-4)**-1 * u
	var tmp e2
	tmp.A0.SetZero()
	tmp.A1.Mul(y, &fourinv)

	var a e2
	a.MulByElement(&x.B2, y)
	z.B2.Mul(&x.B1, &tmp)
	z.B1.Mul(&x.B0, &tmp)
	z.B0.Set(&a)

	return z
}

// mulByVminusFive set z to x*(y*v**-5) and return z (Fp6(v) where v**3=u, v**6=-4, so v**-5 = (-4)**-1*v)
func (z *GT) mulByVminusFive(x *GT, y *fp.Element) *GT {

	fourinv := fp.Element{
		8571757465769615091,
		6221412002326125864,
		16781361031322833010,
		18148962537424854844,
		6497335359600054623,
		17630955688667215145,
		15638647242705587201,
		830917065158682257,
		6848922060227959954,
		4142027113657578586,
		12050453106507568375,
		55644342162350184,
	}

	// tmp = y*(-4)**-1 * u
	var tmp e2
	tmp.A0.SetZero()
	tmp.A1.Mul(y, &fourinv)

	var a e2
	a.Mul(&x.B2, &tmp)
	z.B2.MulByElement(&x.B1, &tmp.A1)
	z.B1.MulByElement(&x.B0, &tmp.A1)
	z.B0.Set(&a)

	return z
}
