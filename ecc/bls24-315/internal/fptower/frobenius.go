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
    a fp.Element
    b fp.Element
    c fp.Element
    d fp.Element
    e fp.Element
    f fp.Element
    g fp.Element
    h fp.Element
    i fp.Element
    j fp.Element
    k fp.Element
    l fp.Element
}

func init() {
    frobCoeff.a = fp.Element{
        18078622854523849680,
        1984927455914812303,
        2087856039593753391,
        10384413649565796150,
        62117205619808039,
    }
    frobCoeff.b = fp.Element{
        16303043965024461850,
        18121025051155953387,
        13066506537353112078,
        1182352075644000644,
        250600645981871933,
    }
    frobCoeff.c = fp.Element{
        7336511025221188090,
        2603771785376329468,
        6562537755091890776,
        9030934061021684028,
        49077327029958380,
    }
    frobCoeff.d = fp.Element{
        2418585497346853195,
        4985300007263584554,
        14090834255033678869,
        8443077587606433532,
        99465464973580433,
    }
    frobCoeff.e = fp.Element{
        16266452697653617742,
        3469624274549632133,
        1532064828142410068,
        17281049885654821422,
        214020232334507350,
    }
    frobCoeff.f = fp.Element{
        319632480799633719,
        12918588655636006616,
        7080179244737088245,
        5761903421758065752,
        223741986209306126,
    }
    frobCoeff.g = fp.Element{
        8984310047302919300,
        2498109052167961353,
        1307418789688509602,
        11960473000634917703,
        283892625570574947,
    }
    frobCoeff.h = fp.Element{
        7656143506020821809,
        15522360441012336084,
        13642716999828979021,
        14792837482779749780,
        272819313239264506,
    }
    frobCoeff.i = fp.Element{
        5276991711591121542,
        1764125630309599080,
        4048361144298871290,
        17215093588476212969,
        305552045589664998,
    }
    frobCoeff.j = fp.Element{
        11164601423358853174,
        17475228851327880835,
        18222098035255651149,
        13126167188689647896,
        69872393236067596,
    }
    frobCoeff.k = fp.Element{
        2851480573204638815,
        1335734525939490983,
        5345966389475061568,
        16856815570427136360,
        235013868839987029,
    }
    frobCoeff.l = fp.Element{
        5645112930776823478,
        18225942248104338392,
        1960505104705117898,
        6830679938910416819,
        243434839969856959,
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
    z.B1.Frobenius(&z.B1).MulByElement(&z.B1, &frobCoeff.a)

	return z
}

// Frobenius sets z in E8 to x^q, returns z
func (z *E8) Frobenius(x *E8) *E8 {

	z.Set(x)
    z.C0.Frobenius(&z.C0)
    z.C1.Frobenius(&z.C1).MulByElement(&z.C1, &frobCoeff.b)

	return z
}

// Frobenius sets z in E24 to x^q, returns z
func (z *E24) Frobenius(x *E24) *E24 {

	z.Set(x)
    z.D0.Frobenius(&z.D0)
    z.D1.Frobenius(&z.D1).MulByElement(&z.D1, &frobCoeff.c)
    z.D2.Frobenius(&z.D2).MulByElement(&z.D2, &frobCoeff.d)

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
    z.C1.FrobeniusSquare(&z.C1).MulByElement(&z.C1, &frobCoeff.a)

	return z
}

// FrobeniusSquare sets z in E24 to x^q2, returns z
func (z *E24) FrobeniusSquare(x *E24) *E24 {

	z.Set(x)
    z.D0.FrobeniusSquare(&z.D0)
    z.D1.FrobeniusSquare(&z.D1).MulByElement(&z.D1, &frobCoeff.d)
    z.D2.FrobeniusSquare(&z.D2).MulByElement(&z.D2, &frobCoeff.e)

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
    z.Conjugate(z).Frobenius(z)

	return z
}

// FrobeniusCube sets z in E8 to x^q3, returns z
func (z *E8) FrobeniusCube(x *E8) *E8 {

	z.Set(x)
    z.C0.FrobeniusCube(&z.C0)
    z.C1.FrobeniusCube(&z.C1).MulByElement(&z.C1, &frobCoeff.f)

	return z
}

// FrobeniusCube sets z in E24 to x^q3, returns z
func (z *E24) FrobeniusCube(x *E24) *E24 {

	z.Set(x)
    z.D0.FrobeniusCube(&z.D0)
    z.D1.FrobeniusCube(&z.D1).MulByElement(&z.D1, &frobCoeff.b)
    z.D2.FrobeniusCube(&z.D2).MulByElement(&z.D2, &frobCoeff.a)

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
    z.D1.FrobeniusQuad(&z.D1).MulByElement(&z.D1, &frobCoeff.e)
    z.D2.FrobeniusQuad(&z.D2).MulByElement(&z.D2, &frobCoeff.g)

	return z
}

// FrobeniusFive sets z in E2 to x^q5, returns z
func (z *E2) FrobeniusFive(x *E2) *E2 {

	z.Set(x)
	z.Conjugate(z)

	return z
}

// FrobeniusFive sets z in E4 to x^q5, returns z
func (z *E4) FrobeniusFive(x *E4) *E4 {

	z.Set(x)
	z.Frobenius(z)

	return z
}

// FrobeniusFive sets z in E8 to x^q5, returns z
func (z *E8) FrobeniusFive(x *E8) *E8 {

	z.Set(x)
	z.Conjugate(z).Frobenius(z)

	return z
}

// FrobeniusFive sets z in E24 to x^q5, returns z
func (z *E24) FrobeniusFive(x *E24) *E24 {

	z.Set(x)
    z.D0.FrobeniusFive(&z.D0)
    z.D1.FrobeniusFive(&z.D1).MulByElement(&z.D1, &frobCoeff.h)
    z.D2.FrobeniusFive(&z.D2).MulByElement(&z.D2, &frobCoeff.i)

	return z
}

// FrobeniusSix sets z in E4 to x^q6, returns z
func (z *E4) FrobeniusSix(x *E4) *E4 {

	z.Set(x)
    z.Conjugate(z)

	return z
}

// FrobeniusSix sets z in E8 to x^q6, returns z
func (z *E8) FrobeniusSix(x *E8) *E8 {

	z.Set(x)
    z.Conjugate(z).FrobeniusSquare(z)

	return z
}

// FrobeniusSix sets z in E24 to x^q6, returns z
func (z *E24) FrobeniusSix(x *E24) *E24 {

	z.Set(x)
    z.D0.FrobeniusSix(&z.D0)
    z.D1.FrobeniusSix(&z.D1).MulByElement(&z.D1, &frobCoeff.a)
    z.D2.FrobeniusSix(&z.D2).MulByElement(&z.D2, &frobCoeff.j)

	return z
}

// FrobeniusSeven sets z in E2 to x^q7, returns z
func (z *E2) FrobeniusSeven(x *E2) *E2 {

	z.Set(x)
    z.Conjugate(z)

	return z
}

// FrobeniusSeven sets z in E4 to x^q7, returns z
func (z *E4) FrobeniusSeven(x *E4) *E4 {

	z.Set(x)
    z.FrobeniusCube(z)

	return z
}

// FrobeniusSeven sets z in E8 to x^q7, returns z
func (z *E8) FrobeniusSeven(x *E8) *E8 {

	z.Set(x)
    z.Conjugate(z).FrobeniusCube(z)

	return z
}

// FrobeniusSeven sets z in E24 to x^q6, returns z
func (z *E24) FrobeniusSeven(x *E24) *E24 {

	z.Set(x)
    z.D0.FrobeniusSeven(&z.D0)
    z.D1.FrobeniusSeven(&z.D1).MulByElement(&z.D1, &frobCoeff.k)
    z.D2.FrobeniusSeven(&z.D2).MulByElement(&z.D2, &frobCoeff.l)

	return z
}
