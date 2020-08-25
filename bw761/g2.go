// Copyright 2020 ConsenSys AG
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

package bw761

import "github.com/consensys/gurvy/bw761/fp"

// G2Jac same type as G1Jac (degree 6 twist, degree 6 extension)
type G2Jac = G1Jac

// G2Proj same type as G1Proj (degree 6 twist, degree 6 extension)
type G2Proj = G1Proj

// G2Affine same type as G1Affine (degree 6 twist, degree 6 extension)
type G2Affine = G1Affine

// g2JacExtended same type as g1JacExtended (degree 6 twist, degree 6 extension)
type g2JacExtended = g1JacExtended

// IsOnTwist returns true if p in on the curve
func (p *G2Proj) IsOnTwist() bool {
	var left, right, tmp fp.Element
	left.Square(&p.Y).
		Mul(&left, &p.Z)
	right.Square(&p.X).
		Mul(&right, &p.X)
	tmp.Square(&p.Z).
		Mul(&tmp, &p.Z).
		Mul(&tmp, &Btwist)
	right.Add(&right, &tmp)
	return left.Equal(&right)
}

// IsOnTwist returns true if p in on the curve
func (p *G2Jac) IsOnTwist() bool {
	var left, right, tmp fp.Element
	left.Square(&p.Y)
	right.Square(&p.X).Mul(&right, &p.X)
	tmp.Square(&p.Z).
		Square(&tmp).
		Mul(&tmp, &p.Z).
		Mul(&tmp, &p.Z).
		Mul(&tmp, &Btwist)
	right.Add(&right, &tmp)
	return left.Equal(&right)
}

// IsOnTwist returns true if p in on the curve
func (p *G2Affine) IsOnTwist() bool {
	var point G2Jac
	point.FromAffine(p)
	return point.IsOnTwist() // call this function to handle infinity point
}
