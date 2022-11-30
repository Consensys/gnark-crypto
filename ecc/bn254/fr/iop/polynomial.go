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

package iop

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// Enum to tell in which basis a polynomial is represented.
type Basis int64

const (
	Canonical Basis = iota
	Lagrange
	LagrangeCoset
)

// Enum to tell if a polynomial is in bit reverse form or
// in the regular form.
type Layout int64

const (
	Regular Layout = iota
	BitReverse
)

// Enum to tell if the polynomial can be modified.
// If the polynomial can not be modified, then whenever
// a function has to do a transformation on it (FFT, bitReverse, etc)
// then a new vector is allocated.
type Status int64

const (
	Locked Status = iota
	Unlocked
)

// Form describes the form of a polynomial.
type Form struct {
	Basis  Basis
	Layout Layout
	Status Status
}

// Polynomial represents a polynomial, the vector of coefficients
// along with the basis and the layout.
type Polynomial struct {
	Coefficients []fr.Element
	Info         Form
}
