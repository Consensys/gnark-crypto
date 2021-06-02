// +build gofuzz

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

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fr

// MulGeneric is a wrapper exposed and used for fuzzing purposes only
func MulGeneric(z, x, y *Element) {
	_mulGeneric(z, x, y)
}

// FromMontGeneric is a wrapper exposed and used for fuzzing purposes only
func FromMontGeneric(z *Element) {
	_fromMontGeneric(z)
}

// AddGeneric is a wrapper exposed and used for fuzzing purposes only
func AddGeneric(z, x, y *Element) {
	_addGeneric(z, x, y)
}

// DoubleGeneric is a wrapper exposed and used for fuzzing purposes only
func DoubleGeneric(z, x *Element) {
	_doubleGeneric(z, x)
}

// SubGeneric is a wrapper exposed and used for fuzzing purposes only
func SubGeneric(z, x, y *Element) {
	_subGeneric(z, x, y)
}

// NegGeneric is a wrapper exposed and used for fuzzing purposes only
func NegGeneric(z, x *Element) {
	_negGeneric(z, x)
}

// ReduceGeneric is a wrapper exposed and used for fuzzing purposes only
func ReduceGeneric(z *Element) {
	_reduceGeneric(z)
}
