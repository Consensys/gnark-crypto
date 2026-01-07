// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package ecdsa

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc"
	ecdsa_bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/ecdsa"
	ecdsa_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/ecdsa"
	ecdsa_bls24315 "github.com/consensys/gnark-crypto/ecc/bls24-315/ecdsa"
	ecdsa_bls24317 "github.com/consensys/gnark-crypto/ecc/bls24-317/ecdsa"
	ecdsa_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/ecdsa"
	ecdsa_bw6633 "github.com/consensys/gnark-crypto/ecc/bw6-633/ecdsa"
	ecdsa_bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761/ecdsa"
	ecdsa_secp256k1 "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	ecdsa_starkcurve "github.com/consensys/gnark-crypto/ecc/stark-curve/ecdsa"
	"github.com/consensys/gnark-crypto/signature"
)

// New takes a source of randomness and returns a new key pair
func New(ss ecc.ID, r io.Reader) (signature.Signer, error) {
	switch ss {
	case ecc.BN254:
		return ecdsa_bn254.GenerateKey(r)
	case ecc.BLS12_381:
		return ecdsa_bls12381.GenerateKey(r)
	case ecc.BLS12_377:
		return ecdsa_bls12377.GenerateKey(r)
	case ecc.BW6_761:
		return ecdsa_bw6761.GenerateKey(r)
	case ecc.BLS24_315:
		return ecdsa_bls24315.GenerateKey(r)
	case ecc.BLS24_317:
		return ecdsa_bls24317.GenerateKey(r)
	case ecc.BW6_633:
		return ecdsa_bw6633.GenerateKey(r)
	case ecc.SECP256K1:
		return ecdsa_secp256k1.GenerateKey(r)
	case ecc.STARK_CURVE:
		return ecdsa_starkcurve.GenerateKey(r)
	default:
		panic("not implemented")
	}
}
