// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fft

import (
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"math/bits"
	"runtime"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bw6-633/fr"

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

// GeneratorFullMultiplicativeGroup returns a generator of 𝔽ᵣˣ
func GeneratorFullMultiplicativeGroup() fr.Element {
	var res fr.Element
	res.SetString("13")
	return res
}

// NewDomain returns a subgroup with a power of 2 cardinality
// cardinality >= m
// shift: when specified, it's the element by which the set of root of unity is shifted.
func NewDomain(m uint64, opts ...DomainOption) *Domain {
	opt := domainOptions(opts...)
	domain := &Domain{}
	x := ecc.NextPowerOfTwo(m)
	domain.Cardinality = uint64(x)
	domain.FrMultiplicativeGen = GeneratorFullMultiplicativeGroup()

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

// Twiddles returns the twiddles factor for the FFT using Generator for each stage of the recursive FFT
// or an error if the domain was created with the WithoutPrecompute option
func (d *Domain) Twiddles() ([][]fr.Element, error) {
	if d.twiddles == nil {
		return nil, errors.New("twiddles not precomputed")
	}
	return d.twiddles, nil
}

// TwiddlesInv returns the twiddles factor for the FFT using GeneratorInv for each stage of the recursive FFT
// or an error if the domain was created with the WithoutPrecompute option
func (d *Domain) TwiddlesInv() ([][]fr.Element, error) {
	if d.twiddlesInv == nil {
		return nil, errors.New("twiddles not precomputed")
	}
	return d.twiddlesInv, nil
}

// CosetTable returns the cosetTable u*<1,g,..,g^(n-1)>
// or an error if the domain was created with the WithoutPrecompute option
func (d *Domain) CosetTable() ([]fr.Element, error) {
	if d.cosetTable == nil {
		return nil, errors.New("cosetTable not precomputed")
	}
	return d.cosetTable, nil
}

// CosetTableInv returns the cosetTableInv u*<1,g,..,g^(n-1)>
// or an error if the domain was created with the WithoutPrecompute option
func (d *Domain) CosetTableInv() ([]fr.Element, error) {
	if d.cosetTableInv == nil {
		return nil, errors.New("cosetTableInv not precomputed")
	}
	return d.cosetTableInv, nil
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
		BuildExpTable(sqrt, t)
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
	if nbStages == 0 {
		return
	}
	if len(t) != int(nbStages) {
		panic("invalid twiddle table")
	}
	// we just compute the first stage
	t[0] = make([]fr.Element, 1+(1<<(nbStages-1)))
	BuildExpTable(omega, t[0])

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

// BuildExpTable precomputes the first n powers of w in parallel
// table[0] = w^0
// table[1] = w^1
// ...
func BuildExpTable(w fr.Element, table []fr.Element) {
	table[0].SetOne()
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
	// note to stay retro compatible with previous version using ecc/encoder, we encode as:
	// d.Cardinality, &d.CardinalityInv, &d.Generator, &d.GeneratorInv, &d.FrMultiplicativeGen, &d.FrMultiplicativeGenInv, &d.withPrecompute

	var written int64
	var err error

	err = binary.Write(w, binary.BigEndian, d.Cardinality)
	if err != nil {
		return written, err
	}
	written += 8

	toEncode := []*fr.Element{&d.CardinalityInv, &d.Generator, &d.GeneratorInv, &d.FrMultiplicativeGen, &d.FrMultiplicativeGenInv}
	for _, v := range toEncode {
		buf := v.Bytes()
		_, err = w.Write(buf[:])
		if err != nil {
			return written, err
		}
		written += fr.Bytes
	}

	err = binary.Write(w, binary.BigEndian, d.withPrecompute)
	if err != nil {
		return written, err
	}
	written += 1

	return written, nil
}

// ReadFrom attempts to decode a domain from Reader
func (d *Domain) ReadFrom(r io.Reader) (int64, error) {

	var read int64
	var err error

	err = binary.Read(r, binary.BigEndian, &d.Cardinality)
	if err != nil {
		return read, err
	}
	read += 8

	toDecode := []*fr.Element{&d.CardinalityInv, &d.Generator, &d.GeneratorInv, &d.FrMultiplicativeGen, &d.FrMultiplicativeGenInv}

	for _, v := range toDecode {
		var buf [fr.Bytes]byte
		_, err = r.Read(buf[:])
		if err != nil {
			return read, err
		}
		read += fr.Bytes
		*v, err = fr.BigEndian.Element(&buf)
		if err != nil {
			return read, err
		}
	}

	err = binary.Read(r, binary.BigEndian, &d.withPrecompute)
	if err != nil {
		return read, err
	}
	read += 1

	if d.withPrecompute {
		d.preComputeTwiddles()
	}

	return read, nil
}
