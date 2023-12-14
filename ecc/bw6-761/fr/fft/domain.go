// Copyright 2020 Consensys Software Inc.
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

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fft

import (
	"io"
	"math/big"
	"math/bits"
	"runtime"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"

	curve "github.com/consensys/gnark-crypto/ecc/bw6-761"

	"github.com/consensys/gnark-crypto/ecc"
)

// Domain with a power of 2 cardinality
// compute a field element of order 2x and store it in FinerGenerator
// all other values can be derived from x, GeneratorSqrt
type Domain struct {
	Cardinality            uint64
	CardinalityInv         fr.Element
	Generator              fr.Element
	GeneratorInv           fr.Element
	FrMultiplicativeGen    fr.Element // generator of Fr*
	FrMultiplicativeGenInv fr.Element

	// this is set with the WithoutPrecompute option;
	// if true, the domain does some pre-computation and stores it.
	// if false, the FFT will compute the twiddles on the fly (this is less CPU efficient, but uses less memory)
	withPrecompute bool

	// the following slices are not serialized and are (re)computed through domain.preComputeTwiddles()

	// twiddles factor for the FFT using Generator for each stage of the recursive FFT
	twiddles [][]fr.Element

	// twiddles factor for the FFT using GeneratorInv for each stage of the recursive FFT
	twiddlesInv [][]fr.Element

	// we precompute these mostly to avoid the memory intensive bit reverse permutation in the groth16.Prover

	// cosetTable u*<1,g,..,g^(n-1)>
	cosetTable []fr.Element

	// cosetTable[i][j] = domain.Generator(i-th)SqrtInv ^ j
	cosetTableInv []fr.Element
}

// NewDomain returns a subgroup with a power of 2 cardinality
// cardinality >= m
// shift: when specified, it's the element by which the set of root of unity is shifted.
func NewDomain(m uint64, opts ...DomainOption) *Domain {
	opt := domainOptions(opts...)
	domain := &Domain{}
	x := ecc.NextPowerOfTwo(m)
	domain.Cardinality = uint64(x)

	// generator of the largest 2-adic subgroup

	domain.FrMultiplicativeGen.SetUint64(15)

	if opt.shift != nil {
		domain.FrMultiplicativeGen.Set(opt.shift)
	}
	domain.FrMultiplicativeGenInv.Inverse(&domain.FrMultiplicativeGen)

	var err error
	domain.Generator, err = Generator(m)
	if err != nil {
		panic(err)
	}
	domain.GeneratorInv.Inverse(&domain.Generator)
	domain.CardinalityInv.SetUint64(uint64(x)).Inverse(&domain.CardinalityInv)

	// twiddle factors
	domain.withPrecompute = opt.withPrecompute
	if domain.withPrecompute {
		domain.preComputeTwiddles()
	}

	return domain
}

// Generator returns a generator for Z/2^(log(m))Z
// or an error if m is too big (required root of unity doesn't exist)
func Generator(m uint64) (fr.Element, error) {
	return fr.Generator(m)
}

func (d *Domain) preComputeTwiddles() {

	// nb fft stages
	nbStages := uint64(bits.TrailingZeros64(d.Cardinality))

	d.twiddles = make([][]fr.Element, nbStages)
	d.twiddlesInv = make([][]fr.Element, nbStages)
	d.cosetTable = make([]fr.Element, d.Cardinality)
	d.cosetTableInv = make([]fr.Element, d.Cardinality)

	var wg sync.WaitGroup

	expTable := func(sqrt fr.Element, t []fr.Element) {
		t[0] = fr.One()
		precomputeExpTable(sqrt, t)
		wg.Done()
	}

	wg.Add(4)
	go func() {
		buildTwiddles(d.twiddles, d.Generator, nbStages)
		wg.Done()
	}()
	go func() {
		buildTwiddles(d.twiddlesInv, d.GeneratorInv, nbStages)
		wg.Done()
	}()
	go expTable(d.FrMultiplicativeGen, d.cosetTable)
	go expTable(d.FrMultiplicativeGenInv, d.cosetTableInv)

	wg.Wait()

}

func buildTwiddles(t [][]fr.Element, omega fr.Element, nbStages uint64) {
	if len(t) != int(nbStages) {
		panic("invalid twiddle table")
	}
	// we just compute the first stage
	t[0] = make([]fr.Element, 1+(1<<(nbStages-1)))
	t[0][0] = fr.One()
	precomputeExpTable(omega, t[0])

	// for the next stages, we just iterate on the first stage with larger stride
	for i := uint64(1); i < nbStages; i++ {
		t[i] = make([]fr.Element, 1+(1<<(nbStages-i-1)))
		k := 0
		for j := 0; j < len(t[i]); j++ {
			t[i][j] = t[0][k]
			k += 1 << i
		}
	}

}

func precomputeExpTable(w fr.Element, table []fr.Element) {
	n := len(table)

	// see if it makes sense to parallelize exp tables pre-computation
	interval := 0
	if runtime.NumCPU() >= 4 {
		interval = (n - 1) / (runtime.NumCPU() / 4)
	}

	// this ratio roughly correspond to the number of multiplication one can do in place of a Exp operation
	// TODO @gbotrel revisit this; Exps in this context will be by a "small power of 2" so faster than this ref ratio.
	const ratioExpMul = 6000 / 17

	if interval < ratioExpMul {
		precomputeExpTableChunk(w, 1, table[1:])
		return
	}

	// we parallelize
	var wg sync.WaitGroup
	for i := 1; i < n; i += interval {
		start := i
		end := i + interval
		if end > n {
			end = n
		}
		wg.Add(1)
		go func() {
			precomputeExpTableChunk(w, uint64(start), table[start:end])
			wg.Done()
		}()
	}
	wg.Wait()
}

func precomputeExpTableChunk(w fr.Element, power uint64, table []fr.Element) {

	// this condition ensures that creating a domain of size 1 with cosets don't fail
	if len(table) > 0 {
		table[0].Exp(w, new(big.Int).SetUint64(power))
		for i := 1; i < len(table); i++ {
			table[i].Mul(&table[i-1], &w)
		}
	}
}

// WriteTo writes a binary representation of the domain (without the precomputed twiddle factors)
// to the provided writer
func (d *Domain) WriteTo(w io.Writer) (int64, error) {

	enc := curve.NewEncoder(w)

	toEncode := []interface{}{d.Cardinality, &d.CardinalityInv, &d.Generator, &d.GeneratorInv, &d.FrMultiplicativeGen, &d.FrMultiplicativeGenInv, &d.withPrecompute}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}

// ReadFrom attempts to decode a domain from Reader
func (d *Domain) ReadFrom(r io.Reader) (int64, error) {

	dec := curve.NewDecoder(r)

	toDecode := []interface{}{&d.Cardinality, &d.CardinalityInv, &d.Generator, &d.GeneratorInv, &d.FrMultiplicativeGen, &d.FrMultiplicativeGenInv, &d.withPrecompute}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	if d.withPrecompute {
		d.preComputeTwiddles()
	}

	return dec.BytesRead(), nil
}
