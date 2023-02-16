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
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

func TestMulMod(t *testing.T) {

	size := 4

	p := make([]fr.Element, size)
	p[0].SetString("2389")
	p[1].SetString("987192")
	p[2].SetString("623")
	p[3].SetString("91")

	q := make([]fr.Element, size)
	q[0].SetString("76755")
	q[1].SetString("232893720")
	q[2].SetString("989273")
	q[3].SetString("675273")

	// creation of the domain
	var shift fr.Element
	shift.SetString("19540430494807482326159819597004422086093766032135589407132600596362845576832")
	domain := fft.NewDomain(uint64(size), shift)

	// mul mod
	domain.FFT(p, fft.DIF, fft.WithCoset())
	domain.FFT(q, fft.DIF, fft.WithCoset())
	r := MulMod(p, q)
	domain.FFTInverse(r, fft.DIT, fft.WithCoset())

	// expected result
	expectedr := make([]fr.Element, 4)
	expectedr[0].SetString("21888242871839275222246405745257275088548364400416034343698204185887558114297")
	expectedr[1].SetString("631644300118")
	expectedr[2].SetString("229913166975959")
	expectedr[3].SetString("1123315390878")

	for i := 0; i < 4; i++ {
		if !expectedr[i].Equal(&r[i]) {
			t.Fatal("product failed")
		}
	}

}

func TestNaiveMul(t *testing.T) {

	size := 64
	d := fft.NewDomain(uint64(2 * size))

	// launch the test 10 times...

	for trials := 0; trials < 10; trials++ {

		// random polynomials
		p := make([]fr.Element, size)
		q := make([]fr.Element, size)
		for i := 0; i < size; i++ {
			p[i].SetRandom()
			q[i].SetRandom()
		}

		// backups to check against fft mul
		_p := make([]fr.Element, 2*size)
		_q := make([]fr.Element, 2*size)
		copy(_p, p)
		copy(_q, q)

		// compute the expected results with fft
		d.FFT(_p, fft.DIF)
		d.FFT(_q, fft.DIF)
		_r := make([]fr.Element, 2*size)
		for i := 0; i < 2*size; i++ {
			_r[i].Mul(&_p[i], &_q[i])
		}
		d.FFTInverse(_r, fft.DIT)

		// compute the result using the naive algo
		r := naiveMul(p, q)

		if !_r[2*size-1].IsZero() {
			t.Fatal("error degree")
		}

		for i := 2*size - 2; i >= 0; i-- {
			if !_r[i].Equal(&r[i]) {
				t.Fatal("error naive mul")
			}
		}

	}

}

func referenceMulMod(d *fft.Domain, p, q []fr.Element) []fr.Element {

	size := len(p)

	_p := make([]fr.Element, size)
	_q := make([]fr.Element, size)
	copy(_p, p)
	copy(_q, q)
	d.FFT(_p, fft.DIF, fft.WithCoset())
	d.FFT(_q, fft.DIF, fft.WithCoset())
	_r := MulMod(_p, _q)
	d.FFTInverse(_r, fft.DIT, fft.WithCoset())

	return _r

}

func TestReduction(t *testing.T) {

	size := 8
	var shift fr.Element
	shift.SetString("14940766826517323942636479241147756311199852622225275649687664389641784935947")
	d := fft.NewDomain(uint64(size), shift)

	// random polynomials
	p := make([]fr.Element, size)
	q := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		p[i].SetRandom()
		q[i].SetRandom()
	}

	// we create the correct result of p*q mod X^d+1, store it in _r
	_r := referenceMulMod(d, p, q)

	// compute the result using naive method
	r := naiveMul(p, q)

	// naive reduction
	rr := naiveReduction(r, size)

	if len(_r) != len(rr) {
		t.Fatal("reduced polynomial size is wrong")
	}

	for i := 0; i < len(_r); i++ {
		if !rr[i].Equal(&_r[i]) {
			t.Fatal("error naive reduction")
		}
	}

}

func TestNaiveMulMod(t *testing.T) {

	size := 8
	var shift fr.Element
	shift.SetString("14940766826517323942636479241147756311199852622225275649687664389641784935947")
	d := fft.NewDomain(uint64(size), shift)

	// random polynomials
	p := make([]fr.Element, size)
	q := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		p[i].SetRandom()
		q[i].SetRandom()
	}

	// expected result
	_r := referenceMulMod(d, p, q)

	// mulMod
	r := NaiveMulMod(p, q)

	// compare...
	if len(r) != len(_r) {
		t.Fatal("lengths are inconsistent")
	}
	for i := 0; i < len(r); i++ {
		if !r[i].Equal(&_r[i]) {
			t.Fatal("error naiveMulMod")
		}
	}

}

