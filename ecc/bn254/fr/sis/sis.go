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
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"math/big"

	"github.com/bits-and-blooms/bitset"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"github.com/consensys/gnark-crypto/internal/parallel"
	"golang.org/x/crypto/blake2b"
)

var (
	ErrNotAPowerOfTwo = errors.New("d must be a power of 2")
)

// Ring-SIS instance
type RSis struct {

	// buffer storing the data to hash
	buffer bytes.Buffer

	// Vectors in ℤ_{p}/Xⁿ+1
	// A[i] is the i-th polynomial.
	// AFftBitreversed the evaluation form of the polynomials in A on the coset √(g) * <g>
	A                    [][]fr.Element
	AfftCosetBitreversed [][]fr.Element

	// LogTwoBound (Inifinty norm) of the vector to hash. It means that each component in m
	// is < 2^B, where m is the vector to hash (the hash being A*m).
	// cd https://hackmd.io/7OODKWQZRRW9RxM5BaXtIw , B >= 3.
	LogTwoBound int

	// maximal number of bytes to sum
	NbBytesToSum int

	// domain for the polynomial multiplication
	Domain *fft.Domain

	// d, the degree of X^{d}+1
	Degree int

	// allocate memory once per instance (used in Sum())
	bufM, bufRes fr.Vector
	bufMValues   *bitset.BitSet
}

// NewRSis creates an instance of RSis.
// seed: seed for the randomness for generating A.
// logTwoDegree: if d := logTwoDegree, the ring will be ℤ_{p}[X]/Xᵈ-1, where X^{2ᵈ} is the 2ᵈ⁺¹-th cyclotomic polynomial
// b: the bound of the vector to hash (using the infinity norm).
// keySize: number of polynomials in A.
func NewRSis(seed int64, logTwoDegree, logTwoBound, keySize int) (hash.Hash, error) {

	var res RSis

	// domains (shift is √{gen} )
	var shift fr.Element
	shift.SetString("19103219067921713944291392827692070036145651957329286315305642004821462161904") // -> 2²⁸-th root of unity of bn254
	e := int64(1 << (28 - (logTwoDegree + 1)))
	shift.Exp(shift, big.NewInt(e))
	res.Domain = fft.NewDomain(uint64(1<<logTwoDegree), shift)

	// bound
	res.LogTwoBound = logTwoBound

	// filling A
	degree := 1 << logTwoDegree
	res.A = make([][]fr.Element, keySize)
	res.AfftCosetBitreversed = make([][]fr.Element, keySize)

	parallel.Execute(keySize, func(start, end int) {
		var buf bytes.Buffer
		for i := start; i < end; i++ {
			res.A[i] = make([]fr.Element, degree)
			res.AfftCosetBitreversed[i] = make([]fr.Element, degree)
			for j := 0; j < degree; j++ {
				res.A[i][j] = genRandom(seed, int64(i), int64(j), &buf)
				res.AfftCosetBitreversed[i][j] = res.A[i][j]
			}
		}
	})

	// filling AfftCosetBitreversed
	for i := 0; i < keySize; i++ {
		res.Domain.FFT(res.AfftCosetBitreversed[i], fft.DIF, fft.WithCoset())
	}

	// computing the maximal size in bytes of a vector to hash
	res.NbBytesToSum = res.LogTwoBound * degree * len(res.A) / 8

	// degree
	res.Degree = degree

	res.bufM = make(fr.Vector, degree*len(res.A))
	res.bufRes = make(fr.Vector, res.Degree)
	res.bufMValues = bitset.New(uint(len(res.A)))

	return &res, nil
}

// Construct a hasher generator. It takes as input the same parameters
// as `NewRingSIS` and outputs a function which returns fresh hasher
// everytime it is called
func NewRingSISMaker(seed int64, logTwoDegree, logTwoBound, keySize int) (func() hash.Hash, error) {
	// domains (shift is √{gen} )
	var shift fr.Element
	shift.SetString("19103219067921713944291392827692070036145651957329286315305642004821462161904") // -> 2²⁸-th root of unity of bn254
	e := int64(1 << (28 - (logTwoDegree + 1)))
	shift.Exp(shift, big.NewInt(e))
	domain := fft.NewDomain(uint64(1<<logTwoDegree), shift)

	// filling A
	degree := 1 << logTwoDegree
	a := make([][]fr.Element, keySize)
	for i := 0; i < keySize; i++ {
		var buf bytes.Buffer
		a[i] = make([]fr.Element, degree)
		for j := 0; j < degree; j++ {
			a[i][j] = genRandom(seed, int64(i), int64(j), &buf)
		}
	}

	// filling AfftCosetBitreversed
	afftCosetBitreversed := make([][]fr.Element, keySize)
	for i := 0; i < keySize; i++ {
		afftCosetBitreversed[i] = make([]fr.Element, degree)
		for j := 0; j < degree; j++ {
			copy(afftCosetBitreversed[i], a[i])
			domain.FFT(afftCosetBitreversed[i], fft.DIF, fft.WithCoset())
		}
	}

	// computing the maximal size in bytes of a vector to hash
	nbBytesToSum := logTwoBound * degree * len(a) / 8

	return func() hash.Hash {
		return &RSis{
			A:                    a,
			AfftCosetBitreversed: afftCosetBitreversed,
			LogTwoBound:          logTwoBound,
			Degree:               degree,
			Domain:               domain,
			NbBytesToSum:         nbBytesToSum,
			bufM:                 make(fr.Vector, degree*len(a)),
			bufRes:               make(fr.Vector, degree),
			bufMValues:           bitset.New(uint(len(a))),
		}
	}, nil

}

