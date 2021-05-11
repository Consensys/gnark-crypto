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

// Package polynomial provides interfaces for polynomial commitment schemes defined in gnark-crypto/ecc/.../fr.
package polynomial

// Polynomial interface that a polynomial should implement
type Polynomial interface {

	// Degree returns the degree of the polynomial
	Degree() uint64

	// Eval computes the evaluation of the polynomial at v
	Eval(v interface{}) interface{}

	// Returns a copy of the polynomial
	Clone() Polynomial

	// Add adds p1 to p, modifying p
	Add(p1, p2 Polynomial) Polynomial

	// AddConstantInPlace adds a constant to the polynomial, modifying p
	AddConstantInPlace(c interface{})

	// AddConstantInPlace subs a constant to the polynomial, modifying p
	SubConstantInPlace(c interface{})

	// ScaleInPlace multiplies the polynomial by a constant c, modifying p
	ScaleInPlace(c interface{})

	// Equal checks equality between two polynomials
	Equal(p1 Polynomial) bool
}