func TestSelectIndex(t *testing.T) {

	bound := 8
	nbBitsBound := 3
	sizePolys := 2
	choices := make([]fr.Element, bound)
	for i := 0; i < bound; i++ {
		choices[i].SetUint64(uint64(i))
	}

	// create all possible combinations
	polys := make([][]fr.Element, bound*bound)
	for i := 0; i < bound; i++ {
		for j := 0; j < bound; j++ {
			polys[i*bound+j] = make([]fr.Element, sizePolys)
			polys[i*bound+j][0].SetUint64(uint64(j))
			polys[i*bound+j][1].SetUint64(uint64(i))
		}
	}

	for i := 0; i < 64; i++ {
		a := selectIndex(polys[i], choices, nbBitsBound)
		if a != i {
			t.Fatal("error selection index")
		}
	}

}

func TestFillBucket(t *testing.T) {

	bound := 8
	sizePolys := 2

	// create all possible combinations
	polys := make([][]fr.Element, bound*bound)
	for i := 0; i < bound; i++ {
		for j := 0; j < bound; j++ {
			polys[i*bound+j] = make([]fr.Element, sizePolys)
			polys[i*bound+j][0].SetUint64(uint64(j))
			polys[i*bound+j][1].SetUint64(uint64(i))
		}
	}

	nbBuckets := bound * bound
	buckets := fillBuckets(polys, nbBuckets, bound)

	// some checks
	if len(buckets) != nbBuckets {
		t.Fatal("error size indices")
	}
	for i := 0; i < len(buckets); i++ {
		if len(buckets[i]) != 1 {
			t.Fatal("each bucket should contain only 1 value")
		}
		if buckets[i][0] != i {
			t.Fatal("buckets not filled correctly")
		}
	}

}

func cmp(p, q []fr.Element) bool {
	if len(p) != len(q) {
		return false
	}
	t := true
	for i := 0; i < len(p); i++ {
		t = t && p[i].Equal(&q[i])
	}
	return t
}

func TestAccumulateSums(t *testing.T) {

	bound := 8
	degree := 2

	// create all possible combinations
	polys := make([][]fr.Element, bound*bound)
	for i := 0; i < bound; i++ {
		for j := 0; j < bound; j++ {
			polys[i*bound+j] = make([]fr.Element, degree)
			polys[i*bound+j][0].SetUint64(uint64(j))
			polys[i*bound+j][1].SetUint64(uint64(i))
		}
	}

	nbBuckets := bound * bound
	buckets := fillBuckets(polys, nbBuckets, bound)

	foldedBuckets := accumulateSums(polys, buckets, degree)

	// some checks
	if len(foldedBuckets) != nbBuckets {
		t.Fatal("error size folded buckets")
	}
	for i := 0; i < nbBuckets; i++ {
		if cmp(foldedBuckets[i], polys[i]) == false {
			t.Fatal("folded buckets is not computed correctly")
		}
	}
}

func TestBuildPoly(t *testing.T) {

	bound := 8
	sizePolys := 2

	// create all possible combinations
	polys := make([][]fr.Element, bound*bound)
	for i := 0; i < bound; i++ {
		for j := 0; j < bound; j++ {
			polys[i*bound+j] = make([]fr.Element, sizePolys)
			polys[i*bound+j][0].SetUint64(uint64(j))
			polys[i*bound+j][1].SetUint64(uint64(i))
		}
	}

	degree := 2
	for i := 0; i < bound*bound; i++ {
		b := buildPoly(degree, bound, i)
		if !cmp(b, polys[i]) {
			t.Fatal("error ")
		}
	}

}

func TestMulBucketMethod(t *testing.T) {

	// create the polynomials
	bound := 8
	degree := 2

	// create all possible combinations
	p := make([][]fr.Element, bound*bound)
	q := make([][]fr.Element, bound*bound)
	for i := 0; i < bound*bound; i++ {
		p[i] = buildPoly(degree, bound, i)
		q[i] = buildPoly(degree, bound, i)
	}

	// compute \sum_i p[i]q[i] using the bucket method
	r := mulModBucketsMethod(p, q, bound, degree)

	// expected result (computed using Sage)
	expectedRes := make([]fr.Element, 2)
	expectedRes[0].SetString("0")
	expectedRes[1].SetString("1568")

	// compare the results
	if !cmp(r, expectedRes) {
		t.Fatal("mul bucket failed")
	}
}

func BenchmarkMulMod(b *testing.B) {

	size := 8

	// random polynomials
	p := make([]fr.Element, size)
	q := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		p[i].SetRandom()
		q[i].SetRandom()
	}

	b.Run("naive mul + naive reduction", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r := naiveMul(p, q)
			naiveReduction(r, size)
		}
	})

	b.Run("naive mulMod", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			NaiveMulMod(p, q)
		}
	})

}
