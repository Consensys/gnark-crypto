// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://wwwApache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fptower

import "github.com/consensys/gnark-crypto/ecc/bls24-315/fp"

// Frobenius sets z in E4 to x^q, returns z
func (z *E4) Frobenius(x *E4) *E4 {

	var t [2]E2

	frobCoeffA := fp.Element{
		18078622854523849680,
		1984927455914812303,
		2087856039593753391,
		10384413649565796150,
		62117205619808039,
	}

	t[0].Conjugate(&x.B0)
	t[1].Conjugate(&x.B1).MulByElement(&t[1], &frobCoeffA)

	z.B0 = t[0]
	z.B1 = t[1]

	return z
}

// Frobenius sets z in E24 to x^q, returns z
func (z *E24) Frobenius(x *E24) *E24 {

	var t [12]E2

	frobCoeffA := fp.Element{
		18078622854523849680,
		1984927455914812303,
		2087856039593753391,
		10384413649565796150,
		62117205619808039,
	}
	frobCoeffB := fp.Element{
		16303043965024461850,
		18121025051155953387,
		13066506537353112078,
		1182352075644000644,
		250600645981871933,
	}
	frobCoeffC := fp.Element{
		7336511025221188090,
		2603771785376329468,
		6562537755091890776,
		9030934061021684028,
		49077327029958380,
	}
	frobCoeffD := fp.Element{
		2418585497346853195,
		4985300007263584554,
		14090834255033678869,
		8443077587606433532,
		99465464973580433,
	}
	frobCoeffE := fp.Element{
		16266452697653617742,
		3469624274549632133,
		1532064828142410068,
		17281049885654821422,
		214020232334507350,
	}
	frobCoeffF := fp.Element{
		319632480799633719,
		12918588655636006616,
		7080179244737088245,
		5761903421758065752,
		223741986209306126,
	}
	frobCoeffG := fp.Element{
		8984310047302919300,
		2498109052167961353,
		1307418789688509602,
		11960473000634917703,
		283892625570574947,
	}
	frobCoeffH := fp.Element{
		7656143506020821809,
		15522360441012336084,
		13642716999828979021,
		14792837482779749780,
		272819313239264506,
	}
	frobCoeffI := fp.Element{
		5276991711591121542,
		1764125630309599080,
		4048361144298871290,
		17215093588476212969,
		305552045589664998,
	}
	frobCoeffJ := fp.Element{
		2851480573204638815,
		1335734525939490983,
		5345966389475061568,
		16856815570427136360,
		235013868839987029,
	}
	frobCoeffK := fp.Element{
		13058879110013405254,
		6425951730151460541,
		8330799211860746257,
		12501476947590434451,
		327313527801552489,
	}

	t[0].Conjugate(&x.D0.C0.B0)
	t[1].Conjugate(&x.D0.C0.B1).MulByElement(&t[1], &frobCoeffA)
	t[2].Conjugate(&x.D0.C1.B0).MulByElement(&t[2], &frobCoeffB)
	t[3].Conjugate(&x.D0.C1.B1).MulByElement(&t[3], &frobCoeffF)
	t[4].Conjugate(&x.D1.C0.B0).MulByElement(&t[4], &frobCoeffC)
	t[5].Conjugate(&x.D1.C0.B1).MulByElement(&t[5], &frobCoeffJ)
	t[6].Conjugate(&x.D1.C1.B0).MulByElement(&t[6], &frobCoeffE)
	t[7].Conjugate(&x.D1.C1.B1).MulByElement(&t[7], &frobCoeffI)
	t[8].Conjugate(&x.D2.C0.B0).MulByElement(&t[8], &frobCoeffD)
	t[9].Conjugate(&x.D2.C0.B1).MulByElement(&t[9], &frobCoeffG)
	t[10].Conjugate(&x.D2.C1.B0).MulByElement(&t[10], &frobCoeffH)
	t[11].Conjugate(&x.D2.C1.B1).MulByElement(&t[11], &frobCoeffK)

	z.D0.C0.B0 = t[0]
	z.D0.C0.B1 = t[1]
	z.D0.C1.B0 = t[2]
	z.D0.C1.B1 = t[3]
	z.D1.C0.B0 = t[4]
	z.D1.C0.B1 = t[5]
	z.D1.C1.B0 = t[6]
	z.D1.C1.B1 = t[7]
	z.D2.C0.B0 = t[8]
	z.D2.C0.B1 = t[9]
	z.D2.C1.B0 = t[10]
	z.D2.C1.B1 = t[11]

	return z
}

// FrobeniusSquare sets z in E24 to x^q2, returns z
func (z *E24) FrobeniusSquare(x *E24) *E24 {

	var t [6]E4

	frobCoeffA := fp.Element{
		18078622854523849680,
		1984927455914812303,
		2087856039593753391,
		10384413649565796150,
		62117205619808039,
	}
	frobCoeffD := fp.Element{
		2418585497346853195,
		4985300007263584554,
		14090834255033678869,
		8443077587606433532,
		99465464973580433,
	}
	frobCoeffE := fp.Element{
		16266452697653617742,
		3469624274549632133,
		1532064828142410068,
		17281049885654821422,
		214020232334507350,
	}
	frobCoeffG := fp.Element{
		8984310047302919300,
		2498109052167961353,
		1307418789688509602,
		11960473000634917703,
		283892625570574947,
	}
	frobCoeffI := fp.Element{
		5276991711591121542,
		1764125630309599080,
		4048361144298871290,
		17215093588476212969,
		305552045589664998,
	}

	t[0].Conjugate(&x.D0.C0)
	t[1].Conjugate(&x.D0.C1).MulByElement(&t[1], &frobCoeffA)
	t[2].Conjugate(&x.D1.C0).MulByElement(&t[2], &frobCoeffD)
	t[3].Conjugate(&x.D1.C1).MulByElement(&t[3], &frobCoeffG)
	t[4].Conjugate(&x.D2.C0).MulByElement(&t[4], &frobCoeffE)
	t[5].Conjugate(&x.D2.C1).MulByElement(&t[5], &frobCoeffI)

	z.D0.C0 = t[0]
	z.D0.C1 = t[1]
	z.D1.C0 = t[2]
	z.D1.C1 = t[3]
	z.D2.C0 = t[4]
	z.D2.C1 = t[5]

	return z
}

// FrobeniusQuad sets z in E24 to x^q4, returns z
func (z *E24) FrobeniusQuad(x *E24) *E24 {

	var t [3]E8

	frobCoeffE := fp.Element{
		16266452697653617742,
		3469624274549632133,
		1532064828142410068,
		17281049885654821422,
		214020232334507350,
	}
	frobCoeffG := fp.Element{
		8984310047302919300,
		2498109052167961353,
		1307418789688509602,
		11960473000634917703,
		283892625570574947,
	}

	t[0].Conjugate(&x.D0)
	t[1].Conjugate(&x.D1).MulByElement(&t[1], &frobCoeffE)
	t[2].Conjugate(&x.D2).MulByElement(&t[2], &frobCoeffG)

	z.D0 = t[0]
	z.D1 = t[1]
	z.D2 = t[2]

	return z
}