func (r *RSis) Write(p []byte) (n int, err error) {
	r.buffer.Write(p)
	return len(p), nil
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
// b is interpreted as a sequence of coefficients of size r.Bound bits long.
// Each coefficient is interpreted in big endian.
// Ex: b = [0xa4, ...] and r.Bound = 4, means that b is decomposed as [10, 4, ...]
// The function returns the hash of the polynomial as a a sequence []fr.Elements, interpreted as []bytes,
// corresponding to sum_i A[i]*m Mod X^{d}+1
func (r *RSis) Sum(b []byte) []byte {
	bufBytes := r.buffer.Bytes()
	if len(bufBytes) > r.NbBytesToSum {
		panic("buffer too large")
	}
	if r.LogTwoBound > 64 {
		panic("r.LogTwoBound too large")
	}

	// clear the buffer of the instance.
	defer func() {
		r.bufMValues.ClearAll()
		for i := 0; i < len(r.bufM); i++ {
			r.bufM[i].SetZero()
		}
		for i := 0; i < len(r.bufRes); i++ {
			r.bufRes[i].SetZero()
		}
	}()

	// bitwise decomposition of the buffer, in order to build m (the vector to hash)
	// as a list of polynomials, whose coefficients are less than r.B bits long.
	nbBitsWritten := len(bufBytes) * 8
	bitAt := func(i int) uint8 {
		k := i / 8
		if k >= len(bufBytes) {
			return 0
		}
		b := bufBytes[k]
		j := i % 8
		return b >> (7 - j) & 1
	}

	// now we can construct m. The input to hash consists of the polynomials
	// m[k*r.Degree:(k+1)*r.Degree]
	m := r.bufM

	// mark blocks m[i*r.Degree : (i+1)*r.Degree] != [0...0]
	mValues := r.bufMValues

	// we process the input buffer by blocks of r.LogTwoBound bits
	// each of these block (<< 64bits) are interpreted as a coefficient
	mPos := 0
	for i := 0; i < nbBitsWritten; mPos++ {
		for j := 0; j < r.LogTwoBound; j++ {
			// r.LogTwoBound < 64; we just use the first word of our element here,
			// and set the bits from LSB to MSB.
			m[mPos][0] |= uint64(bitAt(i) << j)
			i++
		}
		if m[mPos][0] == 0 {
			continue
		}
		mValues.Set(uint(mPos / r.Degree))
	}

	// we can hash now.
	res := r.bufRes

	// method 1: fft
	for i := 0; i < len(r.AfftCosetBitreversed); i++ {
		if !mValues.Test(uint(i)) {
			// means m[i*r.Degree : (i+1)*r.Degree] == [0...0]
			// we can skip this, FFT(0) = 0
			continue
		}
		k := m[i*r.Degree : (i+1)*r.Degree]
		r.Domain.FFT(k, fft.DIF, fft.WithCoset(), fft.WithNbTasks(1))
		mulModAcc(res, r.AfftCosetBitreversed[i], k)
	}
	r.Domain.FFTInverse(res, fft.DIT, fft.WithCoset(), fft.WithNbTasks(1)) // -> reduces mod Xᵈ+1

	// method 2: naive mul THEN naive reduction at the end
	// _res := make([]fr.Element, 2*r.Degree)
	// for i := 0; i < len(r.A); i++ {
	// 	if !mValues.Test(uint(i)) {
	// 		continue
	// 	}
	// 	t := naiveMul(m[i*r.Degree:(i+1)*r.Degree], r.A[i])
	// 	for j := 0; j < 2*r.Degree; j++ {
	// 		_res[j].Add(&t[j], &_res[j])
	// 	}
	// }
	// res = naiveReduction(_res, r.Degree)

	// // method 3: buckets
	// q := make([][]fr.Element, len(r.A))
	// for i := 0; i < len(r.A); i++ { // -> useless conversion, could do it earlier
	// 	q[i] = m[i*r.Degree : (i+1)*r.Degree]
	// }
	// bound := 1 << r.LogTwoBound
	// res = mulModBucketsMethod(r.A, q, bound, r.Degree)

	resBytes, err := res.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return append(b, resBytes[4:]...) // first 4 bytes are uint32(len(res))
}

// Reset resets the Hash to its initial state.
func (r *RSis) Reset() {
	r.buffer.Reset()
}

// Size returns the number of bytes Sum will return.
func (r *RSis) Size() int {

	// The size in bits is the size in bits of a polynomial in A.
	degree := len(r.A[0])
	totalSize := degree * fr.Modulus().BitLen() / 8

	return totalSize
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (r *RSis) BlockSize() int {
	return 0
}

func genRandom(seed, i, j int64, buf *bytes.Buffer) fr.Element {

	buf.Reset()
	buf.WriteString("SIS")
	binary.Write(buf, binary.BigEndian, seed)
	binary.Write(buf, binary.BigEndian, i)
	binary.Write(buf, binary.BigEndian, j)

	digest := blake2b.Sum256(buf.Bytes())

	var res fr.Element
	res.SetBytes(digest[:])

	return res
}
