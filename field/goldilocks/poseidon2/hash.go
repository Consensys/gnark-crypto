// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package poseidon2

import (
	fr "github.com/consensys/gnark-crypto/field/goldilocks"
	gnarkHash "github.com/consensys/gnark-crypto/hash"
	"hash"
	"sync"
)

// NewMerkleDamgardHasher returns a Poseidon2 hasher using the Merkle-Damgard
// construction with the default parameters.
func NewMerkleDamgardHasher() gnarkHash.StateStorer {
	return gnarkHash.NewMerkleDamgardHasher(
		&Permutation{GetDefaultParameters()}, make([]byte, fr.Bytes))
}

// GetDefaultParameters returns a set of parameters for the Poseidon2 permutation.
// The default parameters are,
//
//  1. for compression:
//     - width: 8
//     - nbFullRounds: 6
//     - nbPartialRounds: 17
//
//  2. for sponge:
//     - width: 12
//     - nbFullRounds: 6
//     - nbPartialRounds: 17
var GetDefaultParameters = sync.OnceValue(func() *Parameters {
	return NewParameters(8, 6, 17)
})

var diag8 []fr.Element = make([]fr.Element, 8)
var diag12 []fr.Element = make([]fr.Element, 12)

func init() {
	// diagnoal diag8 for the internal diagonal of the matrix of the compression layer
	diag8[0].SetUint64(12216033376705242021)
	diag8[1].SetUint64(2072934925475504800)
	diag8[2].SetUint64(16432743296706583078)
	diag8[3].SetUint64(1287600597097751715)
	diag8[4].SetUint64(10482065724875379356)
	diag8[5].SetUint64(3057917794534811537)
	diag8[6].SetUint64(4460508886913832365)
	diag8[7].SetUint64(4574242228824269566)

	// diagnoal diag12 for the internal diagonal of the matrix of the sponge layer
	diag12[0].SetUint64(14102670999874605824)
	diag12[1].SetUint64(15585654191999307702)
	diag12[2].SetUint64(940187017142450255)
	diag12[3].SetUint64(8747386241522630711)
	diag12[4].SetUint64(6750641561540124747)
	diag12[5].SetUint64(7440998025584530007)
	diag12[6].SetUint64(6136358134615751536)
	diag12[7].SetUint64(12413576830284969611)
	diag12[8].SetUint64(11675438539028694709)
	diag12[9].SetUint64(17580553691069642926)
	diag12[10].SetUint64(892707462476851331)
	diag12[11].SetUint64(15167485180850043744)

	gnarkHash.RegisterHash(gnarkHash.POSEIDON2_GOLDILOCKS, func() hash.Hash {
		return NewMerkleDamgardHasher()
	})
}
