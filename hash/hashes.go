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

package hash

import (
	"hash"
)

var hashes = make([]func() hash.Hash, maxHash)

// RegisterHash registers a new hash function constructor. Should be called in
// the init function of the hash package.
//
// To register all known hash functions in gnark-crypto, import the
// [github.com/consensys/gnark-crypto/hash/all] package in your code.
func RegisterHash(h Hash, new func() hash.Hash) {
	hashes[h] = new
}

// Hash defines an unique identifier for a hash function.
type Hash uint

const (
	// MIMC_BN254 is the MiMC hash function for the BN254 curve.
	MIMC_BN254 Hash = iota
	// MIMC_BLS12_381 is the MiMC hash function for the BLS12-381 curve.
	MIMC_BLS12_381
	// MIMC_BLS12_377 is the MiMC hash function for the BLS12-377 curve.
	MIMC_BLS12_377
	// MIMC_BW6_761 is the MiMC hash function for the BW6-761 curve.
	MIMC_BW6_761
	// MIMC_BLS24_315 is the MiMC hash function for the BLS24-315 curve.
	MIMC_BLS24_315
	// MIMC_BLS24_317 is the MiMC hash function for the BLS24-317 curve.
	MIMC_BLS24_317
	// MIMC_BW6_633 is the MiMC hash function for the BW6-633 curve.
	MIMC_BW6_633

	maxHash
)

// size of digests in bytes
var digestSize = []uint8{
	MIMC_BN254:     32,
	MIMC_BLS12_381: 48,
	MIMC_BLS12_377: 48,
	MIMC_BW6_761:   96,
	MIMC_BLS24_315: 48,
	MIMC_BLS24_317: 48,
	MIMC_BW6_633:   80,
}

// New initializes the hash function. This is a convenience function which does
// not allow setting hash-specific options.
func (m Hash) New() hash.Hash {
	if m < maxHash {
		f := hashes[m]
		if f != nil {
			return f()
		}
	}
	panic("requested hash function #" + m.String() + " not registered")
}

// String returns the unique identifier of the hash function.
func (m Hash) String() string {
	switch m {
	case MIMC_BN254:
		return "MIMC_BN254"
	case MIMC_BLS12_381:
		return "MIMC_BLS381"
	case MIMC_BLS12_377:
		return "MIMC_BLS377"
	case MIMC_BW6_761:
		return "MIMC_BW761"
	case MIMC_BLS24_315:
		return "MIMC_BLS315"
	case MIMC_BLS24_317:
		return "MIMC_BLS317"
	case MIMC_BW6_633:
		return "MIMC_BW633"
	default:
		return "unknown hash function"
	}
}

// Available returns true if the hash function is available.
func (m Hash) Available() bool {
	return m < maxHash && hashes[m] != nil
}

// Size returns the size of the digest of the corresponding hash function
func (m Hash) Size() int {
	return int(digestSize[m])
}
