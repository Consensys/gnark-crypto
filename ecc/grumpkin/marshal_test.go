// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package grumpkin

import (
	"bytes"
	crand "crypto/rand"
	"io"
	"math/big"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"

	"github.com/consensys/gnark-crypto/ecc/grumpkin/fp"
	"github.com/consensys/gnark-crypto/ecc/grumpkin/fr"
)

const (
	nbFuzzShort = 10
	nbFuzz      = 100
)

func TestEncoder(t *testing.T) {
	t.Parallel()
	// TODO need proper fuzz testing here

	var inA uint64
	var inB fr.Element
	var inC fp.Element
	var inD G1Affine
	var inE G1Affine
	var inG []G1Affine
	var inI []fp.Element
	var inJ []fr.Element
	var inK fr.Vector
	var inL [][]fr.Element
	var inM [][]uint64
	var inN [][][]fr.Element

	// set values of inputs
	inA = rand.Uint64() //#nosec G404 weak rng is fine here
	inB.MustSetRandom()
	inC.MustSetRandom()
	inD.ScalarMultiplication(&g1GenAff, new(big.Int).SetUint64(rand.Uint64())) //#nosec G404 weak rng is fine here
	// inE --> infinity
	inG = make([]G1Affine, 2)
	inG[1] = inD
	inI = make([]fp.Element, 3)
	inI[2] = inD.X
	inJ = make([]fr.Element, 0)
	inK = make(fr.Vector, 42)
	inK[41].SetUint64(42)
	inL = [][]fr.Element{inJ, inK}
	inM = [][]uint64{{1, 2}, {4}, {}}
	inN = make([][][]fr.Element, 4)
	for i := 0; i < 4; i++ {
		inN[i] = make([][]fr.Element, i+2)
		for j := 0; j < i+2; j++ {
			inN[i][j] = make([]fr.Element, j+3)
			fr.Vector(inN[i][j]).MustSetRandom()
		}
	}

	// encode them, compressed and raw
	var buf, bufRaw bytes.Buffer
	enc := NewEncoder(&buf)
	encRaw := NewEncoder(&bufRaw, RawEncoding())
	toEncode := []interface{}{inA, &inB, &inC, &inD, &inE, inG, inI, inJ, inK, inL, inM, inN}
	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			t.Fatal(err)
		}
		if err := encRaw.Encode(v); err != nil {
			t.Fatal(err)
		}
	}

	testDecode := func(t *testing.T, r io.Reader, n int64) {
		dec := NewDecoder(r)
		var outA uint64
		var outB fr.Element
		var outC fp.Element
		var outD G1Affine
		var outE G1Affine
		outE.X.SetOne()
		outE.Y.SetUint64(42)
		var outG []G1Affine
		var outI []fp.Element
		var outJ []fr.Element
		var outK fr.Vector
		var outL [][]fr.Element
		var outM [][]uint64
		var outN [][][]fr.Element

		toDecode := []interface{}{&outA, &outB, &outC, &outD, &outE, &outG, &outI, &outJ, &outK, &outL, &outM, &outN}
		for _, v := range toDecode {
			if err := dec.Decode(v); err != nil {
				t.Fatal(err)
			}
		}

		// compare values
		if inA != outA {
			t.Fatal("didn't encode/decode uint64 value properly")
		}

		if !inB.Equal(&outB) || !inC.Equal(&outC) {
			t.Fatal("decode(encode(Element) failed")
		}
		if !inD.Equal(&outD) || !inE.Equal(&outE) {
			t.Fatal("decode(encode(G1Affine) failed")
		}
		for i := 0; i < len(inG); i++ {
			if !inG[i].Equal(&outG[i]) {
				t.Fatal("decode(encode(slice(points))) failed")
			}
		}
		if (len(inI) != len(outI)) || (len(inJ) != len(outJ)) {
			t.Fatal("decode(encode(slice(elements))) failed")
		}
		for i := 0; i < len(inI); i++ {
			if !inI[i].Equal(&outI[i]) {
				t.Fatal("decode(encode(slice(elements))) failed")
			}
		}
		if !reflect.DeepEqual(inK, outK) {
			t.Fatal("decode(encode(vector)) failed")
		}
		if !reflect.DeepEqual(inL, outL) {
			t.Fatal("decode(encode(slice²(elements))) failed")
		}
		if !reflect.DeepEqual(inM, outM) {
			t.Fatal("decode(encode(slice²(uint64))) failed")
		}
		if !reflect.DeepEqual(inN, outN) {
			t.Fatal("decode(encode(slice^{3}(uint64))) failed")
		}
		if n != dec.BytesRead() {
			t.Fatal("bytes read don't match bytes written")
		}
	}

	// decode them
	testDecode(t, &buf, enc.BytesWritten())
	testDecode(t, &bufRaw, encRaw.BytesWritten())

}

