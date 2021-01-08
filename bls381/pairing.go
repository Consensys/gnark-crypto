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
	"errors"
	"sync"

	"github.com/consensys/gurvy/bls381/internal/fptower"
)

// GT target group of the pairing
type GT = fptower.E12

type lineEvaluation struct {
	r0 fptower.E2
	r1 fptower.E2
	r2 fptower.E2
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

// FinalExponentiation computes the final expo x**(p**6-1)(p**2+1)(p**4 - p**2 +1)/r
func FinalExponentiation(z *GT, _z ...*GT) GT {
	var result GT
	result.Set(z)

	for _, e := range _z {
		result.Mul(&result, e)
	}

	var t [4]GT

	// easy part
	t[0].Conjugate(&result)
	result.Inverse(&result)
	t[0].Mul(&t[0], &result)
	result.FrobeniusSquare(&t[0]).
		Mul(&result, &t[0])

	// hard part (up to permutation)
	// Alg.2 from https://eprint.iacr.org/2016/130.pdf
	t[0].CyclotomicSquare(&result)
	t[1].Expt(&t[0])
	t[2].ExptHalf(&t[1])
	t[3].InverseUnitary(&result)
	t[1].Mul(&t[1], &t[3])
	t[1].InverseUnitary(&t[1])
	t[1].Mul(&t[1], &t[2])
	t[2].Expt(&t[1])
	t[3].Expt(&t[2])
	t[1].InverseUnitary(&t[1])
	t[3].Mul(&t[1], &t[3])
	t[1].InverseUnitary(&t[1])
	t[1].FrobeniusCube(&t[1])
	t[2].FrobeniusSquare(&t[2])
	t[1].Mul(&t[1], &t[2])
	t[2].Expt(&t[3])
	t[2].Mul(&t[2], &t[0])
	t[2].Mul(&t[2], &result)
	t[1].Mul(&t[1], &t[2])
	t[2].Frobenius(&t[3])
	t[1].Mul(&t[1], &t[2])

	result.Set(&t[1])

	return result
}

var lineEvalPool = sync.Pool{
	New: func() interface{} {
		return new([68]lineEvaluation)
	},
}

// MillerLoop Miller loop
func MillerLoop(P []G1Affine, Q []G2Affine) (GT, error) {
	nP := len(P)
	if nP == 0 || nP != len(Q) {
		return GT{}, errors.New("invalid inputs sizes")
	}

	var (
		ch          = make([]chan struct{}, 0, nP)
		evaluations = make([]*[68]lineEvaluation, 0, nP)
	)

	var countInf = 0
	for k := 0; k < nP; k++ {
		if P[k].IsInfinity() || Q[k].IsInfinity() {
			countInf++
			continue
		}

		ch = append(ch, make(chan struct{}, 10))
		evaluations = append(evaluations, lineEvalPool.Get().(*[68]lineEvaluation))

		go preCompute(evaluations[k-countInf], &Q[k], ch[k-countInf])
	}

	nP = nP - countInf

	var result GT
	result.SetOne()

	j := 0
	for i := len(loopCounter) - 2; i >= 0; i-- {

		result.Square(&result)
		for k := 0; k < nP; k++ {
			<-ch[k]
			lineEval(&result, &evaluations[k][j], &P[k])
		}
		j++

		if loopCounter[i] == 1 {
			for k := 0; k < nP; k++ {
				<-ch[k]
				lineEval(&result, &evaluations[k][j], &P[k])
			}
			j++
		}
	}

	result.Conjugate(&result)

	// release objects into the pool
	for i := 0; i < len(evaluations); i++ {
		lineEvalPool.Put(evaluations[i])
	}

	return result, nil
}

func lineEval(z *GT, l *lineEvaluation, P *G1Affine) *GT {

	l.r2.MulByElement(&l.r2, &P.Y)
	l.r1.MulByElement(&l.r1, &P.X)

	z.MulBy014(&l.r0, &l.r1, &l.r2)
	return z
}

// precomputes the line evaluations used during the Miller loop.
func preCompute(evaluations *[68]lineEvaluation, Q *G2Affine, ch chan struct{}) {

	var Qproj g2Proj
	Qproj.FromAffine(Q)

	j := 0

	for i := len(loopCounter) - 2; i >= 0; i-- {

		Qproj.DoubleStep(&evaluations[j])
		ch <- struct{}{}

		if loopCounter[i] != 0 {
			j++
			Qproj.AddMixedStep(&evaluations[j], Q)
			ch <- struct{}{}
		}
		j++
	}
	close(ch)
}

// DoubleStep doubles a point in Homogenous projective coordinates, and evaluates the line in Miller loop
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g2Proj) DoubleStep(evaluations *lineEvaluation) {

	// get some Element from our pool
	var t0, t1, A, B, C, D, E, EE, F, G, H, I, J, K fptower.E2
	t0.Mul(&p.x, &p.y)
	A.MulByElement(&t0, &twoInv)
	B.Square(&p.y)
	C.Square(&p.z)
	D.Double(&C).
		Add(&D, &C)
	E.Mul(&D, &bTwistCurveCoeff)
	F.Double(&E).
		Add(&F, &E)
	G.Add(&B, &F)
	G.MulByElement(&G, &twoInv)
	H.Add(&p.y, &p.z).
		Square(&H)
	t1.Add(&B, &C)
	H.Sub(&H, &t1)
	I.Sub(&E, &B)
	J.Square(&p.x)
	EE.Square(&E)
	K.Double(&EE).
		Add(&K, &EE)

	// X, Y, Z
	p.x.Sub(&B, &F).
		Mul(&p.x, &A)
	p.y.Square(&G).
		Sub(&p.y, &K)
	p.z.Mul(&B, &H)

	// Line evaluation
	evaluations.r0.Set(&I)
	evaluations.r1.Double(&J).
		Add(&evaluations.r1, &J)
	evaluations.r2.Neg(&H)
}

// AddMixedStep point addition in Mixed Homogenous projective and Affine coordinates
// https://eprint.iacr.org/2013/722.pdf (Section 4.3)
func (p *g2Proj) AddMixedStep(evaluations *lineEvaluation, a *G2Affine) {

	// get some Element from our pool
	var Y2Z1, X2Z1, O, L, C, D, E, F, G, H, t0, t1, t2, J fptower.E2
	Y2Z1.Mul(&a.Y, &p.z)
	O.Sub(&p.y, &Y2Z1)
	X2Z1.Mul(&a.X, &p.z)
	L.Sub(&p.x, &X2Z1)
	C.Square(&O)
	D.Square(&L)
	E.Mul(&L, &D)
	F.Mul(&p.z, &C)
	G.Mul(&p.x, &D)
	t0.Double(&G)
	H.Add(&E, &F).
		Sub(&H, &t0)
	t1.Mul(&p.y, &E)

	// X, Y, Z
	p.x.Mul(&L, &H)
	p.y.Sub(&G, &H).
		Mul(&p.y, &O).
		Sub(&p.y, &t1)
	p.z.Mul(&E, &p.z)

	t2.Mul(&L, &a.Y)
	J.Mul(&a.X, &O).
		Sub(&J, &t2)

	// Line evaluation
	evaluations.r0.Set(&J)
	evaluations.r1.Neg(&O)
	evaluations.r2.Set(&L)
}
