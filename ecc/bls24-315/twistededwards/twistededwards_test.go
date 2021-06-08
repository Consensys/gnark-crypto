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

package twistededwards

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
)

func TestMarshal(t *testing.T) {

	var point, unmarshalPoint PointAffine
	point.Set(&edwards.Base)
	for i := 0; i < 20; i++ {
		b := point.Marshal()
		unmarshalPoint.Unmarshal(b)
		if !point.Equal(&unmarshalPoint) {
			t.Fatal("error unmarshal(marshal(point))")
		}
		point.Add(&point, &edwards.Base)
	}
}

func TestAdd(t *testing.T) {

	var p1, p2 PointAffine

	p1.X.SetString("2861651285770559794034091343377448697184139780835112045818187601057344900491")
	p1.Y.SetString("4566653243887596352608254705712316965544071487868883801913476014051372435632")

	p2.X.SetString("3239462834244195151336014620991385997969998964724620516921236505833845606909")
	p2.Y.SetString("5443560678434264661954611726013246003624277823720092161324042476428578455826")

	var expectedX, expectedY fr.Element

	expectedX.SetString("10970949606847471588063681763288201260078321791395639375331302871254875438596")
	expectedY.SetString("520469936960854531602017179843021576180910514245096187577115670676114317535")

	p1.Add(&p1, &p2)

	if !p1.X.Equal(&expectedX) {
		t.Fatal("wrong x coordinate")
	}
	if !p1.Y.Equal(&expectedY) {
		t.Fatal("wrong y coordinate")
	}

}

func TestAddProj(t *testing.T) {

	var p1, p2 PointAffine
	var p1proj, p2proj PointProj

	p1.X.SetString("2861651285770559794034091343377448697184139780835112045818187601057344900491")
	p1.Y.SetString("4566653243887596352608254705712316965544071487868883801913476014051372435632")

	p2.X.SetString("3239462834244195151336014620991385997969998964724620516921236505833845606909")
	p2.Y.SetString("5443560678434264661954611726013246003624277823720092161324042476428578455826")

	p1proj.FromAffine(&p1)
	p2proj.FromAffine(&p2)

	var expectedX, expectedY fr.Element

	expectedX.SetString("10970949606847471588063681763288201260078321791395639375331302871254875438596")
	expectedY.SetString("520469936960854531602017179843021576180910514245096187577115670676114317535")

	p1proj.Add(&p1proj, &p2proj)
	p1.FromProj(&p1proj)

	if !p1.X.Equal(&expectedX) {
		t.Fatal("wrong x coordinate")
	}
	if !p1.Y.Equal(&expectedY) {
		t.Fatal("wrong y coordinate")
	}

}

func TestDouble(t *testing.T) {

	var p PointAffine

	p.X.SetString("7714034250597178209866161557627901990536509610254298591018974201783222420886")
	p.Y.SetString("2993641981652801287937884031498124990237837807065541358157465090538578690179")

	p.Double(&p)

	var expectedX, expectedY fr.Element

	expectedX.SetString("2729808133250287990493483519839683626231554062296052592506429273979000818203")
	expectedY.SetString("1008485385527457545808082794866460438722688514406725394403131994683507639651")

	if !p.X.Equal(&expectedX) {
		t.Fatal("wrong x coordinate")
	}
	if !p.Y.Equal(&expectedY) {
		t.Fatal("wrong y coordinate")
	}
}

func TestDoubleProj(t *testing.T) {

	var p PointAffine
	var pproj PointProj

	p.X.SetString("7714034250597178209866161557627901990536509610254298591018974201783222420886")
	p.Y.SetString("2993641981652801287937884031498124990237837807065541358157465090538578690179")

	pproj.FromAffine(&p).Double(&pproj)

	p.FromProj(&pproj)

	var expectedX, expectedY fr.Element

	expectedX.SetString("2729808133250287990493483519839683626231554062296052592506429273979000818203")
	expectedY.SetString("1008485385527457545808082794866460438722688514406725394403131994683507639651")

	if !p.X.Equal(&expectedX) {
		t.Fatal("wrong x coordinate")
	}
	if !p.Y.Equal(&expectedY) {
		t.Fatal("wrong y coordinate")
	}
}

func TestScalarMul(t *testing.T) {

	// set curve parameters
	ed := GetEdwardsCurve()

	var scalar big.Int
	scalar.SetUint64(23902374)

	var p PointAffine
	p.ScalarMul(&ed.Base, &scalar)

	var expectedX, expectedY fr.Element

	expectedX.SetString("5443630165365522958110505763619304944685728688937477067015993833267732654459")
	expectedY.SetString("5027805281744385394708323878140623534818171762863978182421480579085966004942")

	if !expectedX.Equal(&p.X) {
		t.Fatal("wrong x coordinate")
	}
	if !expectedY.Equal(&p.Y) {
		t.Fatal("wrong y coordinate")
	}

	// test consistancy with negation
	var expected, base PointAffine
	expected.Set(&ed.Base).Neg(&expected)
	scalar.Set(&ed.Order).Lsh(&scalar, 3) // multiply by cofactor=8
	scalar.Sub(&scalar, big.NewInt(1))
	base.Set(&ed.Base)
	base.ScalarMul(&base, &scalar)
	if !base.Equal(&expected) {
		t.Fatal("Mul by order-1 not consistant with neg")
	}
}
