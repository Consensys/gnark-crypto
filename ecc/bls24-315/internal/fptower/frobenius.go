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

package fptower

import "github.com/consensys/gnark-crypto/ecc/bls24-315/fp"

var frobCoeff struct {
    bOne fp.Element
    cOne fp.Element
    dOne fp.Element
    eOne fp.Element
    bTwo fp.Element
    cTwo fp.Element
    dTwo fp.Element
    eTwo fp.Element
}

func init() {
    frobCoeff.bOne = fp.Element{
        18078622854523849680,
        1984927455914812303,
        2087856039593753391,
        10384413649565796150,
        62117205619808039,
    }
    frobCoeff.cOne = fp.Element{
        16303043965024461850,
        18121025051155953387,
        13066506537353112078,
        1182352075644000644,
        250600645981871933,
    }
    frobCoeff.dOne = fp.Element{
        7336511025221188090,
        2603771785376329468,
        6562537755091890776,
        9030934061021684028,
        49077327029958380,
    }
    frobCoeff.eOne = fp.Element{
        2418585497346853195,
        4985300007263584554,
        14090834255033678869,
        8443077587606433532,
        99465464973580433,
    }
    frobCoeff.bTwo = fp.Element{
        11164601423358853174,
        17475228851327880835,
        18222098035255651149,
        13126167188689647896,
        69872393236067596,
    }
    frobCoeff.cTwo = fp.Element{
        18078622854523849680,
        1984927455914812303,
        2087856039593753391,
        10384413649565796150,
        62117205619808039,
    }
    frobCoeff.dTwo = fp.Element{
        2418585497346853195,
        4985300007263584554,
        14090834255033678869,
        8443077587606433532,
        99465464973580433,
    }
    frobCoeff.eTwo = fp.Element{
        16266452697653617742,
        3469624274549632133,
        1532064828142410068,
        17281049885654821422,
        214020232334507350,
    }
}

// Frobenius sets z in E2 to x^q, returns z
func (z *E2) Frobenius(x *E2) *E2 {

	z.Set(x)
	z.Conjugate(z)

	return z
}

// Frobenius sets z in E4 to x^q, returns z
func (z *E4) Frobenius(x *E4) *E4 {

	z.Set(x)
    z.B0.Frobenius(&z.B0)
    z.B1.Frobenius(&z.B1).MulByElement(&z.B1, &frobCoeff.bOne)

	return z
}

// Frobenius sets z in E8 to x^q, returns z
func (z *E8) Frobenius(x *E8) *E8 {

	z.Set(x)
    z.C0.Frobenius(&z.C0)
    z.C1.Frobenius(&z.C1).MulByElement(&z.C1, &frobCoeff.cOne)

	return z
}

// Frobenius sets z in E24 to x^q, returns z
func (z *E24) Frobenius(x *E24) *E24 {

	z.Set(x)
    z.D0.Frobenius(&z.D0)
    z.D1.Frobenius(&z.D1).MulByElement(&z.D1, &frobCoeff.dOne)
    z.D2.Frobenius(&z.D2).MulByElement(&z.D2, &frobCoeff.eOne)

	return z
}

// FrobeniusSquare sets z in E4 to x^q2, returns z
func (z *E4) FrobeniusSquare(x *E4) *E4 {

	z.Set(x)
    z.B1.MulByElement(&z.B1, &frobCoeff.bTwo)

	return z
}

// FrobeniusSquare sets z in E8 to x^q2, returns z
func (z *E8) FrobeniusSquare(x *E8) *E8 {

	z.Set(x)
    z.C0.FrobeniusSquare(&z.C0)
    z.C1.FrobeniusSquare(&z.C1).MulByElement(&z.C1, &frobCoeff.cTwo)

	return z
}

// FrobeniusSquare sets z in E24 to x^q2, returns z
func (z *E24) FrobeniusSquare(x *E24) *E24 {

	z.Set(x)
    z.D0.FrobeniusSquare(&z.D0)
    z.D1.FrobeniusSquare(&z.D1).MulByElement(&z.D1, &frobCoeff.dTwo)
    z.D2.FrobeniusSquare(&z.D2).MulByElement(&z.D2, &frobCoeff.eTwo)

	return z
}
