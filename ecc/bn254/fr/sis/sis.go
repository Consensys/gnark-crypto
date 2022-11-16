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

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
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
}

func genRandom(seed, i, j int64) fr.Element {

	var buf bytes.Buffer
	buf.WriteString("SIS")
	binary.Write(&buf, binary.BigEndian, seed)
	binary.Write(&buf, binary.BigEndian, i)
	binary.Write(&buf, binary.BigEndian, j)

	slice := buf.Bytes()
	digest := blake2b.Sum256(slice)

	var res fr.Element
	res.SetBytes(digest[:])

	return res
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
	for i := 0; i < keySize; i++ {
		res.A[i] = make([]fr.Element, degree)
		for j := 0; j < degree; j++ {
			res.A[i][j] = genRandom(seed, int64(i), int64(j))
		}
	}

	// filling AfftCosetBitreversed
	res.AfftCosetBitreversed = make([][]fr.Element, keySize)
	for i := 0; i < keySize; i++ {
		res.AfftCosetBitreversed[i] = make([]fr.Element, degree)
		for j := 0; j < degree; j++ {
			copy(res.AfftCosetBitreversed[i], res.A[i])
			res.Domain.FFT(res.AfftCosetBitreversed[i], fft.DIF, true)
		}
	}

	// computing the maximal size in bytes of a vector to hash
	res.NbBytesToSum = res.LogTwoBound * degree * len(res.A) / 8

	// degree
	res.Degree = degree

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
		a[i] = make([]fr.Element, degree)
		for j := 0; j < degree; j++ {
			a[i][j] = genRandom(seed, int64(i), int64(j))
		}
	}

	// filling AfftCosetBitreversed
	afftCosetBitreversed := make([][]fr.Element, keySize)
	for i := 0; i < keySize; i++ {
		afftCosetBitreversed[i] = make([]fr.Element, degree)
		for j := 0; j < degree; j++ {
			copy(afftCosetBitreversed[i], a[i])
			domain.FFT(afftCosetBitreversed[i], fft.DIF, true)
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
// The function returns the hash of the polynomial as a a sequence []fr.Elements, interepreted as []bytes,
// corresponding to sum_i A[i]*m Mod X^{d}+1
func (r *RSis) Sum(b []byte) []byte {

	// if NbBytesToSums is not reached, the buffer is padded with zeroes.
	sizeBuffer := r.buffer.Len()
	if sizeBuffer < r.NbBytesToSum {
		toPadd := make([]byte, r.NbBytesToSum-sizeBuffer)
		_, err := r.buffer.Write(toPadd)
		if err != nil {
			panic(err)
		}
	}

	// bitwise decomposition of the buffer, in order to build m (the vector to hash)
	// as a list of polynomials, whose coefficients are less than r.B bits long.
	totalNbBits := r.NbBytesToSum * 8
	mBits := make([]byte, totalNbBits)
	var tmp [1]byte
	for i := 0; i < r.NbBytesToSum; i++ {
		_, err := r.buffer.Read(tmp[:])
		if err != nil {
			panic(err)
		}
		for j := 0; j < 8; j++ {
			mBits[i*8+j] = (tmp[0] >> (7 - j)) & 1
		}
	}

	// now we can construct m. The input to hash consists of the polynomials
	// m[k*r.Degree:(k+1)*r.Degree]
	nbBytesPerCoefficients := (r.LogTwoBound - (r.LogTwoBound % 8)) / 8
	nbBitsPerCoefficients := r.LogTwoBound
	offset := nbBitsPerCoefficients % 8
	sizeM := r.Degree * len(r.A)
	buf := make([]byte, nbBytesPerCoefficients+1)
	m := make([]fr.Element, sizeM)
	for i := 0; i < totalNbBits/nbBitsPerCoefficients; i++ {
		for j := 0; j < offset; j++ {
			buf[0] += (mBits[i*nbBitsPerCoefficients+j]) << (offset - 1 - j)
		}
		for j := 0; j < nbBytesPerCoefficients; j++ {
			for k := 0; k < 8; k++ {
				buf[j+1] += (mBits[i*nbBitsPerCoefficients+offset+8*j+k]) << (7 - k)
			}
		}
		m[i].SetBytes(buf)
		for j := 0; j < nbBytesPerCoefficients+1; j++ {
			buf[j] = 0
		}
	}

	// we can hash now.
	res := make([]fr.Element, r.Degree)

	// method 1: fft
	if r.Degree > 3 { // we keep this track to have a complete code...
		for i := 0; i < len(r.AfftCosetBitreversed); i++ {
			r.Domain.FFT(m[i*r.Degree:(i+1)*r.Degree], fft.DIF, true)
			t := MulMod(r.AfftCosetBitreversed[i], m[i*r.Degree:(i+1)*r.Degree])
			for j := 0; j < len(res); j++ {
				res[j].Add(&res[j], &t[j])
			}
		}
		r.Domain.FFTInverse(res, fft.DIT, true) // -> automagically reduces mod Xᵈ+1
	} else if r.Degree == 2 { // method 2: naive mulMod+reductions
		// nbCPUs := runtime.NumCPU()
		// _res := make([][2]fr.Element, nbCPUs)
		// chDone := make(chan int, nbCPUs-1)
		// sizePerTask := int(len(r.A) / nbCPUs)
		// for i := 0; i < nbCPUs-1; i++ {
		// 	start := i * sizePerTask
		// 	end := (i + 1) * sizePerTask
		// 	go func(start, end, i int) {
		// 		for j := start; j < end; j++ {
		// 			t := naiveMulMod2(m[j*r.Degree:(j+1)*r.Degree], r.A[j])
		// 			_res[i][0].Add(&t[0], &_res[i][0])
		// 			_res[i][1].Add(&t[1], &_res[i][1])
		// 		}
		// 		chDone <- i
		// 	}(start, end, i)
		// }
		// start := (nbCPUs - 1) * sizePerTask
		// end := len(r.A)
		// for j := start; j < end; j++ {
		// 	t := naiveMulMod2(m[j*r.Degree:(j+1)*r.Degree], r.A[j])
		// 	_res[nbCPUs-1][0].Add(&t[0], &_res[nbCPUs-1][0])
		// 	_res[nbCPUs-1][1].Add(&t[1], &_res[nbCPUs-1][1])
		// }

		// for i := 0; i < nbCPUs-1; i++ {
		// 	done := <-chDone
		// 	res[0].Add(&res[0], &_res[done][0])
		// 	res[1].Add(&res[1], &_res[done][1])
		// }
		// close(chDone)
		// res[0].Add(&res[0], &_res[nbCPUs-1][0])
		// res[1].Add(&res[1], &_res[nbCPUs-1][1])

		for i := 0; i < len(r.A); i++ {
			t := naiveMulMod2(m[i*r.Degree:(i+1)*r.Degree], r.A[i])
			res[0].Add(&t[0], &res[0])
			res[1].Add(&t[1], &res[1])
		}
	} else {
		panic("SIS must be > 1")
	}

	// method 3: naive mul THEN naive reduction at the end
	// _res := make([]fr.Element, 2*r.Degree)
	// for i := 0; i < len(r.A); i++ {
	// 	t := naiveMul(m[i*r.Degree:(i+1)*r.Degree], r.A[i])
	// 	for j := 0; j < 2*r.Degree; j++ {
	// 		_res[j].Add(&t[j], &_res[j])
	// 	}
	// }
	// res = naiveReduction(_res, r.Degree)

	// method 4: buckets
	// q := make([][]fr.Element, len(r.A))
	// for i := 0; i < len(r.A); i++ { // -> useless conversion, could do it earlier
	// 	q[i] = m[i*r.Degree : (i+1)*r.Degree]
	// }
	// bound := 1 << r.LogTwoBound
	// res = mulModBucketsMethod(r.A, q, bound, r.Degree)

	sizeFrElmt := len(res[0].Bytes())
	resBytes := make([]byte, sizeFrElmt*r.Degree)
	for i := 0; i < r.Degree; i++ {
		b := res[i].Bytes()
		copy(resBytes[i*sizeFrElmt:(i+1)*sizeFrElmt], b[:])
	}

	return append(b, resBytes...)
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
