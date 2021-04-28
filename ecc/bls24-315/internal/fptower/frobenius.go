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
    cTwo fp.Element
    dTwo fp.Element
    eTwo fp.Element
    bThree fp.Element
    cThree fp.Element
    dThree fp.Element
    eThree fp.Element
    dQuad fp.Element
    eQuad fp.Element
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
    frobCoeff.bThree = fp.Element{
        8431819647309378609,
        2779570725743559026,
        13963483320145043377,
        4889343876951054201,
        280783099323629353,
    }
    frobCoeff.cThree = fp.Element{
        319632480799633719,
        12918588655636006616,
        7080179244737088245,
        5761903421758065752,
        223741986209306126,
    }
    frobCoeff.dThree = fp.Element{
        16303043965024461850,
        18121025051155953387,
        13066506537353112078,
        1182352075644000644,
        250600645981871933,
    }
    frobCoeff.eThree = fp.Element{
        18078622854523849680,
        1984927455914812303,
        2087856039593753391,
        10384413649565796150,
        62117205619808039,
    }
    frobCoeff.dQuad = fp.Element{
        16266452697653617742,
        3469624274549632133,
        1532064828142410068,
        17281049885654821422,
        214020232334507350,
    }
    frobCoeff.eQuad = fp.Element{
        8984310047302919300,
        2498109052167961353,
        1307418789688509602,
        11960473000634917703,
        283892625570574947,
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
    z.Conjugate(z)

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

// FrobeniusCube sets z in E2 to x^q3, returns z
func (z *E2) FrobeniusCube(x *E2) *E2 {

	z.Set(x)
    z.Conjugate(z)

	return z
}

// FrobeniusCube sets z in E4 to x^q3, returns z
func (z *E4) FrobeniusCube(x *E4) *E4 {

	z.Set(x)
    z.B0.FrobeniusCube(&z.B0)
    z.B1.FrobeniusCube(&z.B1).MulByElement(&z.B1, &frobCoeff.bThree)

	return z
}

// FrobeniusCube sets z in E8 to x^q3, returns z
func (z *E8) FrobeniusCube(x *E8) *E8 {

	z.Set(x)
    z.C0.FrobeniusCube(&z.C0)
    z.C1.FrobeniusCube(&z.C1).MulByElement(&z.C1, &frobCoeff.cThree)

	return z
}

// FrobeniusCube sets z in E24 to x^q3, returns z
func (z *E24) FrobeniusCube(x *E24) *E24 {

	z.Set(x)
    z.D0.FrobeniusCube(&z.D0)
    z.D1.FrobeniusCube(&z.D1).MulByElement(&z.D1, &frobCoeff.dThree)
    z.D2.FrobeniusCube(&z.D2).MulByElement(&z.D2, &frobCoeff.eThree)

	return z
}

// FrobeniusQuad sets z in E8 to x^q4, returns z
func (z *E8) FrobeniusQuad(x *E8) *E8 {

	z.Set(x)
	z.Conjugate(z)

	return z
}

// FrobeniusQuad sets z in E24 to x^q4, returns z
func (z *E24) FrobeniusQuad(x *E24) *E24 {

	z.Set(x)
    z.D0.FrobeniusQuad(&z.D0)
    z.D1.FrobeniusQuad(&z.D1).MulByElement(&z.D1, &frobCoeff.dQuad)
    z.D2.FrobeniusQuad(&z.D2).MulByElement(&z.D2, &frobCoeff.eQuad)

	return z
}
