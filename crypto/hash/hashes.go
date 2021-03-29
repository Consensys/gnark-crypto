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

// Package hash gathers the different custom hash functions (which
// are not available in golang's crypto library) built for being
// within a ZKP scheme. The structure of the package is similar to what
// can be found in golang's crypto/ package.
package hash

import (
	"hash"

	bls377 "github.com/consensys/gurvy/curve/bls377/fr/mimc"
	bls381 "github.com/consensys/gurvy/curve/bls381/fr/mimc"
	bn254 "github.com/consensys/gurvy/curve/bn254/fr/mimc"
	bw761 "github.com/consensys/gurvy/curve/bw761/fr/mimc"
)

type Hash uint

const (
	MIMC_BN254 Hash = iota
	MIMC_BLS381
	MIMC_BLS377
	MIMC_BW761
)

// size of digests in bytes
var digestSize = []uint8{
	MIMC_BN254:  32,
	MIMC_BLS381: 48,
	MIMC_BLS377: 48,
	MIMC_BW761:  96,
}

// New creates the corresponding mimc hash function.
func (m Hash) New(seed string) hash.Hash {
	switch m {
	case MIMC_BN254:
		return bn254.NewMiMC(seed)
	case MIMC_BLS381:
		return bls381.NewMiMC(seed)
	case MIMC_BLS377:
		return bls377.NewMiMC(seed)
	case MIMC_BW761:
		return bw761.NewMiMC(seed)
	default:
		panic("Unknown mimc ID")
	}
}

// String returns the mimc ID to string format.
func (m Hash) String() string {
	switch m {
	case MIMC_BN254:
		return "MIMC_BN254"
	case MIMC_BLS381:
		return "MIMC_BLS381"
	case MIMC_BLS377:
		return "MIMC_BLS377"
	case MIMC_BW761:
		return "MIMC_BW761"
	default:
		panic("Unknown mimc ID")
	}
}

// Size returns the size of the digest of
// the corresponding hash function
func (m Hash) Size() int {
	return int(digestSize[m])
}
