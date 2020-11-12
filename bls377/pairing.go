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

package bls377

// GT target group of the pairing
type GT = e12

type lineEvaluation struct {
	r0 e2
	r1 e2
	r2 e2
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

// FinalExponentiation sets z to the final expo x**((p**12 - 1)/r), returns z
func (z *GT) FinalExponentiation(x *GT) *GT {

	// https://eprint.iacr.org/2016/130.pdf
	var result GT
	result.Set(x)

	// memalloc
	var t [6]GT

	// easy part
	t[0].Conjugate(&result)
	result.Inverse(&result)
	t[0].Mul(&t[0], &result)
	result.FrobeniusSquare(&t[0]).
		Mul(&result, &t[0])

	// hard part (up to permutation)
	t[0].InverseUnitary(&result).Square(&t[0])
	t[5].expt(&result)
	t[1].CyclotomicSquare(&t[5])
	t[3].Mul(&t[0], &t[5])

	t[0].expt(&t[3])
	t[2].expt(&t[0])
	t[4].expt(&t[2])

	t[4].Mul(&t[1], &t[4])
	t[1].expt(&t[4])
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
func MillerLoop(P G1, Q G2) *GT {

	var result GT
	result.SetOne()

	if P.IsInfinity() || Q.IsInfinity() {
		return &result
	}

	ch := make(chan struct{}, 20)

	var evaluations [69]lineEvaluation
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

	result.r1.MulByElement(&result.r1, &P.X)
	result.r0.MulByElement(&result.r0, &P.Y)
}

func (z *GT) mulAssign(l *lineEvaluation) *GT {

	var a, b, c GT
	a.MulByVW(z, &l.r1)
	b.MulByV(z, &l.r0)
	c.MulByV2W(z, &l.r2)
	z.Add(&a, &b).Add(z, &c)

	return z
}

// precomputes the line evaluations used during the Miller loop.
func preCompute(evaluations *[69]lineEvaluation, Q *G2, P *G1, ch chan struct{}) {

	var Q1, Q2, Qbuf g2Jac
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

// MulByVW set z to x*(y*v*w) and return z
// here y*v*w means the GT element with C1.B1=y and all other components 0
func (z *GT) MulByVW(x *GT, y *e2) *GT {

	var result GT
	var yNR e2

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
// here y*v means the GT element with C0.B1=y and all other components 0
func (z *GT) MulByV(x *GT, y *e2) *GT {

	var result GT
	var yNR e2

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
// here y*v^2*w means the GT element with C1.B2=y and all other components 0
func (z *GT) MulByV2W(x *GT, y *e2) *GT {

	var result GT
	var yNR e2

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

// expt set z to x^t in GT and return z
func (z *GT) expt(x *GT) *GT {
	// const tAbsVal uint64 = 9586122913090633729
	// tAbsVal in binary: 1000010100001000110000000000000000000000000000000000000000000001
	// drop the low 46 bits (all 0 except the least significant bit): 100001010000100011 = 136227
	// Shortest addition chains can be found at https://wwwhomes.uni-bielefeld.de/achim/addition_chain.html

	var result, x33 GT

	// a shortest addition chain for 136227
	result.Set(x)                    // 0                1
	result.CyclotomicSquare(&result) // 1( 0)            2
	result.CyclotomicSquare(&result) // 2( 1)            4
	result.CyclotomicSquare(&result) // 3( 2)            8
	result.CyclotomicSquare(&result) // 4( 3)           16
	result.CyclotomicSquare(&result) // 5( 4)           32
	result.Mul(&result, x)           // 6( 5, 0)        33
	x33.Set(&result)                 // save x33 for step 14
	result.CyclotomicSquare(&result) // 7( 6)           66
	result.CyclotomicSquare(&result) // 8( 7)          132
	result.CyclotomicSquare(&result) // 9( 8)          264
	result.CyclotomicSquare(&result) // 10( 9)          528
	result.CyclotomicSquare(&result) // 11(10)         1056
	result.CyclotomicSquare(&result) // 12(11)         2112
	result.CyclotomicSquare(&result) // 13(12)         4224
	result.Mul(&result, &x33)        // 14(13, 6)      4257
	result.CyclotomicSquare(&result) // 15(14)         8514
	result.CyclotomicSquare(&result) // 16(15)        17028
	result.CyclotomicSquare(&result) // 17(16)        34056
	result.CyclotomicSquare(&result) // 18(17)        68112
	result.Mul(&result, x)           // 19(18, 0)     68113
	result.CyclotomicSquare(&result) // 20(19)       136226
	result.Mul(&result, x)           // 21(20, 0)    136227

	// the remaining 46 bits
	for i := 0; i < 46; i++ {
		result.CyclotomicSquare(&result)
	}
	result.Mul(&result, x)

	z.Set(&result)
	return z
}
