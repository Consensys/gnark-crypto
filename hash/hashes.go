// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash

import (
	"hash"

	bls377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/mimc"
	bls381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr/mimc"
	bls315 "github.com/consensys/gnark-crypto/ecc/bls24-315/fr/mimc"
	bls317 "github.com/consensys/gnark-crypto/ecc/bls24-317/fr/mimc"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	bw633 "github.com/consensys/gnark-crypto/ecc/bw6-633/fr/mimc"
	bw761 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/mimc"
)

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

// New initializes the hash function.
func (m Hash) New() hash.Hash {
	switch m {
	case MIMC_BN254:
		return bn254.NewMiMC()
	case MIMC_BLS12_381:
		return bls381.NewMiMC()
	case MIMC_BLS12_377:
		return bls377.NewMiMC()
	case MIMC_BW6_761:
		return bw761.NewMiMC()
	case MIMC_BLS24_315:
		return bls315.NewMiMC()
	case MIMC_BLS24_317:
		return bls317.NewMiMC()
	case MIMC_BW6_633:
		return bw633.NewMiMC()
	default:
		panic("Unknown mimc ID")
	}
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
		panic("Unknown mimc ID")
	}
}

// Size returns the size of the digest of
// the corresponding hash function
func (m Hash) Size() int {
	return int(digestSize[m])
}
