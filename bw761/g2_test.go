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

import (
	"testing"

	"github.com/consensys/gurvy/bw761/fp"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestG2IsOnCurve(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)
	genFuzz1 := GenFp()
	properties.Property("g2Gen (affine) should be on the curve", prop.ForAll(
		func(a fp.Element) bool {
			var op1, op2 G2Affine
			op1.FromJacobian(&g2Gen)
			op2.FromJacobian(&g2Gen)
			op2.Y.Mul(&op2.Y, &a)
			return op1.IsOnTwist() && !op2.IsOnTwist()
		},
		genFuzz1,
	))

	properties.Property("g2Gen (Jacobian) should be on the curve", prop.ForAll(
		func(a fp.Element) bool {
			var op1, op2, op3 G2Jac
			op1.Set(&g2Gen)
			op3.Set(&g2Gen)

			op2 = fuzzJacobianG1(&g2Gen, a)
			op3.Y.Mul(&op3.Y, &a)
			return op1.IsOnTwist() && op2.IsOnTwist() && !op3.IsOnTwist()
		},
		genFuzz1,
	))

	properties.Property("g2Gen (projective) should be on the curve", prop.ForAll(
		func(a fp.Element) bool {
			var op1, op2, op3 G2Proj
			op1.FromJacobian(&g2Gen)
			op2.FromJacobian(&g2Gen)
			op3.FromJacobian(&g2Gen)

			op2 = fuzzProjectiveG1(&op1, a)
			op3.Y.Mul(&op3.Y, &a)
			return op1.IsOnTwist() && op2.IsOnTwist() && !op3.IsOnTwist()
		},
		genFuzz1,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
