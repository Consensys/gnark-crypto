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

// Package secq256k1 efficient elliptic curve implementation for secq256k1. This curve forms a 2-cycle with secp256k1 which is defined in Standards for Efficient Cryptography (SEC) (Certicom Research, http://www.secg.org/sec2-v2.pdf) and appears in the Bitcoin and Ethereum ECDSA signatures.
//
// secq256k1: A j=0 curve with
//
//	𝔽r: p=115792089237316195423570985008687907853269984665640564039457584007908834671663 (2^256 - 2^32 - 977)
//	𝔽p: r=115792089237316195423570985008687907852837564279074904382605163141518161494337
//	(E/𝔽p): Y²=X³+7
//
// Security: estimated 128-bit level using Pollard's \rho attack
// (r is 256 bits)
//
// # Warning
//
// This code has been partially audited and is provided as-is. In particular, there is no security guarantees such as constant time implementation or side-channel attack resistance.
package secq256k1

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/secq256k1/fp"
	"github.com/consensys/gnark-crypto/ecc/secq256k1/fr"
)

// ID secq256k1 ID
const ID = ecc.SECQ256K1

// bCurveCoeff b coeff of the curve Y²=X³+b
var bCurveCoeff fp.Element

// generator of the r-torsion group
var g1Gen G1Jac

var g1GenAff G1Affine

// point at infinity
var g1Infinity G1Jac

// Parameters useful for the GLV scalar multiplication. The third roots define the
// endomorphisms ϕ₁ for <G1Affine>. lambda is such that <r, ϕ-λ> lies above
// <r> in the ring Z[ϕ]. More concretely it's the associated eigenvalue
// of ϕ₁ restricted to <G1Affine>
// see https://www.cosic.esat.kuleuven.be/nessie/reports/phase2/GLV.pdf
var thirdRootOneG1 fp.Element
var lambdaGLV big.Int

// glvBasis stores R-linearly independent vectors (a,b), (c,d)
// in ker((u,v) → u+vλ[r]), and their determinant
var glvBasis ecc.Lattice

func init() {

	bCurveCoeff.SetUint64(7)

	g1Gen.X.SetString("53718550993811904772965658690407829053653678808745171666022356150019200052646")
	g1Gen.Y.SetString("28941648020349172432234515805717979317553499307621291159490218670604692907903")
	g1Gen.Z.SetOne()

	g1GenAff.FromJacobian(&g1Gen)

	// (X,Y,Z) = (1,1,0)
	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()

	thirdRootOneG1.SetString("78074008874160198520644763525212887401909906723592317393988542598630163514318") // 2^((p-1)/3)
	lambdaGLV.SetString("60197513588986302554485582024885075108884032450952339817679072026166228089408", 10)  // 3^((r-1)/3)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

}

// Generators return the generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
func Generators() (g1Jac G1Jac, g1Aff G1Affine) {
	g1Aff = g1GenAff
	g1Jac = g1Gen
	return
}
