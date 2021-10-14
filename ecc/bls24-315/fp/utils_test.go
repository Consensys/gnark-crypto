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

package fp

import (
	"math/big"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestElementHalve(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	genA := gen()

	properties.Property("Halve should be equal to Mul by 1/2", prop.ForAll(
		func(a testPairElement) bool {
			var c, d, twoInv Element
			twoInv.SetOne().Double(&twoInv).Inverse(&twoInv)
			c.Halve(&a.element)
			d.Mul(&a.element, &twoInv)

			return c.Equal(&d)
		},
		genA,
	))

	specialValueTest := func() {
		// test special values
		testValues := make([]Element, len(staticTestValues))
		copy(testValues, staticTestValues)

		for _, a := range testValues {
			var aBig big.Int
			a.ToBigIntRegular(&aBig)
			var c Element
			c.Square(&a)

			var d, e big.Int
			d.Mul(&aBig, &aBig).Mod(&d, Modulus())

			if c.FromMont().ToBigInt(&e).Cmp(&d) != 0 {
				t.Fatal("Square failed special test values")
			}
		}
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
	specialValueTest()
	// if we have ADX instruction enabled, test both path in assembly
	if supportAdx {
		t.Log("disabling ADX")
		supportAdx = false
		properties.TestingRun(t, gopter.ConsoleReporter(false))
		specialValueTest()
		supportAdx = true
	}
}
