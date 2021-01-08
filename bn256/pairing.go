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

import (
	"errors"
	"sync"

	"github.com/consensys/gurvy/bn256/internal/fptower"
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

	// https://eprint.iacr.org/2008/490.pdf
	var mt [4]GT // mt[i] is m^(t^i)

	// easy part
	mt[0].Set(&result)
	var temp GT
	temp.Conjugate(&mt[0])
	mt[0].Inverse(&mt[0])
	temp.Mul(&temp, &mt[0])
	mt[0].FrobeniusSquare(&temp).
		Mul(&mt[0], &temp)

	// hard part
	mt[1].Expt(&mt[0])
	mt[2].Expt(&mt[1])
	mt[3].Expt(&mt[2])

	var y [7]GT

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
	mt[0].CyclotomicSquare(&y[6])
	mt[0].Mul(&mt[0], &y[4])
	mt[0].Mul(&mt[0], &y[5])
	mt[1].Mul(&y[3], &y[5])
	mt[1].Mul(&mt[1], &mt[0])
	mt[0].Mul(&mt[0], &y[2])
	mt[1].CyclotomicSquare(&mt[1])
	mt[1].Mul(&mt[1], &mt[0])
	mt[1].CyclotomicSquare(&mt[1])
	mt[0].Mul(&mt[1], &y[1])
	mt[1].Mul(&mt[1], &y[0])
	mt[0].CyclotomicSquare(&mt[0])
	result.Mul(&mt[0], &mt[1])

	return result
}

var lineEvalPool = sync.Pool{
	New: func() interface{} {
		return new([88]lineEvaluation)
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
		evaluations = make([]*[88]lineEvaluation, 0, nP)
		Paff        = make([]G1Affine, nP)
	)

	var countInf = 0
	for k := 0; k < nP; k++ {
		if P[k].IsInfinity() || Q[k].IsInfinity() {
			countInf++
			continue
		}

		ch = append(ch, make(chan struct{}, 10))
		evaluations = append(evaluations, lineEvalPool.Get().(*[88]lineEvaluation))

		Paff[k-countInf].Set(&P[k])
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
			lineEval(&result, &evaluations[k][j], &Paff[k])
		}
		j++

		if loopCounter[i] != 0 {
			for k := 0; k < nP; k++ {
				<-ch[k]
				lineEval(&result, &evaluations[k][j], &Paff[k])
			}
			j++
		}
	}

	// cf https://eprint.iacr.org/2010/354.pdf for instance for optimal Ate Pairing
	for k := 0; k < nP; k++ {
		<-ch[k]
		lineEval(&result, &evaluations[k][j], &Paff[k])
	}
	j++
	for k := 0; k < nP; k++ {
		<-ch[k]
		lineEval(&result, &evaluations[k][j], &Paff[k])
	}

	// release objects into the pool
	go func() {
		for i := 0; i < len(evaluations); i++ {
			lineEvalPool.Put(evaluations[i])
		}
	}()

	return result, nil
}

func lineEval(z *GT, l *lineEvaluation, P *G1Affine) *GT {

	l.r0.MulByElement(&l.r0, &P.Y)
	l.r1.MulByElement(&l.r1, &P.X)

	z.MulBy034(&l.r0, &l.r1, &l.r2)

	return z
}

// precomputes the line evaluations used during the Miller loop.
func preCompute(evaluations *[88]lineEvaluation, Q *G2Affine, ch chan struct{}) {

	var Qproj g2Proj
	Qproj.FromAffine(Q)
	var Qneg G2Affine
	Qneg.Neg(Q)

	j := 0

	for i := len(loopCounter) - 2; i >= 0; i-- {

		Qproj.DoubleStep(&evaluations[j])
		ch <- struct{}{}

		if loopCounter[i] == 1 {
			j++
			Qproj.AddMixedStep(&evaluations[j], Q)
			ch <- struct{}{}
		} else if loopCounter[i] == -1 {
			j++
			Qproj.AddMixedStep(&evaluations[j], &Qneg)
			ch <- struct{}{}
		}
		j++
	}

	var Q1, Q2 G2Affine
	//Q1 = Frob(Q)
	Q1.X.Conjugate(&Q.X).MulByNonResidue1Power2(&Q1.X)
	Q1.Y.Conjugate(&Q.Y).MulByNonResidue1Power3(&Q1.Y)

	// Q2 = -Frob2(Q)
	Q2.X.MulByNonResidue2Power2(&Q.X)
	Q2.Y.MulByNonResidue2Power3(&Q.Y).Neg(&Q2.Y)

	Qproj.AddMixedStep(&evaluations[j], &Q1)
	ch <- struct{}{}
	j++
	Qproj.AddMixedStep(&evaluations[j], &Q2)
	ch <- struct{}{}

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
	evaluations.r0.Neg(&H)
	evaluations.r1.Double(&J).
		Add(&evaluations.r1, &J)
	evaluations.r2.Set(&I)
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
	evaluations.r0.Set(&L)
	evaluations.r1.Neg(&O)
	evaluations.r2.Set(&J)
}
