// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash_test

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark-crypto/internal/hashtesting"
)

// stubCompressor is a compression function over blocks of any length.
type stubCompressor struct{}

// CompressField sets dst[i] = 3·left[i] + right[i].
func (stubCompressor) CompressField(dst, left, right []koalabear.Element) {
	for i := range dst {
		var t koalabear.Element
		t.Double(&left[i]).Add(&t, &left[i]).Add(&t, &right[i])
		dst[i] = t
	}
}

// byteStub is stubCompressor on the marshalled elements of blocks of blockLen elements.
type byteStub struct{ blockLen int }

func (s byteStub) BlockSize() int { return s.blockLen * koalabear.Bytes }

func (s byteStub) Compress(left, right []byte) ([]byte, error) {
	x := [2][]koalabear.Element{make([]koalabear.Element, s.blockLen), make([]koalabear.Element, s.blockLen)}
	for i, in := range [][]byte{left, right} {
		for j := range x[i] {
			if err := x[i][j].SetBytesCanonical(in[j*koalabear.Bytes : (j+1)*koalabear.Bytes]); err != nil {
				return nil, err
			}
		}
	}
	dst := make([]koalabear.Element, s.blockLen)
	stubCompressor{}.CompressField(dst, x[0], x[1])
	return hashtesting.Marshal[koalabear.Element, *koalabear.Element](dst), nil
}

// TestFieldMerkleDamgardHasher checks the hasher at several block lengths, see
// [hashtesting.CheckMerkleDamgard].
func TestFieldMerkleDamgardHasher(t *testing.T) {
	randomElements := func(n int) []koalabear.Element {
		e := make([]koalabear.Element, n)
		koalabear.Vector(e).MustSetRandom()
		return e
	}
	for _, blockLen := range []int{1, 3, 4, 8} {
		t.Run(fmt.Sprint(blockLen), func(t *testing.T) {
			newHasher := func() hashtesting.Hasher[koalabear.Element, []koalabear.Element] {
				return hash.NewFieldMerkleDamgardHasher[koalabear.Element, *koalabear.Element](stubCompressor{}, make([]koalabear.Element, blockLen))
			}
			newRef := func() hash.StateStorer {
				return hash.NewMerkleDamgardHasher(byteStub{blockLen}, make([]byte, blockLen*koalabear.Bytes))
			}
			hashtesting.CheckMerkleDamgard(t, newHasher, newRef, hashtesting.Marshal[koalabear.Element, *koalabear.Element], blockLen, randomElements)
		})
	}
}
