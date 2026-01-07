// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package secp256k1 efficient elliptic curve implementation for secp256k1. This curve is defined in Standards for Efficient Cryptography (SEC) (Certicom Research, http://www.secg.org/sec2-v2.pdf) and appears in the Bitcoin and Ethereum ECDSA signatures.
//
// secp256k1: A j=0 curve with
//
//	𝔽r: r=115792089237316195423570985008687907852837564279074904382605163141518161494337
//	𝔽p: p=115792089237316195423570985008687907853269984665640564039457584007908834671663 (2^256 - 2^32 - 977)
//	(E/𝔽p): Y²=X³+7
//
// Security: estimated 128-bit level using Pollard's \rho attack
// (r is 256 bits)
//
// # Warning
//
// This code has been partially audited and is provided as-is. In particular, there is no security guarantees such as constant time implementation or side-channel attack resistance.
package secp256k1

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/secp256k1/fp"
	"github.com/consensys/gnark-crypto/ecc/secp256k1/fr"
)

// ID secp256k1 ID
const ID = ecc.SECP256K1

// aCurveCoeff is the a coefficients of the curve Y²=X³+ax+b
var aCurveCoeff fp.Element
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
// see https://link.springer.com/content/pdf/10.1007/3-540-36492-7_3
var thirdRootOneG1 fp.Element
var lambdaGLV big.Int

// glvBasis stores R-linearly independent vectors (a,b), (c,d)
// in ker((u,v) → u+vλ[r]), and their determinant
var glvBasis ecc.Lattice

// g1ScalarMulChoose and g2ScalarmulChoose indicate the bitlength of the scalar
// in scalar multiplication from which it is more efficient to use the GLV
// decomposition. It is computed from the GLV basis and considers the overhead
// for the GLV decomposition. It is heuristic and may change in the future.
var g1ScalarMulChoose, g2ScalarMulChoose int

func init() {
	aCurveCoeff.SetUint64(0)
	bCurveCoeff.SetUint64(7)

	g1Gen.X.SetString("55066263022277343669578718895168534326250603453777594175500187360389116729240")
	g1Gen.Y.SetString("32670510020758816978083085130507043184471273380659243275938904335757337482424")
	g1Gen.Z.SetOne()

	g1GenAff.FromJacobian(&g1Gen)

	// (X,Y,Z) = (1,1,0)
	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()

	thirdRootOneG1.SetString("55594575648329892869085402983802832744385952214688224221778511981742606582254") // 2^((p-1)/3)
	lambdaGLV.SetString("37718080363155996902926221483475020450927657555482586988616620542887997980018", 10)  // 3^((r-1)/3)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)
	g1ScalarMulChoose = fr.Bits/16 + max(glvBasis.V1[0].BitLen(), glvBasis.V1[1].BitLen(), glvBasis.V2[0].BitLen(), glvBasis.V2[1].BitLen())
	g2ScalarMulChoose = fr.Bits/32 + max(glvBasis.V1[0].BitLen(), glvBasis.V1[1].BitLen(), glvBasis.V2[0].BitLen(), glvBasis.V2[1].BitLen())
}

// Generators return the generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
func Generators() (g1Jac G1Jac, g1Aff G1Affine) {
	g1Aff = g1GenAff
	g1Jac = g1Gen
	return
}

// CurveCoefficients returns the a, b coefficients of the curve equation.
func CurveCoefficients() (a, b fp.Element) {
	return aCurveCoeff, bCurveCoeff
}
