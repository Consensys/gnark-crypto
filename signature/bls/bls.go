/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bls

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc"
	bls_bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/signatures/bls"
	bls_bls12378 "github.com/consensys/gnark-crypto/ecc/bls12-378/signatures/bls"
	bls_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/signatures/bls"
	bls_bls24315 "github.com/consensys/gnark-crypto/ecc/bls24-315/signatures/bls"
	bls_bls24317 "github.com/consensys/gnark-crypto/ecc/bls24-317/signatures/bls"
	bls_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/signatures/bls"
	bls_bw6633 "github.com/consensys/gnark-crypto/ecc/bw6-633/signatures/bls"
	bls_bw6756 "github.com/consensys/gnark-crypto/ecc/bw6-756/signatures/bls"
	bls_bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761/signatures/bls"
	"github.com/consensys/gnark-crypto/signature"
)

// New takes a source of randomness and returns a new key pair
func New(ss ecc.ID, r io.Reader) (signature.Signer, error) {
	switch ss {
	case ecc.BN254:
		return bls_bn254.GenerateKey(r)
	case ecc.BLS12_381:
		return bls_bls12381.GenerateKey(r)
	case ecc.BLS12_377:
		return bls_bls12377.GenerateKey(r)
	case ecc.BLS12_378:
		return bls_bls12378.GenerateKey(r)
	case ecc.BW6_761:
		return bls_bw6761.GenerateKey(r)
	case ecc.BW6_756:
		return bls_bw6756.GenerateKey(r)
	case ecc.BLS24_315:
		return bls_bls24315.GenerateKey(r)
	case ecc.BLS24_317:
		return bls_bls24317.GenerateKey(r)
	case ecc.BW6_633:
		return bls_bw6633.GenerateKey(r)
	default:
		panic("not implemented")
	}
}
