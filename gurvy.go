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

// Package gurvy is an elliptic curve (+pairing) library. It currently generates efficient code for
// bls381, bls377 and bn256
package gurvy

// do not modify the order of this enum
const (
	UNKNOWN ID = iota
	BLS377
	BLS381
	BN256
)

// ID represent a unique ID for a curve
// (used in serialization checks)
type ID uint16

func (id ID) String() string {
	switch id {
	case BLS377:
		return "bls377"
	case BLS381:
		return "bls381"
	case BN256:
		return "bn256"
	default:
		panic("unimplemented curve ID")
	}
}