func TestIsCompressed(t *testing.T) {
	t.Parallel()
	var g1Inf, g1 G1Affine

	g1 = g1GenAff

	{
		b := g1Inf.Bytes()
		if !isCompressed(b[0]) {
			t.Fatal("g1Inf.Bytes() should be compressed")
		}
	}

	{
		b := g1Inf.RawBytes()
		if isCompressed(b[0]) {
			t.Fatal("g1Inf.RawBytes() should be uncompressed")
		}
	}

	{
		b := g1.Bytes()
		if !isCompressed(b[0]) {
			t.Fatal("g1.Bytes() should be compressed")
		}
	}

	{
		b := g1.RawBytes()
		if isCompressed(b[0]) {
			t.Fatal("g1.RawBytes() should be uncompressed")
		}
	}

}

func TestG1AffineSerialization(t *testing.T) {
	t.Parallel()
	// test round trip serialization of infinity
	{
		// compressed
		{
			var p1, p2 G1Affine
			p2.X.MustSetRandom()
			p2.Y.MustSetRandom()
			buf := p1.Bytes()
			n, err := p2.SetBytes(buf[:])
			if err != nil {
				t.Fatal(err)
			}
			if n != SizeOfG1AffineCompressed {
				t.Fatal("invalid number of bytes consumed in buffer")
			}
			if !(p2.X.IsZero() && p2.Y.IsZero()) {
				t.Fatal("deserialization of uncompressed infinity point is not infinity")
			}
		}

		// uncompressed
		{
			var p1, p2 G1Affine
			p2.X.MustSetRandom()
			p2.Y.MustSetRandom()
			buf := p1.RawBytes()
			n, err := p2.SetBytes(buf[:])
			if err != nil {
				t.Fatal(err)
			}
			if n != SizeOfG1AffineUncompressed {
				t.Fatal("invalid number of bytes consumed in buffer")
			}
			if !(p2.X.IsZero() && p2.Y.IsZero()) {
				t.Fatal("deserialization of uncompressed infinity point is not infinity")
			}
		}
	}

	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[G1] Affine SetBytes(RawBytes) should stay the same", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G1Affine
			var ab big.Int
			a.BigInt(&ab)
			start.ScalarMultiplication(&g1GenAff, &ab)

			buf := start.RawBytes()
			n, err := end.SetBytes(buf[:])
			if err != nil {
				return false
			}
			if n != SizeOfG1AffineUncompressed {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		GenFp(),
	))

	properties.Property("[G1] Affine SetBytes(Bytes()) should stay the same", prop.ForAll(
		func(a fp.Element) bool {
			var start, end G1Affine
			var ab big.Int
			a.BigInt(&ab)
			start.ScalarMultiplication(&g1GenAff, &ab)

			buf := start.Bytes()
			n, err := end.SetBytes(buf[:])
			if err != nil {
				return false
			}
			if n != SizeOfG1AffineCompressed {
				return false
			}
			return start.X.Equal(&end.X) && start.Y.Equal(&end.Y)
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// define Gopters generators

// GenFr generates an Fr element
func GenFr() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var elmt fr.Element
		elmt.MustSetRandom()

		return gopter.NewGenResult(elmt, gopter.NoShrinker)
	}
}

// GenFp generates an Fp element
func GenFp() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var elmt fp.Element
		elmt.MustSetRandom()

		return gopter.NewGenResult(elmt, gopter.NoShrinker)
	}
}

// GenBigInt generates a big.Int
func GenBigInt() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var s big.Int
		var b [fp.Bytes]byte
		_, err := crand.Read(b[:])
		if err != nil {
			panic(err)
		}
		s.SetBytes(b[:])
		genResult := gopter.NewGenResult(s, gopter.NoShrinker)
		return genResult
	}
}
