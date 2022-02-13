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

package bls24315

import (
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"testing"
)

func TestG1SqrtRatio(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	properties := gopter.NewProperties(parameters)
	gen := genCoordPairG1(t)

	properties.Property("G1SqrtRatio must square back to the right value", prop.ForAll(
		func(uv []fp.Element) bool {
			u := &uv[0]
			v := &uv[1]

			var ref fp.Element
			ref.Div(u, v)
			var qrRef bool
			if ref.Legendre() == -1 {
				var Z fp.Element
				g1SetZ(&Z)
				ref.Mul(&ref, &Z)
				qrRef = false
			} else {
				qrRef = true
			}

			var seen fp.Element
			qr := g1SqrtRatio(&seen, u, v) == 0
			seen.Square(&seen)

			// Allowing qr(0)=false because the generic algorithm "for any field" seems to think so
			return seen == ref && (ref.IsZero() || qr == qrRef)

		}, gen))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func genCoordPairG1(t *testing.T) gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {

		genRandomPair := func() (fp.Element, fp.Element) {
			var a, b fp.Element

			if _, err := a.SetRandom(); err != nil {
				t.Error(err)
			}

			if _, err := b.SetRandom(); err != nil {
				t.Error(err)
			}

			return a, b
		}
		a, b := genRandomPair()

		genResult := gopter.NewGenResult([]fp.Element{a, b}, gopter.NoShrinker)
		return genResult
	}
}
