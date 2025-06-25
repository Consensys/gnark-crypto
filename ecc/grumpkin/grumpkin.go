// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package grumpkin efficient elliptic curve and hash to curve implementation for grumpkin. This curve appears forms a 2-cycle with bn254 [https://aztecprotocol.github.io/aztec-connect/primitives.html].
//
// grumpkin: A j=0 curve with
//
//	𝔽r: r=21888242871839275222246405745257275088696311157297823662689037894645226208583
//	𝔽p: p=21888242871839275222246405745257275088548364400416034343698204186575808495617
//	(E/𝔽p): Y²=X³-17
//	r ∣ #E(Fp)
//
// Security: estimated 127-bit level against Pollard's Rho attack
// (r is 254 bits)
//
// # Warning
//
// This code has been partially audited and is provided as-is. In particular, there is no security guarantees such as constant time implementation or side-channel attack resistance.
package grumpkin

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/grumpkin/fp"
	"github.com/consensys/gnark-crypto/ecc/grumpkin/fr"
)

// ID grumpkin ID
const ID = ecc.GRUMPKIN

// aCurveCoeff is the a coefficients of the curve Y²=X³+ax+b
var aCurveCoeff fp.Element
var bCurveCoeff fp.Element

// generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
var g1Gen G1Jac

var g1GenAff G1Affine

// point at infinity
var g1Infinity G1Jac

// Parameters useful for the GLV scalar multiplication. The third roots define the
// endomorphisms ϕ₁ for <G1Affine>. lambda is such that <r, ϕ-λ> lies above
// <r> in the ring Z[ϕ]. More concretely it's the associated eigenvalue
// of ϕ₁ restricted to <G1Affine>.
// see https://link.springer.com/content/pdf/10.1007/3-540-36492-7_3
var thirdRootOneG1 fp.Element
var lambdaGLV big.Int

// seed x₀ of the BN254 curve
var xGen big.Int

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
	bCurveCoeff.SetUint64(17).Neg(&bCurveCoeff)

	g1Gen.X.SetOne()
	g1Gen.Y.SetString("17631683881184975370165255887551781615748388533673675138860") // sqrt(-16) % p
	g1Gen.Z.SetOne()

	g1GenAff.FromJacobian(&g1Gen)

	// (X,Y,Z) = (1,1,0)
	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()

	thirdRootOneG1.SetString("4407920970296243842393367215006156084916469457145843978461")
	lambdaGLV.SetString("2203960485148121921418603742825762020974279258880205651966", 10)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)
	g1ScalarMulChoose = fr.Bits/16 + max(glvBasis.V1[0].BitLen(), glvBasis.V1[1].BitLen(), glvBasis.V2[0].BitLen(), glvBasis.V2[1].BitLen())
	g2ScalarMulChoose = fr.Bits/32 + max(glvBasis.V1[0].BitLen(), glvBasis.V1[1].BitLen(), glvBasis.V2[0].BitLen(), glvBasis.V2[1].BitLen())

	xGen.SetString("4965661367192848881", 10)
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
