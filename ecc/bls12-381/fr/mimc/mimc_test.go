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

package mimc_test

import (
	"bytes"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr/mimc"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/consensys/gnark-crypto/hash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiMCFiatShamir(t *testing.T) {
	fs := fiatshamir.NewTranscript(mimc.NewMiMC(), "c0")
	zero := make([]byte, mimc.BlockSize)
	err := fs.Bind("c0", zero)
	assert.NoError(t, err)
	_, err = fs.ComputeChallenge("c0")
	assert.NoError(t, err)
}

func TestByteOrder(t *testing.T) {
	assert := require.New(t)

	var buf [fr.Bytes]byte
	// if the 31 first bytes are FF, it's a valid FF in little endian, but not in big endian
	for i := 0; i < fr.Bytes-1; i++ {
		buf[i] = 0xFF
	}
	_, err := fr.BigEndian.Element(&buf)
	assert.Error(err)
	_, err = fr.LittleEndian.Element(&buf)
	assert.NoError(err)

	{
		// hashing buf with big endian should fail
		mimcHash := mimc.NewMiMC(mimc.WithByteOrder(fr.BigEndian))
		_, err := mimcHash.Write(buf[:])
		assert.Error(err)
	}

	{
		// hashing buf with little endian should succeed
		mimcHash := mimc.NewMiMC(mimc.WithByteOrder(fr.LittleEndian))
		_, err := mimcHash.Write(buf[:])
		assert.NoError(err)
	}

	buf = [fr.Bytes]byte{}
	// if the 31 bytes are FF, it's a valid FF in big endian, but not in little endian
	for i := 1; i < fr.Bytes; i++ {
		buf[i] = 0xFF
	}
	_, err = fr.BigEndian.Element(&buf)
	assert.NoError(err)
	_, err = fr.LittleEndian.Element(&buf)
	assert.Error(err)

	{
		// hashing buf with big endian should succeed
		mimcHash := mimc.NewMiMC(mimc.WithByteOrder(fr.BigEndian))
		_, err := mimcHash.Write(buf[:])
		assert.NoError(err)
	}

	{
		// hashing buf with little endian should fail
		mimcHash := mimc.NewMiMC(mimc.WithByteOrder(fr.LittleEndian))
		_, err := mimcHash.Write(buf[:])
		assert.Error(err)
	}
}

func TestSetState(t *testing.T) {
	// we use for hashing and retrieving the state
	h1 := mimc.NewMiMC()
	// only hashing
	h2 := mimc.NewMiMC()
	// we use for restoring from state
	h3 := mimc.NewMiMC()

	randInputs := make([]fr.Element, 10)
	for i := range randInputs {
		randInputs[i].SetRandom()
	}

	storedStates := make([][]byte, len(randInputs))

	hs, ok := h1.(hash.StateStorer)
	if !ok {
		t.Fatal("hash function does not implement stateStorer interface")
	}

	for i := range randInputs {
		storedStates[i] = hs.State()

		h1.Write(randInputs[i].Marshal())
		h2.Write(randInputs[i].Marshal())
	}
	dgst1 := h1.Sum(nil)
	dgst2 := h2.Sum(nil)
	if !bytes.Equal(dgst1, dgst2) {
		t.Fatal("hashes do not match")
	}

	h3s, ok := h3.(hash.StateStorer)
	if !ok {
		t.Fatal("hash function does not implement stateStorer interface")
	}
	for i := range storedStates {
		if err := h3s.SetState(storedStates[i]); err != nil {
			t.Fatal(err)
		}
		for j := i; j < len(randInputs); j++ {
			h3.Write(randInputs[j].Marshal())
		}
		dgst3 := h3.Sum(nil)
		if !bytes.Equal(dgst1, dgst3) {
			t.Fatal("hashes do not match")
		}
	}
}
