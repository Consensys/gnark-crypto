// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package eddsa

import (
	"io"

	eddsa_bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/twistededwards/eddsa"
	eddsa_bls12381_bandersnatch "github.com/consensys/gnark-crypto/ecc/bls12-381/bandersnatch/eddsa"
	eddsa_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/twistededwards/eddsa"
	eddsa_bls24315 "github.com/consensys/gnark-crypto/ecc/bls24-315/twistededwards/eddsa"
	eddsa_bls24317 "github.com/consensys/gnark-crypto/ecc/bls24-317/twistededwards/eddsa"
	eddsa_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	eddsa_bw6633 "github.com/consensys/gnark-crypto/ecc/bw6-633/twistededwards/eddsa"
	eddsa_bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761/twistededwards/eddsa"
	"github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark-crypto/signature"
)

// New takes a source of randomness and returns a new key pair
func New(ss twistededwards.ID, r io.Reader) (signature.Signer, error) {
	switch ss {
	case twistededwards.BN254:
		return eddsa_bn254.GenerateKey(r)
	case twistededwards.BLS12_381:
		return eddsa_bls12381.GenerateKey(r)
	case twistededwards.BLS12_381_BANDERSNATCH:
		return eddsa_bls12381_bandersnatch.GenerateKey(r)
	case twistededwards.BLS12_377:
		return eddsa_bls12377.GenerateKey(r)
	case twistededwards.BW6_761:
		return eddsa_bw6761.GenerateKey(r)
	case twistededwards.BLS24_315:
		return eddsa_bls24315.GenerateKey(r)
	case twistededwards.BLS24_317:
		return eddsa_bls24317.GenerateKey(r)
	case twistededwards.BW6_633:
		return eddsa_bw6633.GenerateKey(r)
	default:
		panic("not implemented")
	}
}
