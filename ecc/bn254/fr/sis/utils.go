// Copyright 2020 ConsenSys Software Inc.
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

package sis

import (
	"math/bits"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// MulMod computes p * q in ℤ_{p}[X]/Xᵈ+1.
// Is assumed that pLagrangeShifted and qLagrangeShifted are of the corret sizes
// and that they are in evaluation form on √(g) * <g>
// The result is not FFTinversed. The fft inverse is done once every
// multiplications are done.
func MulMod(pLagrangeCosetBitReversed, qLagrangeCosetBitReversed []fr.Element) []fr.Element {

	res := make([]fr.Element, len(pLagrangeCosetBitReversed))
	for i := 0; i < len(pLagrangeCosetBitReversed); i++ {
		res[i].Mul(&pLagrangeCosetBitReversed[i], &qLagrangeCosetBitReversed[i])
	}

	// NOT fft inv for now, wait until every part of the keys have been multiplied
	// r.Domain.FFTInverse(res, fft.DIT, true)

	return res

}

// naiveMul computes the naive multiplication between polynomials
// of the same size. The polynomials (inputs and outputs) are in
// canonical form.
// /!\ The size check is not done in the function /!\
func naiveMul(p, q []fr.Element) []fr.Element {

	n := len(p)
	res := make([]fr.Element, 2*n)

	// q[0]*p + q[1]*p + .. + q[n-1]*p
	var tmp fr.Element
	for i := 0; i < len(q); i++ {
		for j := 0; j < len(p); j++ {
			tmp.Mul(&q[i], &p[j])
			res[i+j].Add(&res[i+j], &tmp)
		}
	}
	return res
}

// naiveReduction computes p Mod X^d+1.
// It assumes that len(p)>=d.
// /!\ The size check is not done in the function /!\
func naiveReduction(p []fr.Element, d int) []fr.Element {
	n := len(p)
	res := make([]fr.Element, n)
	copy(res, p)
	for n > d {
		res[n-d-1].Sub(&res[n-d-1], &res[n-1])
		n--
	}
	return res[:d]
}

// naiveMulMod computes a*b mod X^d+1, where d=len(a).
// It is supposed that len(a) = len(b).
// /!\ The size check is not done in the function /!\
func NaiveMulMod(p, q []fr.Element) []fr.Element {

	d := len(p)
	res := make([]fr.Element, d)

	var tmp fr.Element
	for i := 0; i < d; i++ {
		for j := 0; j < d-i; j++ {
			tmp.Mul(&p[j], &q[i])
			res[i+j].Add(&tmp, &res[i+j])
		}
		for j := d - i; j < d; j++ {
			tmp.Mul(&p[j], &q[i])
			res[j-d+i].Sub(&res[j-d+i], &tmp)
		}
	}

	return res
}

// naiveMulMod2 naiveMulMod with hardcoded degree = 2
func naiveMulMod2(p, q []fr.Element) [2]fr.Element {

	var res [2]fr.Element

	// (p0+p1*X)*(q0+q1*X) Mod X^2+1 = p0q0-p1q1+(p0q1+p1q0)*X
	// We do that in 3 muls instead of 4:
	// a = p0q0
	// b = p1q1
	// c = (p0+p1)*(q0+q1)
	// r = a - b + (c-a-b)*X
	var a, b, c fr.Element
	a.Mul(&p[0], &q[0])
	b.Mul(&p[1], &q[1])
	res[0].Sub(&a, &b)
	c.Add(&p[0], &p[1])
	res[1].Add(&q[0], &q[1]).
		Mul(&res[1], &c).
		Sub(&res[1], &a).
		Sub(&res[1], &b)

	// d := len(p)
	// var tmp fr.Element
	// for i := 0; i < d; i++ {
	// 	for j := 0; j < d-i; j++ {
	// 		tmp.Mul(&p[j], &q[i])
	// 		res[i+j].Add(&tmp, &res[i+j])
	// 	}
	// 	for j := d - i; j < d; j++ {
	// 		tmp.Mul(&p[j], &q[i])
	// 		res[j-d+i].Sub(&res[j-d+i], &tmp)
	// 	}
	// }

	return res
}

// write pols[i] = anX^n + .. + a0, then return the number p_i = a0||..||an
// choices are the possible values of the a_i.
// nbBitsBound is the number of bits of the bound.
func selectIndex(p []fr.Element, choices []fr.Element, nbBitsBound int) int {
	r := 0
	for i := len(p) - 1; i >= 0; i-- {
		for j := 1; j < len(choices); j++ { // don't count 0
			if p[i].Equal(&choices[j]) {
				r += (j << (nbBitsBound * i))
				break
			}
		}
	}
	return r
}

// write pols[i] = anX^n + .. + a0, then write the number p_i = a0||..||an
// and set bucket[p_i] = i.
// The a_i are small (3 bits max) and bounded by bound, and n is small.
// nbBuckets number of possible buckets (it's bound**degree where degree is the
// / degree of the polynomials at play)
// bound is a power of 2.
func fillBuckets(pols [][]fr.Element, nbBuckets int, bound int) [][]int {

	res := make([][]int, nbBuckets)
	for i := 0; i < nbBuckets; i++ {
		res[i] = make([]int, 0, nbBuckets)
	}

	choices := make([]fr.Element, nbBuckets)
	for i := 0; i < nbBuckets; i++ {
		choices[i].SetUint64(uint64(i))
	}
	nbBitsBound := bits.TrailingZeros(uint(bound))
	nbPols := len(pols)
	for i := 0; i < nbPols; i++ {
		ind := selectIndex(pols[i], choices, nbBitsBound)
		res[ind] = append(res[ind], i)
	}
	// parallel.Execute(nbPols, func(start, end int) {
	// 	for i := start; i < end; i++ {
	// 		ind := selectIndex(pols[i], choices, nbBitsBound)
	// 		res[ind] = append(res[ind], i)
	// 	}
	// })

	return res
}

// addPolys p <- p + q
// p and q are assumed to be of the same size.
func addPolys(p, q []fr.Element) {
	for i := 0; i < len(p); i++ {
		p[i].Add(&p[i], &q[i])
	}
}

// accumulateSums returns a list l of len(bucketsList) polynomials,
// such that l[i] = \sum_{k\in\bucketsList[i]}polys[k]
// degree is the degree of the polynomials
func accumulateSums(polys [][]fr.Element, bucketsList [][]int, degree int) [][]fr.Element {

	nbBuckets := len(bucketsList)

	res := make([][]fr.Element, nbBuckets)
	for i := 0; i < nbBuckets; i++ {
		res[i] = make([]fr.Element, degree)
	}

	// to parallelise
	// parallel.Execute(nbBuckets, func(start, end int) {
	// 	for i := start; i < end; i++ {
	// 		for j := 0; j < len(bucketsList[i]); j++ {
	// 			addPolys(res[i], polys[bucketsList[i][j]])
	// 		}
	// 	}
	// })
	for i := 0; i < nbBuckets; i++ {
		for j := 0; j < len(bucketsList[i]); j++ {
			addPolys(res[i], polys[bucketsList[i][j]])
		}
	}

	return res
}

// buildPoly retuns \sum_ai X^i such that a_i = i-th digit (in base bound)
// of id.
// bound is supposed to be a power of two.
// degree is the degree of the expected polynomial.
func buildPoly(degree, bound, id int) []fr.Element {
	res := make([]fr.Element, degree)
	mask := bound - 1
	b := bits.TrailingZeros(uint(bound))
	for i := 0; i < degree; i++ {
		res[i].SetUint64(uint64(id & mask))
		id = id >> b
	}
	return res
}

// mulModBucketsMethod computes \sum_i p[i]*q[i] Mod X^degree+1.
// It is assumed that len(p)=len(q) and that the polynomials are
// of the same size.
// q is the list of polynomials of small norms, < bound.
// /!\ the checks on the sizes are not done in the function /!\
func mulModBucketsMethod(p, q [][]fr.Element, bound, degree int) []fr.Element {

	// 1 - fill the buckets
	nbBuckets := 1
	for i := 0; i < degree; i++ {
		nbBuckets *= bound
	}
	buckets := fillBuckets(q, nbBuckets, bound)

	// 2 - accumulate the sums
	foldedBuckets := accumulateSums(p, buckets, degree)

	// 3 - compute the multiplication on each bucket
	naiveMulPerBuckets := make([][]fr.Element, nbBuckets)
	for i := 0; i < nbBuckets; i++ {
		b := buildPoly(degree, bound, i)
		naiveMulPerBuckets[i] = NaiveMulMod(foldedBuckets[i], b)
	}

	// 4 - sum the results
	for i := 1; i < nbBuckets; i++ {
		for j := 0; j < degree; j++ {
			naiveMulPerBuckets[0][j].Add(&naiveMulPerBuckets[0][j], &naiveMulPerBuckets[i][j])
		}
	}

	// return res
	return naiveMulPerBuckets[0]

}
